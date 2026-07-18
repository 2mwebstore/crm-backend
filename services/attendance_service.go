package services

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type AttendanceService interface {
	CheckIn(userID, branchID uint, lat, lng float64, reason string) (*models.Attendance, error)
	CheckOut(userID, branchID uint, lat, lng float64, reason string) (*models.Attendance, error)
	Today(userID, branchID uint) (*models.Attendance, error)
	List(filter repositories.AttendanceFilter, page, pageSize int) ([]models.Attendance, int64, error)
	// AdminUpdate lets an admin correct an existing record's check-in
	// and/or check-out timestamp (e.g. a forgotten punch) — both
	// optional ("" = leave that side untouched), each a
	// "2006-01-02T15:04" datetime-local string interpreted in
	// Asia/Phnom_Penh (same as every other timestamp in this app). The
	// timeliness status for whichever side changes is recomputed against
	// the user's shift times for the record's own Date, exactly the same
	// way a normal check-in/check-out would compute it — never left
	// stale.
	AdminUpdate(id uint, checkInAtStr, checkOutAtStr string) (*models.Attendance, error)
	// Summary computes a day-by-day ATTEND/ABSENT/LEAVE breakdown
	// per user over a date range — computed server-side against real
	// time.Time/date values (not string matching) specifically to avoid
	// the DB-driver date round-trip quirk that broke the earlier
	// frontend-only version of this report.
	Summary(callerID uint, dateFrom, dateTo string, userID, branchID uint) ([]UserAttendanceSummary, error)
}

// DaySummary is one calendar day's attendance status for one user.
type DaySummary struct {
	Date                string  `json:"date"`
	Weekday             string  `json:"weekday"`
	Status              string  `json:"status"` // ATTEND, ABSENT, LEAVE, "Half Day - Morning", "Half Day - Afternoon"
	CheckInAt           *string `json:"check_in_at,omitempty"`
	CheckOutAt          *string `json:"check_out_at,omitempty"`
	CheckInStatus       string  `json:"check_in_status,omitempty"`
	CheckOutStatus      string  `json:"check_out_status,omitempty"`
	CheckInViaActivity  bool    `json:"check_in_via_activity"`
	CheckOutViaActivity bool    `json:"check_out_via_activity"`
	// CheckInReason/CheckOutReason are each only ever populated when that
	// side was auto-driven by an approved Activity request — a normal
	// manual check-in/out never carries a reason. Kept separate rather
	// than merged into one field, since a day can have a different
	// reason on each side (e.g. Activity-driven check-in, ordinary
	// manual check-out).
	CheckInReason  string `json:"check_in_reason,omitempty"`
	CheckOutReason string `json:"check_out_reason,omitempty"`
}

// UserAttendanceSummary is one user's totals + daily breakdown for the
// requested date range. Attend/Absent/Leave are floats since a single day
// can contribute a half (0.5) to more than one bucket — e.g. a half-day
// leave with no attendance the other half counts 0.5 leave + 0.5 absent,
// and a check-in with no check-out counts 1 attend + 0.5 absent (an
// incomplete day is only half "attended" in practice).
type UserAttendanceSummary struct {
	UserID      uint         `json:"user_id"`
	UserName    string       `json:"user_name"`
	BranchNames string       `json:"branch_names"`
	Attend      float64      `json:"attend"`
	Absent      float64      `json:"absent"`
	Leave       float64      `json:"leave"`
	Days        []DaySummary `json:"days"`
}

type attendanceService struct {
	repo                 repositories.AttendanceRepository
	branchRepo           repositories.BranchRepository
	activityRepo         repositories.ActivityRequestRepository
	userRepo             repositories.UserRepository
	scheduleOverrideRepo repositories.UserScheduleOverrideRepository
	leaveRepo            repositories.LeaveRequestRepository
}

func NewAttendanceService(
	repo repositories.AttendanceRepository,
	branchRepo repositories.BranchRepository,
	activityRepo repositories.ActivityRequestRepository,
	userRepo repositories.UserRepository,
	scheduleOverrideRepo repositories.UserScheduleOverrideRepository,
	leaveRepo repositories.LeaveRequestRepository,
) AttendanceService {
	return &attendanceService{repo, branchRepo, activityRepo, userRepo, scheduleOverrideRepo, leaveRepo}
}

