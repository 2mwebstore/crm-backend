package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"crm-backend/models"
	"crm-backend/repositories"
)

type LeaveRequestInput struct {
	BranchID    *uint
	LeaveTypeID uint
	DayType     models.LeaveRequestDayType
	DateFrom    string
	DateTo      string
	Reason      string
}

type LeaveRequestService interface {
	Create(userID uint, input LeaveRequestInput) (*models.LeaveRequest, error)
	List(filter repositories.LeaveRequestFilter, page, pageSize int) ([]models.LeaveRequest, int64, error)
	GetByID(id uint) (*models.LeaveRequest, error)
	Approve(id uint, approverID uint) (*models.LeaveRequest, error)
	Reject(id uint, approverID uint, reason string) (*models.LeaveRequest, error)
	// EditReason lets the ORIGINAL requester (self-service, same as
	// Create) change their own request's reason, but only while it's
	// still pending — matches OvertimeRequestService.EditReason.
	EditReason(userID, id uint, reason string) (*models.LeaveRequest, error)
	// Cancel lets the ORIGINAL requester withdraw their own pending
	// request — sets status to "cancelled" rather than deleting the row.
	Cancel(userID, id uint) (*models.LeaveRequest, error)
}

type leaveRequestService struct {
	repo          repositories.LeaveRequestRepository
	leaveTypeRepo repositories.LeaveTypeRepository
	userRepo      repositories.UserRepository
}

func NewLeaveRequestService(repo repositories.LeaveRequestRepository, leaveTypeRepo repositories.LeaveTypeRepository, userRepo repositories.UserRepository) LeaveRequestService {
	return &leaveRequestService{repo, leaveTypeRepo, userRepo}
}

func (s *leaveRequestService) Create(userID uint, input LeaveRequestInput) (*models.LeaveRequest, error) {
	if input.LeaveTypeID == 0 {
		return nil, errors.New("leave_type_id is required")
	}
	if input.DateFrom == "" || input.DateTo == "" {
		return nil, errors.New("date_from and date_to are required")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("reason is required for a leave request")
	}
	fromDate, err := time.Parse("2006-01-02", input.DateFrom)
	if err != nil {
		return nil, errors.New("date_from must be YYYY-MM-DD")
	}
	toDate, err := time.Parse("2006-01-02", input.DateTo)
	if err != nil {
		return nil, errors.New("date_to must be YYYY-MM-DD")
	}
	if toDate.Before(fromDate) {
		return nil, errors.New("date_to cannot be before date_from")
	}

	dayType := input.DayType
	if dayType == "" {
		dayType = models.LeaveDayFull
	}
	isHalfDay := dayType == models.LeaveDayHalfMorning || dayType == models.LeaveDayHalfAfternoon
	if isHalfDay && input.DateFrom != input.DateTo {
		return nil, errors.New("a half day request must use a single date (date_from must equal date_to)")
	}

	var requestedDays float64
	if isHalfDay {
		requestedDays = 0.5
	} else {
		requestedDays = float64(int(toDate.Sub(fromDate).Hours()/24) + 1)
	}

	// Cross Day (Night Shift) users are exempt from the duplicate-date
	// check — their schedule is designed to straddle midnight, so two
	// requests on "consecutive" calendar dates aren't necessarily a real
	// duplicate the way they would be for a Normal Day user.
	isCrossDay := false
	if user, uerr := s.userRepo.FindByID(userID); uerr == nil {
		isCrossDay = user.ShiftType == models.ShiftTypeCrossDay
	}
	if !isCrossDay {
		overlapping, err := s.repo.HasOverlapping(userID, input.DateFrom, input.DateTo, 0)
		if err != nil {
			return nil, err
		}
		if overlapping {
			return nil, errors.New("you already have a leave request covering one of these dates — each date can only be requested once")
		}
	}

	if err := s.checkLimits(userID, input.LeaveTypeID, fromDate, requestedDays, 0); err != nil {
		return nil, err
	}

	req := &models.LeaveRequest{
		UserID:      userID,
		BranchID:    input.BranchID,
		LeaveTypeID: input.LeaveTypeID,
		DayType:     dayType,
		DateFrom:    input.DateFrom,
		DateTo:      input.DateTo,
		Duration:    requestedDays,
		Reason:      input.Reason,
		Status:      models.LeaveRequestPending,
	}
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(req.ID)
}

