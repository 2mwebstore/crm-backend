package services

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"crm-backend/models"
	"crm-backend/repositories"
)

// PayrollExportService builds the "Detail Report Payroll" workbook.
// Attend/Absent/Leave classification is delegated entirely to
// AttendanceService.Summary() (the same engine behind the in-app
// Attendance Detail Report) rather than being re-implemented here. Only
// OT Hours (from Overtime requests) and the Annual Leave monthly-limit
// lookup are computed independently here, since AttendanceService.Summary()
// doesn't cover either of those.
type PayrollExportService interface {
	Generate(callerID uint, dateFrom, dateTo string, userID, branchID uint) (*excelize.File, error)
}

type payrollExportService struct {
	userRepo      repositories.UserRepository
	attendanceSvc AttendanceService
	overtimeRepo  repositories.OvertimeRequestRepository
	leaveTypeRepo repositories.LeaveTypeRepository
	leaveRepo     repositories.LeaveRequestRepository
}

func NewPayrollExportService(
	userRepo repositories.UserRepository,
	attendanceSvc AttendanceService,
	overtimeRepo repositories.OvertimeRequestRepository,
	leaveTypeRepo repositories.LeaveTypeRepository,
	leaveRepo repositories.LeaveRequestRepository,
) PayrollExportService {
	return &payrollExportService{userRepo, attendanceSvc, overtimeRepo, leaveTypeRepo, leaveRepo}
}