const defaultCheckInRadiusMeters = 200

// Timeliness labels — reused for both check-in (against
// User.ShiftCheckInTime) and check-out (against User.ShiftCheckOutTime):
//   - "early" — before the shift time
//   - "good"  — between the shift time and shift time + 15 minutes
//   - "late"  — more than 15 minutes after the shift time
const (
	TimelinessEarly = "early"
	TimelinessGood  = "good"
	TimelinessLate  = "late"
)

const lateGraceMinutes = 15

// todayDateString returns today's date (Asia/Phnom_Penh) as "2006-01-02" —
// matches nowInCambodia()'s timezone (defined in daily_start_balance_service.go,
// shared across this package) so "today" means the same calendar day
// everywhere in the app, not whatever timezone the server happens to run in.
func todayDateString() string {
	return nowInCambodia().Format("2006-01-02")
}

// parseHHMMToMinutes converts "HH:MM" to minutes-since-midnight, or ok=false
// if the string isn't in that exact shape.
func parseHHMMToMinutes(s string) (minutes int, ok bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, false
	}
	return h*60 + m, true
}

// computeTimelinessStatus compares check-in time against ShiftCheckInTime —
// "early" (before), "good" (shift time to shift time + 15min), or "late"
// (beyond that). Returns "" if shiftTime is nil/empty — nothing to compare
// against, and that's not the same as being late.
func computeTimelinessStatus(actualTime string, shiftTime *string) string {
	if shiftTime == nil || *shiftTime == "" {
		return ""
	}
	shiftMinutes, ok1 := parseHHMMToMinutes(*shiftTime)
	actualMinutes, ok2 := parseHHMMToMinutes(actualTime)
	if !ok1 || !ok2 {
		return ""
	}
	switch {
	case actualMinutes < shiftMinutes:
		return TimelinessEarly
	case actualMinutes <= shiftMinutes+lateGraceMinutes:
		return TimelinessGood
	default:
		return TimelinessLate
	}
}

// computeCheckOutTimeliness compares check-out time against
// ShiftCheckOutTime — deliberately only "early" (before shift-out time) or
// "good" (at or after it). Staying past shift-out time isn't a problem the
// way arriving late is, so there's no "late" tier here at all: checking
// out anytime at or after shift-out time is simply "good", no matter how
// much later. Returns "" if shiftTime is nil/empty.
func computeCheckOutTimeliness(actualTime string, shiftTime *string) string {
	if shiftTime == nil || *shiftTime == "" {
		return ""
	}
	shiftMinutes, ok1 := parseHHMMToMinutes(*shiftTime)
	actualMinutes, ok2 := parseHHMMToMinutes(actualTime)
	if !ok1 || !ok2 {
		return ""
	}
	if actualMinutes < shiftMinutes {
		return TimelinessEarly
	}
	return TimelinessGood
}

// resolveShiftTimes returns the effective ShiftCheckInTime/ShiftCheckOutTime
// for userID on date — an active UserScheduleOverride covering that date
// takes priority; either side left nil on the override falls back to the
// user's own default for that side. If there's no override at all, or the
// user record can't be loaded, falls back to the user's own defaults (or
// nil/nil if that also fails — computeTimelinessStatus/computeCheckOutTimeliness
// both treat nil as "nothing to compare against", not an error).
func (s *attendanceService) resolveShiftTimes(userID uint, date string) (checkIn *string, checkOut *string) {
	user, uerr := s.userRepo.FindByID(userID)
	if uerr == nil {
		checkIn = user.ShiftCheckInTime
		checkOut = user.ShiftCheckOutTime
	}
	override, oerr := s.scheduleOverrideRepo.FindActiveForDate(userID, date)
	if oerr == nil {
		if override.ShiftCheckInTime != nil {
			checkIn = override.ShiftCheckInTime
		}
		if override.ShiftCheckOutTime != nil {
			checkOut = override.ShiftCheckOutTime
		}
	}
	return checkIn, checkOut
}

