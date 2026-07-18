package services

import (
	"errors"
	"time"

	"crm-backend/models"
	"crm-backend/repositories"
)

type UserScheduleOverrideInput struct {
	UserID            uint
	DateFrom          string
	DateTo            string
	ShiftCheckInTime  *string
	ShiftCheckOutTime *string
	Reason            string
}

type UserScheduleOverrideService interface {
	Create(createdByID uint, input UserScheduleOverrideInput) (*models.UserScheduleOverride, error)
	List(filter repositories.UserScheduleOverrideFilter, page, pageSize int) ([]models.UserScheduleOverride, int64, error)
	ListForUser(userID uint) ([]models.UserScheduleOverride, error)
	Update(id uint, input UserScheduleOverrideInput) (*models.UserScheduleOverride, error)
	Delete(id uint) error
}

type userScheduleOverrideService struct {
	repo repositories.UserScheduleOverrideRepository
}

func NewUserScheduleOverrideService(repo repositories.UserScheduleOverrideRepository) UserScheduleOverrideService {
	return &userScheduleOverrideService{repo}
}

func (s *userScheduleOverrideService) validate(input UserScheduleOverrideInput) error {
	if input.UserID == 0 {
		return errors.New("user_id is required")
	}
	if input.DateFrom == "" || input.DateTo == "" {
		return errors.New("date_from and date_to are required")
	}
	fromDate, err := time.Parse("2006-01-02", input.DateFrom)
	if err != nil {
		return errors.New("date_from must be YYYY-MM-DD")
	}
	toDate, err := time.Parse("2006-01-02", input.DateTo)
	if err != nil {
		return errors.New("date_to must be YYYY-MM-DD")
	}
	if toDate.Before(fromDate) {
		return errors.New("date_to cannot be before date_from")
	}
	if input.ShiftCheckInTime == nil && input.ShiftCheckOutTime == nil {
		return errors.New("at least one of shift_check_in_time or shift_check_out_time is required")
	}
	if err := validateTime24h("shift_check_in_time", input.ShiftCheckInTime); err != nil {
		return err
	}
	if err := validateTime24h("shift_check_out_time", input.ShiftCheckOutTime); err != nil {
		return err
	}
	return nil
}

func (s *userScheduleOverrideService) Create(createdByID uint, input UserScheduleOverrideInput) (*models.UserScheduleOverride, error) {
	if err := s.validate(input); err != nil {
		return nil, err
	}
	override := &models.UserScheduleOverride{
		UserID:            input.UserID,
		DateFrom:          input.DateFrom,
		DateTo:            input.DateTo,
		ShiftCheckInTime:  input.ShiftCheckInTime,
		ShiftCheckOutTime: input.ShiftCheckOutTime,
		Reason:            input.Reason,
		CreatedByID:       createdByID,
	}
	if err := s.repo.Create(override); err != nil {
		return nil, err
	}
	return s.repo.FindByID(override.ID)
}

func (s *userScheduleOverrideService) List(filter repositories.UserScheduleOverrideFilter, page, pageSize int) ([]models.UserScheduleOverride, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

func (s *userScheduleOverrideService) ListForUser(userID uint) ([]models.UserScheduleOverride, error) {
	return s.repo.ListForUser(userID)
}

func (s *userScheduleOverrideService) Update(id uint, input UserScheduleOverrideInput) (*models.UserScheduleOverride, error) {
	if err := s.validate(input); err != nil {
		return nil, err
	}
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("schedule override not found")
	}
	existing.UserID = input.UserID
	existing.DateFrom = input.DateFrom
	existing.DateTo = input.DateTo
	existing.ShiftCheckInTime = input.ShiftCheckInTime
	existing.ShiftCheckOutTime = input.ShiftCheckOutTime
	existing.Reason = input.Reason
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *userScheduleOverrideService) Delete(id uint) error {
	return s.repo.Delete(id)
}