func (s *payrollExportService) Generate(callerID uint, dateFrom, dateTo string, userID, branchID uint) (*excelize.File, error) {
	if dateFrom == "" || dateTo == "" {
		return nil, fmt.Errorf("date_from and date_to are required")
	}
	dates, err := dateRangeList(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	// The SAME engine that drives the in-app Attendance Detail Report —
	// Attend/Absent/Leave classification all comes from here, not
	// re-derived independently.
	summaries, err := s.attendanceSvc.Summary(callerID, dateFrom, dateTo, userID, branchID)
	if err != nil {
		return nil, err
	}

	scopedUsers, err := s.userRepo.GetUsersInScope(callerID)
	if err != nil {
		return nil, err
	}
	userByID := map[uint]models.User{}
	for _, u := range scopedUsers {
		userByID[u.ID] = u
	}

	// Overtime: fetch every status (no Status filter), then keep only
	// approved and pending.
	otFilter := repositories.OvertimeRequestFilter{DateFrom: dateFrom, DateTo: dateTo, BranchID: branchID}
	if userID != 0 {
		otFilter.UserID = userID
	}
	otRowsRaw, _, err := s.overtimeRepo.List(otFilter, 1, 10000)
	if err != nil {
		return nil, err
	}
	otByUser := map[uint]float64{}
	for _, o := range otRowsRaw {
		if o.Status != models.OvertimeRequestApproved && o.Status != models.OvertimeRequestPending {
			continue
		}
		if o.Duration != nil {
			otByUser[o.UserID] += *o.Duration
		}
	}

	leaveTypes, err := s.leaveTypeRepo.List(callerID, true, nil)
	if err != nil {
		return nil, err
	}

	// Annual Leave (AL) usage per user — a leave request counts toward
	// this only if its LeaveType's name is "Annual Leave"
	// (case-insensitive) or its code is "AL". Separate from
	// summary.Leave (which totals EVERY leave type combined).
	annualLeaveFilter := repositories.LeaveRequestFilter{DateFrom: dateFrom, DateTo: dateTo, BranchID: branchID}
	if userID != 0 {
		annualLeaveFilter.UserID = userID
	}
	allLeaveRows, _, err := s.leaveRepo.List(annualLeaveFilter, 1, 10000)
	if err != nil {
		return nil, err
	}
	annualLeaveByUser := map[uint]float64{}
	for _, l := range allLeaveRows {
		if l.Status != models.LeaveRequestApproved && l.Status != models.LeaveRequestPending {
			continue
		}
		if l.LeaveType != nil && isAnnualLeaveType(*l.LeaveType) {
			annualLeaveByUser[l.UserID] += l.Duration
		}
	}

	f := excelize.NewFile()
	f.SetSheetName(f.GetSheetName(0), "Detail Report Payroll")
	sh := "Detail Report Payroll"

	// ── Column layout ──────────────────────────────────────────────────
	// A: Name, B: Branch (Staff ID / Position & Department removed)
	// C..(C+len(dates)-1): one column per calendar day
	nameCol := 1   // A
	branchCol := 2 // B
	dayColStart := 3
	numDayCols := len(dates)
	col := func(n int) string { c, _ := excelize.ColumnNumberToName(n); return c }

	dayWorkCol := dayColStart + numDayCols
	attendCol := dayWorkCol + 1
	workTimeCol := attendCol + 1
	lateQCol := workTimeCol + 1
	lateCountCol := lateQCol + 1
	earlyQCol := lateCountCol + 1
	earlyCountCol := earlyQCol + 1
	otQCol := earlyCountCol + 1
	otCountCol := otQCol + 1
	absentCol := otCountCol + 1
	leaveTotalCol := absentCol + 1
	annualLeaveCol := leaveTotalCol + 1 // "AL" usage
	otDayCol := annualLeaveCol + 1      // "OT Day" — in the Other group
	otDayx2Col := otDayCol + 1          // "OT Day x2" — static 0
	otHoursCol := otDayx2Col + 1        // Overtime group starts here
	otx2Col := otHoursCol + 1           // "OT Hours x2" — static 0
	lastCol := otx2Col

	blackBorder := []excelize.Border{
		{Type: "top", Color: "#000000", Style: 1}, {Type: "bottom", Color: "#000000", Style: 1},
		{Type: "left", Color: "#000000", Style: 1}, {Type: "right", Color: "#000000", Style: 1},
	}

	// ── Title row ───────────────────────────────────────────────────────
	monthLabel := ""
	if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
		monthLabel = t.Format("January 2006")
	}
	f.MergeCell(sh, "A1", col(lastCol)+"1")
	f.SetCellValue(sh, "A1", "Detail Report Payroll "+monthLabel)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Color: "#FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#5A80B8"}, Pattern: 1},
	})
	f.SetCellStyle(sh, "A1", col(lastCol)+"1", titleStyle)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F3F4F6"}, Pattern: 1},
		Border:    blackBorder,
	})
	// "Other" group header (was "Leave") — colored text on the same gray
	// fill as every other group header.
	// "Other" group header uses the same default headerStyle as every
	// other group header (no colored text) — applied via the blanket
	// f.SetCellStyle(sh, col(nameCol)+"2", col(lastCol)+"3", headerStyle)
	// call below, so no separate style object is needed for it.
	// Sub-header style for "Leave"/"OT Day"/"OT Day x2" under the Other
	// group — same #F5C344 colored text as the group header itself.
	subHeaderYellowStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "#F5C344"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F3F4F6"}, Pattern: 1},
		Border:    blackBorder,
	})
	// "AL" sub-header — red text, standing apart from the rest of the
	// Other group's yellow sub-headers.
	subHeaderRedStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "#DC2626"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F3F4F6"}, Pattern: 1},
		Border:    blackBorder,
	})

	// ── Row 2: group headers ───────────────────────────────────────────
	f.MergeCell(sh, col(nameCol)+"2", col(branchCol)+"2")
	f.SetCellValue(sh, col(nameCol)+"2", "Staff")
	for i, d := range dates {
		t, _ := time.Parse("2006-01-02", d)
		c := col(dayColStart + i)
		f.SetCellValue(sh, c+"2", t.Format("02"))
	}
	f.SetCellValue(sh, col(dayWorkCol)+"2", "Day Work")
	f.SetCellValue(sh, col(attendCol)+"2", "Attendance")
	f.SetCellValue(sh, col(workTimeCol)+"2", "Work Time")
	f.MergeCell(sh, col(lateQCol)+"2", col(lateCountCol)+"2")
	f.SetCellValue(sh, col(lateQCol)+"2", "Checkin Late")
	f.MergeCell(sh, col(earlyQCol)+"2", col(earlyCountCol)+"2")
	f.SetCellValue(sh, col(earlyQCol)+"2", "Checkout Early")
	f.MergeCell(sh, col(otQCol)+"2", col(otCountCol)+"2")
	f.SetCellValue(sh, col(otQCol)+"2", "Checkout Overtime")
	f.SetCellValue(sh, col(absentCol)+"2", "Absent")
	f.MergeCell(sh, col(leaveTotalCol)+"2", col(otDayx2Col)+"2")
	f.SetCellValue(sh, col(leaveTotalCol)+"2", "Other")
	f.MergeCell(sh, col(otHoursCol)+"2", col(otx2Col)+"2")
	f.SetCellValue(sh, col(otHoursCol)+"2", "Overtime")

	// ── Row 3: sub headers ──────────────────────────────────────────────
	f.SetCellValue(sh, col(nameCol)+"3", "Name")
	f.SetCellValue(sh, col(branchCol)+"3", "Branch")
	f.MergeCell(sh, col(dayColStart)+"3", col(dayColStart+numDayCols-1)+"3")
	f.SetCellValue(sh, col(dayColStart)+"3", "Check in - Check out")
	f.SetCellValue(sh, col(dayWorkCol)+"3", "Count")
	f.SetCellValue(sh, col(attendCol)+"3", "Count")
	f.SetCellValue(sh, col(workTimeCol)+"3", "Q-mn")
	f.SetCellValue(sh, col(lateQCol)+"3", "Q-mn")
	f.SetCellValue(sh, col(lateCountCol)+"3", "Count")
	f.SetCellValue(sh, col(earlyQCol)+"3", "Q-mn")
	f.SetCellValue(sh, col(earlyCountCol)+"3", "Count")
	f.SetCellValue(sh, col(otQCol)+"3", "Q-mn")
	f.SetCellValue(sh, col(otCountCol)+"3", "Count")
	f.SetCellValue(sh, col(absentCol)+"3", "Absent")
	f.SetCellValue(sh, col(leaveTotalCol)+"3", "Leave")
	f.SetCellValue(sh, col(annualLeaveCol)+"3", "AL")
	f.SetCellValue(sh, col(otDayCol)+"3", "OT Day")
	f.SetCellValue(sh, col(otDayx2Col)+"3", "OT Day x2")
	f.SetCellValue(sh, col(otHoursCol)+"3", "OT Hours")
	f.SetCellValue(sh, col(otx2Col)+"3", "OT Hours x2")

	f.SetCellStyle(sh, col(nameCol)+"2", col(lastCol)+"3", headerStyle)
	f.SetCellStyle(sh, col(leaveTotalCol)+"3", col(leaveTotalCol)+"3", subHeaderYellowStyle)
	f.SetCellStyle(sh, col(annualLeaveCol)+"3", col(annualLeaveCol)+"3", subHeaderRedStyle)
	f.SetCellStyle(sh, col(otDayCol)+"3", col(otDayCol)+"3", subHeaderYellowStyle)
	f.SetCellStyle(sh, col(otDayx2Col)+"3", col(otDayx2Col)+"3", subHeaderYellowStyle)

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    blackBorder,
	})
	halfDayStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Bold: true, Color: "#F5C344"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    blackBorder,
	})
	fullDayStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F5C344"}, Pattern: 1},
		Border:    blackBorder,
	})
	absentStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#DC2626"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    blackBorder,
	})
	// Red background fill for an ABSENT day column cell (as opposed to
	// absentStyle above, which is red TEXT used on the Absent summary
	// count column).
	absentDayStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FECACA"}, Pattern: 1}, // red-200
		Border:    blackBorder,
	})

	sort.Slice(summaries, func(i, j int) bool { return summaries[i].UserName < summaries[j].UserName })

	// ── Data rows ───────────────────────────────────────────────────────
	row := 4
	for _, summary := range summaries {
		u := userByID[summary.UserID]

		f.SetCellStyle(sh, col(nameCol)+strconv.Itoa(row), col(lastCol)+strconv.Itoa(row), cellStyle)
		f.SetRowHeight(sh, row, 30)

		var workMinutes, lateMinutes, earlyMinutes, otMinutes int
		var lateCount, earlyCount, otCountVal int

		for i, d := range summary.Days {
			dayCol := col(dayColStart + i)
			status := d.Status

			switch {
			case strings.Contains(status, "ATTEND"):
				cellText := ""
				if d.CheckInAt != nil {
					cellText = *d.CheckInAt
					if d.CheckOutAt != nil {
						cellText = *d.CheckInAt + " - " + *d.CheckOutAt
						if ci, ok := parseHHMMToMinutes(*d.CheckInAt); ok {
							if co, ok := parseHHMMToMinutes(*d.CheckOutAt); ok {
								workMinutes += co - ci
							}
						}
					}
				}
				f.SetCellValue(sh, dayCol+strconv.Itoa(row), cellText)

				if strings.Contains(status, "Half Day") {
					f.SetCellStyle(sh, dayCol+strconv.Itoa(row), dayCol+strconv.Itoa(row), halfDayStyle)
				} else if strings.Contains(status, "LEAVE") {
					f.SetCellStyle(sh, dayCol+strconv.Itoa(row), dayCol+strconv.Itoa(row), fullDayStyle)
				}

				if d.CheckInStatus == "late" && d.CheckInAt != nil {
					lateCount++
					if mins, ok := minutesPastStr(u.ShiftCheckInTime, *d.CheckInAt); ok {
						lateMinutes += mins
					}
				}
				if d.CheckOutAt != nil {
					if d.CheckOutStatus == "early" {
						earlyCount++
						if mins, ok := minutesBeforeStr(u.ShiftCheckOutTime, *d.CheckOutAt); ok {
							earlyMinutes += mins
						}
					} else if d.CheckOutStatus == "good" {
						if mins, ok := minutesPastStr(u.ShiftCheckOutTime, *d.CheckOutAt); ok && mins > 0 {
							otCountVal++
							otMinutes += mins
						}
					}
				}

			case strings.Contains(status, "Half Day"):
				f.SetCellStyle(sh, dayCol+strconv.Itoa(row), dayCol+strconv.Itoa(row), halfDayStyle)

			case strings.Contains(status, "LEAVE"):
				f.SetCellStyle(sh, dayCol+strconv.Itoa(row), dayCol+strconv.Itoa(row), fullDayStyle)

			default:
				// ABSENT (DAY_OFF is unreachable currently) — red
				// background so an absence stands out on the calendar
				// grid, not just in the summary count column.
				f.SetCellStyle(sh, dayCol+strconv.Itoa(row), dayCol+strconv.Itoa(row), absentDayStyle)
			}
		}

		f.SetCellValue(sh, col(nameCol)+strconv.Itoa(row), summary.UserName)
		f.SetCellValue(sh, col(branchCol)+strconv.Itoa(row), summary.BranchNames)
		f.SetCellValue(sh, col(dayWorkCol)+strconv.Itoa(row), len(dates))
		f.SetCellValue(sh, col(attendCol)+strconv.Itoa(row), summary.Attend)
		f.SetCellValue(sh, col(workTimeCol)+strconv.Itoa(row), fmtHoursMinutes(workMinutes))
		f.SetCellValue(sh, col(lateQCol)+strconv.Itoa(row), fmtHoursMinutes(lateMinutes))
		f.SetCellValue(sh, col(lateCountCol)+strconv.Itoa(row), lateCount)
		f.SetCellValue(sh, col(earlyQCol)+strconv.Itoa(row), fmtHoursMinutes(earlyMinutes))
		f.SetCellValue(sh, col(earlyCountCol)+strconv.Itoa(row), earlyCount)
		f.SetCellValue(sh, col(otQCol)+strconv.Itoa(row), fmtHoursMinutes(otMinutes))
		f.SetCellValue(sh, col(otCountCol)+strconv.Itoa(row), otCountVal)
		f.SetCellValue(sh, col(absentCol)+strconv.Itoa(row), summary.Absent)
		f.SetCellStyle(sh, col(absentCol)+strconv.Itoa(row), col(absentCol)+strconv.Itoa(row), absentStyle)
		f.SetCellValue(sh, col(leaveTotalCol)+strconv.Itoa(row), summary.Leave)
		f.SetCellValue(sh, col(annualLeaveCol)+strconv.Itoa(row), annualLeaveByUser[summary.UserID])
		remainingLeaveDays := summary.Leave - float64(monthlyLeaveLimit(u, leaveTypes))
		f.SetCellValue(sh, col(otDayCol)+strconv.Itoa(row), remainingLeaveDays)
		f.SetCellValue(sh, col(otDayx2Col)+strconv.Itoa(row), 0) // static per request
		f.SetCellValue(sh, col(otHoursCol)+strconv.Itoa(row), otByUser[summary.UserID])
		f.SetCellValue(sh, col(otx2Col)+strconv.Itoa(row), 0) // static per request

		row++
	}

	f.SetColWidth(sh, col(nameCol), col(nameCol), 18)
	f.SetColWidth(sh, col(branchCol), col(branchCol), 16)
	f.SetColWidth(sh, col(dayColStart), col(dayColStart+numDayCols-1), 13)
	f.SetColWidth(sh, col(dayWorkCol), col(lastCol), 11)
	f.SetRowHeight(sh, 1, 30)
	f.SetRowHeight(sh, 2, 30)
	f.SetRowHeight(sh, 3, 30)

	f.SetActiveSheet(0)

	return f, nil
}