// newShiftWindowMinutes is how long before a shift's own start time
// counts as "approaching a new shift" — see hasReachedNewShiftCutoff.
const newShiftWindowMinutes = 60

// hasReachedNewShiftCutoff returns whether now is within the hour BEFORE
// shiftCheckInTime ("HH:MM"), or any time after it. Used so a Cross Day
// user's forgotten checkout from the previous shift stops being shown as
// "today's" status once it's time to start a new one — otherwise the
// Check In button would stay stuck disabled forever behind an old,
// never-closed shift.
// hasReachedNewShiftCutoff returns whether now is AT OR PAST the point 1
// hour before shiftCheckInTime ("HH:MM") — deliberately open-ended on the
// upper end (not a narrow window), since once it's time to start a new
// shift, that stays true for the rest of the day, not just for one hour.
// A version of this that flipped back to false once shiftCheckInTime
// itself had passed would keep re-showing yesterday's forgotten checkout
// hours later, which is exactly the bug this fixes.
func hasReachedNewShiftCutoff(now time.Time, shiftCheckInTime *string) bool {
	if shiftCheckInTime == nil || *shiftCheckInTime == "" {
		return false
	}
	shiftMinutes, ok := parseHHMMToMinutes(*shiftCheckInTime)
	if !ok {
		return false
	}
	nowMinutes := now.Hour()*60 + now.Minute()
	cutoff := shiftMinutes - newShiftWindowMinutes
	if cutoff < 0 {
		// Shift starts early enough (e.g. 00:30) that the cutoff wraps
		// past midnight into the previous clock day — "at or past cutoff"
		// then means either late in the previous day, or anytime from
		// midnight up to (but not past) the shift start itself; after
		// shiftMinutes it's simply a new day's "yesterday" anyway.
		cutoff += 24 * 60
		return nowMinutes >= cutoff || nowMinutes < shiftMinutes
	}
	return nowMinutes >= cutoff
}

