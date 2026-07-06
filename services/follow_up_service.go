package services

import (
	"errors"

	"gorm.io/gorm"

	followupdto "crm-backend/dto/follow_up"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type FollowUpService interface {
	Create(createdByID uint, req followupdto.CreateRequest) (*models.ClientFollowUp, error)
	GetByID(id uint, scopeIDs []uint) (*models.ClientFollowUp, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter followupdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.ClientFollowUp, int64, error)
}

type followUpService struct {
	repo repositories.FollowUpRepository
	db   *gorm.DB
}

func NewFollowUpService(repo repositories.FollowUpRepository, db *gorm.DB) FollowUpService {
	return &followUpService{repo, db}
}

func (s *followUpService) Create(createdByID uint, req followupdto.CreateRequest) (*models.ClientFollowUp, error) {
	f := &models.ClientFollowUp{
		ClientID:      req.ClientID,
		BranchID:      req.BranchID,
		FollowUpAt:    req.FollowUpAt.Time,
		BonusOptionID: req.BonusOptionID,
		Interest:      req.Interest,
		GivenAccount:  req.GivenAccount,
		BankAccount:   req.BankAccount,
		Remark:        req.Remark,
		CreatedByID:   createdByID,
	}
	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	fu, err := s.repo.FindByID(f.ID, nil)
	if err != nil {
		return nil, err
	}
	return fu, nil
}

func (s *followUpService) GetByID(id uint, scopeIDs []uint) (*models.ClientFollowUp, error) {
	f, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("follow-up not found")
		}
		return nil, err
	}
	return f, nil
}

func (s *followUpService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("follow-up not found")
	}
	return s.repo.Delete(id, scopeIDs)
}

func (s *followUpService) List(filter followupdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.ClientFollowUp, int64, error) {
	return s.repo.List(filter, p, userID)
}
