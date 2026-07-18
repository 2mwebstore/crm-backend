package services

import (
	"errors"
	"time"

	"crm-backend/models"
	"crm-backend/repositories"
)

type LeaveTypeService interface {
	Create(createdByID uint, x *models.LeaveType) (*models.LeaveType, error)
	List(callerID uint, showAll bool, branchID *uint) ([]models.LeaveType, error)
	GetByID(id uint, callerID uint) (*models.LeaveType, error)
	Update(id uint, callerID uint, upd *models.LeaveType) (*models.LeaveType, error)
	Delete(id uint) error
}

type leaveTypeService struct {
	repo             repositories.LeaveTypeRepository
	leaveRequestRepo repositories.LeaveRequestRepository
}

func NewLeaveTypeService(repo repositories.LeaveTypeRepository, leaveRequestRepo repositories.LeaveRequestRepository) LeaveTypeService {
	return &leaveTypeService{repo, leaveRequestRepo}
}

func (s *leaveTypeService) Create(createdByID uint, x *models.LeaveType) (*models.LeaveType, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, createdByID)
}

// List attaches each leave type's MonthlyUsed/AnnualUsed for callerID —
// how many days THIS caller has already used this month/year — so the
// Submit a Request form can show "2 of 3 used" instead of just the raw
// limit. Only computed for types that actually have a limit set, to avoid
// pointless extra queries for unlimited ones.
func (s *leaveTypeService) List(callerID uint, showAll bool, branchID *uint) ([]models.LeaveType, error) {
	items, err := s.repo.List(callerID, showAll, branchID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	yearEnd := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location())

	for i := range items {
		if items[i].MonthlyLimit != nil {
			used, err := s.leaveRequestRepo.SumDaysInPeriod(callerID, items[i].ID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"), 0)
			if err == nil {
				items[i].MonthlyUsed = &used
			}
		}
		if items[i].AnnualLimit != nil {
			used, err := s.leaveRequestRepo.SumDaysInPeriod(callerID, items[i].ID, yearStart.Format("2006-01-02"), yearEnd.Format("2006-01-02"), 0)
			if err == nil {
				items[i].AnnualUsed = &used
			}
		}
	}
	return items, nil
}

func (s *leaveTypeService) GetByID(id uint, callerID uint) (*models.LeaveType, error) {
	x, err := s.repo.FindByID(id, callerID)
	if err != nil {
		return nil, errors.New("leave type not found")
	}
	return x, nil
}

func (s *leaveTypeService) Update(id uint, callerID uint, upd *models.LeaveType) (*models.LeaveType, error) {
	x, err := s.repo.FindByID(id, callerID)
	if err != nil {
		return nil, errors.New("leave type not found")
	}
	upd.ID = x.ID
	upd.CreatedByID = x.CreatedByID
	if err := s.repo.Update(upd); err != nil {
		return nil, err
	}
	return upd, nil
}

func (s *leaveTypeService) Delete(id uint) error {
	return s.repo.Delete(id)
}