// CheckIn validates distance from the branch's configured location UNLESS
// the user has an APPROVED Activity request covering today — in
// that case the distance check is skipped entirely, since the whole point
// of an activity request is that they're not expected to be at the branch.
func (s *attendanceService) CheckIn(userID, branchID uint, lat, lng float64, reason string) (*models.Attendance, error) {
	today := todayDateString()

	existing, err := s.repo.FindByUserAndDate(userID, today)
	if err == nil && existing.CheckInAt != nil {
		return nil, errors.New("already checked in today")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	isActivity, err := s.activityRepo.HasForDate(userID, today)
	if err != nil {
		return nil, err
	}

	// Distance is always computed when the branch has a location
	// configured — even for an activity-exempted check-in — so the record
	// still shows how far away they actually were. isActivity only skips
	// the ENFORCEMENT (rejecting the check-in), never the calculation
	// itself, since that's useful data to have either way.
	var distance *float64
	branch, branchErr := s.branchRepo.FindByID(branchID)
	if branchErr == nil && branch.Latitude != nil && branch.Longitude != nil {
		d := utils.HaversineDistanceMeters(*branch.Latitude, *branch.Longitude, lat, lng)
		distance = &d
	}

	if !isActivity {
		if branchErr != nil {
			return nil, errors.New("branch not found")
		}
		if branch.Latitude == nil || branch.Longitude == nil {
			return nil, errors.New("this branch has no location configured yet — contact an admin before checking in")
		}
		radius := branch.CheckInRadiusMeters
		if radius <= 0 {
			radius = defaultCheckInRadiusMeters
		}
		if *distance > float64(radius) {
			return nil, fmt.Errorf("you are %.0fm from the branch — must be within %dm to check in (or submit an Activity request for today)", *distance, radius)
		}
	}

	now := nowInCambodia()

	var status string
	shiftCheckIn, _ := s.resolveShiftTimes(userID, today)
	status = computeTimelinessStatus(now.Format("15:04"), shiftCheckIn)

	// A late check-in needs a reason on record — computed BEFORE the row
	// is created/updated, so a missing reason blocks the check-in
	// entirely rather than silently saving without one.
	if status == TimelinessLate && strings.TrimSpace(reason) == "" {
		return nil, errors.New("you're checking in late — please provide a reason")
	}

	if existing != nil && existing.ID != 0 {
		existing.CheckInAt = &now
		existing.CheckInLat = &lat
		existing.CheckInLng = &lng
		existing.CheckInDistance = distance
		existing.CheckInViaActivity = isActivity
		existing.CheckInStatus = status
		if reason != "" {
			existing.CheckInReason = reason
		}
		if err := s.repo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	att := &models.Attendance{
		UserID:             userID,
		BranchID:           branchID,
		Date:               today,
		CheckInAt:          &now,
		CheckInLat:         &lat,
		CheckInLng:         &lng,
		CheckInDistance:    distance,
		CheckInViaActivity: isActivity,
		CheckInStatus:      status,
		CheckInReason:      reason,
	}
	if err := s.repo.Create(att); err != nil {
		return nil, err
	}
	return att, nil
}

// CheckOut normally requires that a check-in already exists for today — no
// distance gate blocks it either way, matching how activity-exempted
// check-ins also need to be able to check out normally at end of day.
//
// Exception: if the user has an APPROVED Activity request for
// today and never checked in at all (the whole point of an activity day is
// they're not expected to be at the branch, so there may be nothing to
// check in AGAINST), check-out is still allowed — it creates today's
// attendance row on the fly with only the check-out side filled in
// (CheckInAt stays nil), rather than requiring a check-in first.
//
// Distance is still recorded for reference in both cases, just never
// enforced as a requirement.
func (s *attendanceService) CheckOut(userID, branchID uint, lat, lng float64, reason string) (*models.Attendance, error) {
	today := todayDateString()

	att, err := s.repo.FindByUserAndDate(userID, today)
	hasRow := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Cross Day (Night Shift) users can check in one calendar day and
	// check out the next — if today has no open check-in, look for a
	// still-open one from YESTERDAY before falling back to the activity
	// exemption / "haven't checked in" error below. This is what lets a
	// normal midnight-crossing checkout land on the correct original
	// attendance row instead of failing or creating a bogus new one.
	if !hasRow || att.CheckInAt == nil {
		if user, uerr := s.userRepo.FindByID(userID); uerr == nil && user.ShiftType == models.ShiftTypeCrossDay {
			yesterday := nowInCambodia().AddDate(0, 0, -1).Format("2006-01-02")
			if yAtt, yerr := s.repo.FindByUserAndDate(userID, yesterday); yerr == nil && yAtt.CheckInAt != nil && yAtt.CheckOutAt == nil {
				att = yAtt
				hasRow = true
			}
		}
	}

	if hasRow && att.CheckOutAt != nil {
		return nil, errors.New("already checked out today")
	}

	isActivity, oerr := s.activityRepo.HasForDate(userID, today)
	if oerr != nil {
		return nil, oerr
	}

	if !hasRow || att.CheckInAt == nil {
		if !isActivity {
			return nil, errors.New("you haven't checked in today yet")
		}
	}

	var distance *float64
	if branch, err := s.branchRepo.FindByID(branchID); err == nil && branch.Latitude != nil && branch.Longitude != nil {
		d := utils.HaversineDistanceMeters(*branch.Latitude, *branch.Longitude, lat, lng)
		distance = &d
	}

	now := nowInCambodia()

	var status string
	effectiveDate := today
	if hasRow {
		effectiveDate = att.Date
	}
	_, shiftCheckOut := s.resolveShiftTimes(userID, effectiveDate)
	status = computeCheckOutTimeliness(now.Format("15:04"), shiftCheckOut)

	// An early check-out needs a reason on record — computed BEFORE the
	// row is created/updated, so a missing reason blocks the check-out
	// entirely rather than silently saving without one.
	if status == TimelinessEarly && strings.TrimSpace(reason) == "" {
		return nil, errors.New("you're checking out early — please provide a reason")
	}

	if !hasRow {
		att = &models.Attendance{
			UserID:   userID,
			BranchID: branchID,
			Date:     today,
		}
	}
	att.CheckOutAt = &now
	att.CheckOutLat = &lat
	att.CheckOutLng = &lng
	att.CheckOutDistance = distance
	att.CheckOutStatus = status
	att.CheckOutViaActivity = isActivity
	if reason != "" {
		att.CheckOutReason = reason
	}

	if !hasRow {
		if err := s.repo.Create(att); err != nil {
			return nil, err
		}
		return att, nil
	}
	if err := s.repo.Update(att); err != nil {
		return nil, err
	}
	return att, nil
}

func (s *attendanceService) Today(userID, branchID uint) (*models.Attendance, error) {
	att, err := s.repo.FindByUserAndDate(userID, todayDateString())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Cross Day (Night Shift) users may still be mid-shift from
			// YESTERDAY (checked in before midnight, not checked out
			// yet) — show that instead of "not checked in", same
			// yesterday-lookback CheckOut already does.
			if user, uerr := s.userRepo.FindByID(userID); uerr == nil && user.ShiftType == models.ShiftTypeCrossDay {
				yesterday := nowInCambodia().AddDate(0, 0, -1).Format("2006-01-02")
				if yAtt, yerr := s.repo.FindByUserAndDate(userID, yesterday); yerr == nil && yAtt.CheckInAt != nil && yAtt.CheckOutAt == nil {
					// Once we're within an hour of the NEW shift's start
					// time, stop showing yesterday's forgotten-checkout
					// row — let the person start a fresh day instead of
					// staying stuck behind an old unclosed shift. Outside
					// that window (genuinely still mid-shift), keep
					// showing it as before.
					if !hasReachedNewShiftCutoff(nowInCambodia(), user.ShiftCheckInTime) {
						return yAtt, nil
					}
				}
			}
			return nil, nil // no row yet today — not an error, just "hasn't checked in"
		}
		return nil, err
	}
	return att, nil
}

