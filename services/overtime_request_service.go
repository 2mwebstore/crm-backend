package services

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"crm-backend/models"
	"crm-backend/repositories"
)

type OvertimeRequestInput struct {
	BranchID  *uint
	Date      string
	StartTime *string
	EndTime   *string
	Reason    string
}

type OvertimeRequestService interface {
	Create(userID uint, input OvertimeRequestInput) (*models.OvertimeRequest, error)
	List(filter repositories.OvertimeRequestFilter, page, pageSize int) ([]models.OvertimeRequest, int64, error)
	GetByID(id uint) (*models.OvertimeRequest, error)
	Approve(id uint, approverID uint) (*models.OvertimeRequest, error)
	Reject(id uint, approverID uint, reason string) (*models.OvertimeRequest, error)
	// EditReason lets the ORIGINAL requester (not an admin — self-service,
	// same as Create) change their own request's reason, but only while
	// it's still pending. Once approved/rejected/cancelled, the record is
	// final.
	EditReason(userID, id uint, reason string) (*models.OvertimeRequest, error)
	// Cancel lets the ORIGINAL requester withdraw their own pending
	// request — sets status to "cancelled" rather than deleting the row,
	// so there's still a record it existed.
	Cancel(userID, id uint) (*models.OvertimeRequest, error)
}

type overtimeRequestService struct {
	repo     repositories.OvertimeRequestRepository
	userRepo repositories.UserRepository
}

func NewOvertimeRequestService(repo repositories.OvertimeRequestRepository, userRepo repositories.UserRepository) OvertimeRequestService {
	return &overtimeRequestService{repo, userRepo}
}

// time24hPattern matches a strict 24-hour "HH:MM" string (e.g. "18:00") —
// used to reject anything that isn't that exact shape, such as a leftover
// "6:00 PM"-style value, before it ever reaches the database.
var time24hPattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func validateTime24h(label string, t *string) error {
	if t == nil || *t == "" {
		return nil
	}
	v := strings.TrimSpace(*t)
	if !time24hPattern.MatchString(v) {
		return errors.New(label + " must be a 24-hour \"HH:MM\" time (e.g. \"18:00\"), not AM/PM")
	}
	return nil
}

// computeDuration returns hours between start and end (both "HH:MM"), or
// nil if either is missing/invalid. An end time earlier than start is
// treated as crossing midnight (e.g. 22:00 -> 02:00 = 4 hours) rather than
// rejected, since overtime commonly runs past midnight.
func computeDuration(start, end *string) *float64 {
	if start == nil || end == nil || *start == "" || *end == "" {
		return nil
	}
	toMinutes := func(t string) (int, bool) {
		parts := strings.Split(t, ":")
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
	startMin, ok1 := toMinutes(*start)
	endMin, ok2 := toMinutes(*end)
	if !ok1 || !ok2 {
		return nil
	}
	diff := endMin - startMin
	if diff <= 0 {
		diff += 24 * 60
	}
	hours := float64(diff) / 60.0
	return &hours
}

func (s *overtimeRequestService) Create(userID uint, input OvertimeRequestInput) (*models.OvertimeRequest, error) {
	if input.Date == "" {
		return nil, errors.New("date is required")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("reason is required for an overtime request")
	}
	if err := validateTime24h("start_time", input.StartTime); err != nil {
		return nil, err
	}
	if err := validateTime24h("end_time", input.EndTime); err != nil {
		return nil, err
	}

	// Cross Day (Night Shift) users are exempt from the duplicate-date
	// check — see the same reasoning in LeaveRequestService.Create.
	isCrossDay := false
	if user, uerr := s.userRepo.FindByID(userID); uerr == nil {
		isCrossDay = user.ShiftType == models.ShiftTypeCrossDay
	}
	if !isCrossDay {
		overlapping, err := s.repo.HasOverlapping(userID, input.Date, 0)
		if err != nil {
			return nil, err
		}
		if overlapping {
			return nil, errors.New("you already have an overtime request for this date — each date can only be requested once")
		}
	}

	req := &models.OvertimeRequest{
		UserID:    userID,
		BranchID:  input.BranchID,
		Date:      input.Date,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Duration:  computeDuration(input.StartTime, input.EndTime),
		Reason:    input.Reason,
		Status:    models.OvertimeRequestPending,
	}
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(req.ID)
}

func (s *overtimeRequestService) List(filter repositories.OvertimeRequestFilter, page, pageSize int) ([]models.OvertimeRequest, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

func (s *overtimeRequestService) GetByID(id uint) (*models.OvertimeRequest, error) {
	return s.repo.FindByID(id)
}

func (s *overtimeRequestService) Approve(id uint, approverID uint) (*models.OvertimeRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.Status != models.OvertimeRequestPending {
		return nil, errors.New("only a pending request can be approved")
	}
	now := nowInCambodia()
	req.Status = models.OvertimeRequestApproved
	req.ApprovedByID = &approverID
	req.ApprovedAt = &now
	req.RejectReason = ""
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *overtimeRequestService) Reject(id uint, approverID uint, reason string) (*models.OvertimeRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.Status != models.OvertimeRequestPending {
		return nil, errors.New("only a pending request can be rejected")
	}
	now := nowInCambodia()
	req.Status = models.OvertimeRequestRejected
	req.ApprovedByID = &approverID
	req.ApprovedAt = &now
	req.RejectReason = reason
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *overtimeRequestService) EditReason(userID, id uint, reason string) (*models.OvertimeRequest, error) {
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
	if req.Status != models.OvertimeRequestPending {
		return nil, errors.New("only a pending request can be edited")
	}
	req.Reason = reason
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *overtimeRequestService) Cancel(userID, id uint) (*models.OvertimeRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("request not found")
	}
	if req.UserID != userID {
		return nil, errors.New("you can only cancel your own request")
	}
	if req.Status != models.OvertimeRequestPending {
		return nil, errors.New("only a pending request can be cancelled")
	}
	req.Status = models.OvertimeRequestCancelled
	if err := s.repo.Update(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}