// minutesPastStr / minutesBeforeStr are string-time variants of
// minutesPast/minutesBefore — for use with DaySummary's already-formatted
// "HH:MM" CheckInAt/CheckOutAt strings, rather than a *time.Time.
func minutesPastStr(shiftTime *string, actualHHMM string) (int, bool) {
	shiftMin, ok := parseHHMMToMinutes(derefOrEmpty(shiftTime))
	if !ok {
		return 0, false
	}
	actualMin, ok := parseHHMMToMinutes(actualHHMM)
	if !ok {
		return 0, false
	}
	diff := actualMin - shiftMin
	if diff < 0 {
		return 0, false
	}
	return diff, true
}

func minutesBeforeStr(shiftTime *string, actualHHMM string) (int, bool) {
	shiftMin, ok := parseHHMMToMinutes(derefOrEmpty(shiftTime))
	if !ok {
		return 0, false
	}
	actualMin, ok := parseHHMMToMinutes(actualHHMM)
	if !ok {
		return 0, false
	}
	diff := shiftMin - actualMin
	if diff < 0 {
		return 0, false
	}
	return diff, true
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// isAnnualLeaveType matches a leave type by name ("Annual Leave",
// case-insensitive) or code ("AL").
func isAnnualLeaveType(lt models.LeaveType) bool {
	return strings.EqualFold(strings.TrimSpace(lt.Name), "Annual Leave") || strings.EqualFold(strings.TrimSpace(lt.Code), "AL")
}

// monthlyLeaveLimit returns the MonthlyLimit of the Annual Leave type
// applicable to this user (global or matching one of their branches) —
// NOT a sum across every leave type. Returns 0 if no Annual Leave type
// applies or it has no MonthlyLimit set.
func monthlyLeaveLimit(u models.User, leaveTypes []models.LeaveType) int {
	userBranchIDs := map[uint]bool{}
	for _, b := range u.Branches {
		userBranchIDs[b.ID] = true
	}
	for _, lt := range leaveTypes {
		if !lt.IsActive || !isAnnualLeaveType(lt) {
			continue
		}
		applies := lt.BranchID == nil || userBranchIDs[*lt.BranchID]
		if applies && lt.MonthlyLimit != nil {
			return *lt.MonthlyLimit
		}
	}
	return 0
}

// fmtHoursMinutes formats a minute count as "Xh Ym", or "—" for zero.
func fmtHoursMinutes(totalMinutes int) string {
	if totalMinutes <= 0 {
		return "—"
	}
	h := totalMinutes / 60
	m := totalMinutes % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