func (s *attendanceService) AdminUpdate(id uint, checkInAtStr, checkOutAtStr string) (*models.Attendance, error) {
	att, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("attendance record not found")
	}

	shiftCheckIn, shiftCheckOut := s.resolveShiftTimes(att.UserID, att.Date)

	if checkInAtStr != "" {
		t, perr := time.ParseInLocation("2006-01-02T15:04", checkInAtStr, cambodiaLoc)
		if perr != nil {
			return nil, errors.New("check_in_at must be a valid date/time")
		}
		att.CheckInAt = &t
		att.CheckInStatus = computeTimelinessStatus(t.Format("15:04"), shiftCheckIn)
	}
	if checkOutAtStr != "" {
		t, perr := time.ParseInLocation("2006-01-02T15:04", checkOutAtStr, cambodiaLoc)
		if perr != nil {
			return nil, errors.New("check_out_at must be a valid date/time")
		}
		att.CheckOutAt = &t
		att.CheckOutStatus = computeCheckOutTimeliness(t.Format("15:04"), shiftCheckOut)
	}

	if att.CheckInAt != nil && att.CheckOutAt != nil && att.CheckOutAt.Before(*att.CheckInAt) {
		return nil, errors.New("check-out time cannot be before check-in time")
	}

	if err := s.repo.Update(att); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *attendanceService) List(filter repositories.AttendanceFilter, page, pageSize int) ([]models.Attendance, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

// normalizeDate strips anything after the first 10 characters — the
// MySQL driver sometimes round-trips a DATE column as
// "2026-07-01T00:00:00+07:00" instead of a clean "2026-07-01". Comparing
// those raw strings directly (as the earlier frontend-only version of
// this report did) silently fails every lookup, since the two forms never
// string-equal each other even though they're the same calendar day. This
// keeps every date used as a map key or range bound in the same clean
// "YYYY-MM-DD" shape regardless of which form the DB actually returned.
func normalizeDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// dateRangeList returns every calendar date (YYYY-MM-DD) from `from` to
// `to` inclusive.
func dateRangeList(from, to string) ([]string, error) {
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, errors.New("date_from must be YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, errors.New("date_to must be YYYY-MM-DD")
	}
	var out []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		out = append(out, d.Format("2006-01-02"))
	}
	return out, nil
}

