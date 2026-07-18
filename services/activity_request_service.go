package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type ActivityRequestInput struct {
	BranchID uint
	Date     string
	Reason   string
}

type ActivityRequestService interface {
	Create(userID uint, input ActivityRequestInput) (*models.ActivityRequest, error)
	List(filter repositories.ActivityRequestFilter, page, pageSize int) ([]models.ActivityRequest, int64, error)
	GetByID(id uint) (*models.ActivityRequest, error)
}

type activityRequestService struct {
	repo           repositories.ActivityRequestRepository
	attendanceRepo repositories.AttendanceRepository
}

func NewActivityRequestService(repo repositories.ActivityRequestRepository, attendanceRepo repositories.AttendanceRepository) ActivityRequestService {
	return &activityRequestService{repo, attendanceRepo}
}

// Create always auto-approves (see the model) — Activity is self-declared
// and needs to be effective immediately, unlike Leave/Overtime. Blocked
// outright if today's attendance already has BOTH a check-in and
// check-out — nothing left for an activity request to drive at that point
// (see ActivityRequestController.Create for the part that actually drives
// Check In/Check Out once this succeeds).
func (s *activityRequestService) Create(userID uint, input ActivityRequestInput) (*models.ActivityRequest, error) {
	if input.BranchID == 0 {
		return nil, errors.New("branch_id is required")
	}
	if input.Date == "" {
		return nil, errors.New("date is required")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("reason is required for an activity request")
	}

	today := todayDateString()
	if input.Date == today {
		att, err := s.attendanceRepo.FindByUserAndDate(userID, today)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil && att.CheckInAt != nil && att.CheckOutAt != nil {
			return nil, errors.New("you've already checked in and checked out today — an Activity request isn't needed")
		}
	}

	req := &models.ActivityRequest{
		UserID:     userID,
		BranchID:   input.BranchID,
		Date:       input.Date,
		Reason:     input.Reason,
		Status:     "approved",
		ApprovedAt: nowInCambodia(),
	}
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(req.ID)
}

func (s *activityRequestService) List(filter repositories.ActivityRequestFilter, page, pageSize int) ([]models.ActivityRequest, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

func (s *activityRequestService) GetByID(id uint) (*models.ActivityRequest, error) {
	return s.repo.FindByID(id)
}