// checkLimits enforces LeaveType.AnnualLimit/MonthlyLimit — either nil
// means unlimited for that period. requestedDays is 0.5 for a Half Day
// request, otherwise the full day count.
func (s *leaveRequestService) checkLimits(userID, leaveTypeID uint, from time.Time, requestedDays float64, excludeID uint) error {
	lt, err := s.leaveTypeRepo.FindByID(leaveTypeID, userID)
	if err != nil {
		return errors.New("leave type not found")
	}

	if lt.MonthlyLimit != nil {
		monthStart := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		monthEnd := monthStart.AddDate(0, 1, -1)
		used, err := s.repo.SumDaysInPeriod(userID, leaveTypeID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"), excludeID)
		if err != nil {
			return err
		}
		if used+requestedDays > float64(*lt.MonthlyLimit) {
			return fmt.Errorf("this exceeds the monthly limit for %s: %g day(s) already used this month, limit is %d", lt.Name, used, *lt.MonthlyLimit)
		}
	}

	if lt.AnnualLimit != nil {
		yearStart := time.Date(from.Year(), 1, 1, 0, 0, 0, 0, from.Location())
		yearEnd := time.Date(from.Year(), 12, 31, 0, 0, 0, 0, from.Location())
		used, err := s.repo.SumDaysInPeriod(userID, leaveTypeID, yearStart.Format("2006-01-02"), yearEnd.Format("2006-01-02"), excludeID)
		if err != nil {
			return err
		}
		if used+requestedDays > float64(*lt.AnnualLimit) {
			return fmt.Errorf("this exceeds the annual limit for %s: %g day(s) already used this year, limit is %d", lt.Name, used, *lt.AnnualLimit)
		}
	}

	return nil
}

func (s *leaveRequestService) List(filter repositories.LeaveRequestFilter, page, pageSize int) ([]models.LeaveRequest, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

func (s *leaveRequestService) GetByID(id uint) (*models.LeaveRequest, error) {
	return s.repo.FindByID(id)
}

func (s *leaveRequestService) Approve(id uint, approverID uint) (*models.LeaveRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.Status != models.LeaveRequestPending {
		return nil, errors.New("only a pending request can be approved")
	}
	now := nowInCambodia()
	req.Status = models.LeaveRequestApproved
	req.ApprovedByID = &approverID
	req.ApprovedAt = &now
	req.RejectReason = ""
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *leaveRequestService) Reject(id uint, approverID uint, reason string) (*models.LeaveRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.Status != models.LeaveRequestPending {
		return nil, errors.New("only a pending request can be rejected")
	}
	now := nowInCambodia()
	req.Status = models.LeaveRequestRejected
	req.ApprovedByID = &approverID
	req.ApprovedAt = &now
	req.RejectReason = reason
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *leaveRequestService) EditReason(userID, id uint, reason string) (*models.LeaveRequest, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("reason is required")
	}
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.UserID != userID {
		return nil, errors.New("you can only edit your own request")
	}
	if req.Status != models.LeaveRequestPending {
		return nil, errors.New("only a pending request can be edited")
	}
	req.Reason = reason
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// Cancel sets status to "cancelled" — since SumDaysInPeriod only counts
// pending/approved requests toward a leave type's limit, cancelling one
// automatically frees up that day count for future requests without any
// extra bookkeeping needed here.
func (s *leaveRequestService) Cancel(userID, id uint) (*models.LeaveRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.UserID != userID {
		return nil, errors.New("you can only cancel your own request")
	}
	if req.Status != models.LeaveRequestPending {
		return nil, errors.New("only a pending request can be cancelled")
	}
	req.Status = models.LeaveRequestCancelled
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}