func (s *attendanceService) Summary(callerID uint, dateFrom, dateTo string, userID, branchID uint) ([]UserAttendanceSummary, error) {
	if dateFrom == "" || dateTo == "" {
		return nil, errors.New("date_from and date_to are required")
	}
	dates, err := dateRangeList(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	// Scoped users — same visibility rules as every other "who can I see"
	// dropdown in this app.
	scopedUsers, err := s.userRepo.GetUsersInScope(callerID)
	if err != nil {
		return nil, err
	}
	// Only active staff — a deactivated account shouldn't show up in a
	// current attendance report.
	{
		filtered := scopedUsers[:0]
		for _, u := range scopedUsers {
			if u.IsActive {
				filtered = append(filtered, u)
			}
		}
		scopedUsers = filtered
	}
	if userID != 0 {
		filtered := scopedUsers[:0]
		for _, u := range scopedUsers {
			if u.ID == userID {
				filtered = append(filtered, u)
			}
		}
		scopedUsers = filtered
	}

	// Attendance rows in range.
	attFilter := repositories.AttendanceFilter{DateFrom: dateFrom, DateTo: dateTo, BranchID: branchID}
	if userID != 0 {
		attFilter.UserID = userID
	}
	attRows, _, err := s.repo.List(attFilter, 1, 10000)
	if err != nil {
		return nil, err
	}

	// Leave rows in range — fetch every status (no Status filter on the
	// query), then keep only approved and pending. This used to be
	// approved-only; broadened to match the payroll export's rule, since
	// both now go through this same function and a pending leave request
	// should still show up here rather than being silently excluded
	// until someone gets around to approving it.
	leaveFilter := repositories.LeaveRequestFilter{DateFrom: dateFrom, DateTo: dateTo, BranchID: branchID}
	if userID != 0 {
		leaveFilter.UserID = userID
	}
	leaveRowsRaw, _, err := s.leaveRepo.List(leaveFilter, 1, 10000)
	if err != nil {
		return nil, err
	}
	leaveRows := leaveRowsRaw[:0]
	for _, l := range leaveRowsRaw {
		if l.Status == models.LeaveRequestApproved || l.Status == models.LeaveRequestPending {
			leaveRows = append(leaveRows, l)
		}
	}

	// user_id -> normalized date -> attendance row
	attByUser := map[uint]map[string]models.Attendance{}
	// user_id -> set of branch names seen
	branchByUser := map[uint]map[string]bool{}
	addBranch := func(userID uint, branch *models.Branch) {
		if branch == nil || branch.Name == "" {
			return
		}
		if branchByUser[userID] == nil {
			branchByUser[userID] = map[string]bool{}
		}
		branchByUser[userID][branch.Name] = true
	}
	for _, a := range attRows {
		if attByUser[a.UserID] == nil {
			attByUser[a.UserID] = map[string]models.Attendance{}
		}
		attByUser[a.UserID][normalizeDate(a.Date)] = a
		addBranch(a.UserID, a.Branch)
	}

	type dateRange struct {
		from, to string
		dayType  string // "full", "half_morning", "half_afternoon"
	}
	leaveByUser := map[uint][]dateRange{}
	for _, l := range leaveRows {
		leaveByUser[l.UserID] = append(leaveByUser[l.UserID], dateRange{normalizeDate(l.DateFrom), normalizeDate(l.DateTo), string(l.DayType)})
		addBranch(l.UserID, l.Branch)
	}

	// Branch filter is applied here — AFTER attendance/leave have been
	// fetched (already narrowed to this branch via attFilter/leaveFilter
	// above) — rather than purely on u.Branches. A user's own
	// user_branches assignment isn't always reliable/present, so a user
	// is kept if EITHER they're formally assigned to the branch OR they
	// actually have an attendance/leave record logged at it. Relying on
	// u.Branches alone caused every user to be excluded whenever that
	// assignment was missing, even for people who clearly worked there.
	if branchID != 0 {
		filtered := scopedUsers[:0]
		for _, u := range scopedUsers {
			assigned := false
			for _, b := range u.Branches {
				if b.ID == branchID {
					assigned = true
					break
				}
			}
			_, hasAtt := attByUser[u.ID]
			_, hasLeave := leaveByUser[u.ID]
			if assigned || hasAtt || hasLeave {
				filtered = append(filtered, u)
			}
		}
		scopedUsers = filtered
	}

	out := make([]UserAttendanceSummary, 0, len(scopedUsers))
	for _, u := range scopedUsers {
		var branchNames []string
		if len(u.Branches) > 0 {
			for _, b := range u.Branches {
				branchNames = append(branchNames, b.Name)
			}
		} else {
			// Fallback for a user with no branch assignment on record —
			// still show whatever branch their actual attendance/leave
			// rows happened to be logged under, if any.
			for name := range branchByUser[u.ID] {
				branchNames = append(branchNames, name)
			}
		}
		sort.Strings(branchNames)

		// Hide anyone with no branch info at all — neither an actual
		// assignment nor a record placing them at one.
		if len(branchNames) == 0 {
			continue
		}

		summary := UserAttendanceSummary{
			UserID:      u.ID,
			UserName:    u.Name,
			BranchNames: strings.Join(branchNames, ", "),
			Days:        make([]DaySummary, 0, len(dates)),
		}

		for _, d := range dates {
			t, _ := time.Parse("2006-01-02", d)
			ds := DaySummary{Date: d, Weekday: t.Weekday().String()[:3]}

			var statusParts []string
			hasAttendance := false

			if att, ok := attByUser[u.ID][d]; ok && att.CheckInAt != nil {
				hasAttendance = true
				statusParts = append(statusParts, "ATTEND")
				summary.Attend++
				ci := att.CheckInAt.Format("15:04")
				ds.CheckInAt = &ci
				ds.CheckInStatus = att.CheckInStatus
				ds.CheckInViaActivity = att.CheckInViaActivity
				if att.CheckOutAt != nil {
					co := att.CheckOutAt.Format("15:04")
					ds.CheckOutAt = &co
					ds.CheckOutStatus = att.CheckOutStatus
					ds.CheckOutViaActivity = att.CheckOutViaActivity
				} else {
					// Checked in but never checked out — only half a
					// completed day in practice, so it still counts as
					// ATTEND but also docks 0.5 toward Absent rather than
					// being silently treated as a full, ordinary day.
					summary.Absent += 0.5
				}
				ds.CheckInReason = att.CheckInReason
				ds.CheckOutReason = att.CheckOutReason
			}

			// Checked independently of attendance — NOT an else branch —
			// so a half-day leave still shows up even on a day the person
			// also attended for the other half (e.g. leave in the
			// morning, worked the afternoon). The earlier version treated
			// these as mutually exclusive and silently dropped the leave
			// whenever attendance existed for the same day.
			var dayType string
			leaveMatched := false
			for _, r := range leaveByUser[u.ID] {
				if d >= r.from && d <= r.to {
					dayType = r.dayType
					leaveMatched = true
					break
				}
			}
			switch dayType {
			case "half_morning":
				statusParts = append(statusParts, "Half Day - Morning")
				summary.Leave += 0.5
				// The uncovered half of the day only counts against
				// Absent if they didn't ALSO attend — attending the
				// other half means the day is fully accounted for.
				if !hasAttendance {
					summary.Absent += 0.5
				}
			case "half_afternoon":
				statusParts = append(statusParts, "Half Day - Afternoon")
				summary.Leave += 0.5
				if !hasAttendance {
					summary.Absent += 0.5
				}
			default:
				// Covers dayType == "full" as well as any
				// empty/unrecognized value on a date that DID match a
				// leave range — e.g. legacy leave requests created
				// before DayType existed, or a migration gap that left
				// it blank instead of the column's own "full" default.
				// Requiring an exact "full" match here silently dropped
				// these from Leave entirely; matching on "a leave range
				// covered this date and it wasn't half-day" is more
				// robust.
				if leaveMatched {
					statusParts = append(statusParts, "LEAVE")
					summary.Leave++
				}
			}

			if len(statusParts) == 0 {
				ds.Status = "ABSENT"
				summary.Absent++
			} else {
				ds.Status = strings.Join(statusParts, " + ")
			}
			summary.Days = append(summary.Days, ds)
		}
		out = append(out, summary)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].UserName < out[j].UserName })
	return out, nil
}
