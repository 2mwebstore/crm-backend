package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type BonusOptionTypeService interface {
	Create(createdByID uint, scopeIDs []uint, x *models.BonusOptionType) (*models.BonusOptionType, error)
	List(scopeIDs []uint, showAll bool) ([]models.BonusOptionType, error)
	ListForUser(userID uint, showAll bool, branchID *uint) ([]models.BonusOptionType, error)
	GetByID(id uint, scopeIDs []uint) (*models.BonusOptionType, error)
	Update(id uint, scopeIDs []uint, x *models.BonusOptionType) (*models.BonusOptionType, error)
	Delete(id uint, scopeIDs []uint) error
	Preview(id uint, scopeIDs []uint, baseValue float64) (float64, error)
}

type bonusOptionTypeService struct {
	repo repositories.BonusOptionTypeRepository
}

func NewBonusOptionTypeService(repo repositories.BonusOptionTypeRepository) BonusOptionTypeService {
	return &bonusOptionTypeService{repo}
}

func (s *bonusOptionTypeService) Create(createdByID uint, scopeIDs []uint, x *models.BonusOptionType) (*models.BonusOptionType, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, scopeIDs)
}

func (s *bonusOptionTypeService) List(scopeIDs []uint, showAll bool) ([]models.BonusOptionType, error) {
	return s.repo.List(scopeIDs, showAll)
}

func (s *bonusOptionTypeService) ListForUser(userID uint, showAll bool, branchID *uint) ([]models.BonusOptionType, error) {
	return s.repo.ListForUser(userID, showAll, branchID)
}

func (s *bonusOptionTypeService) GetByID(id uint, scopeIDs []uint) (*models.BonusOptionType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bonus option not found")
		}
		return nil, err
	}
	return x, nil
}

func (s *bonusOptionTypeService) Update(id uint, scopeIDs []uint, upd *models.BonusOptionType) (*models.BonusOptionType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("bonus option not found")
	}

	// Only copy client-editable fields so CreatedByID/CreatedAt/ID aren't
	// clobbered by a full struct overwrite.
	x.Name = upd.Name
	x.BranchID = upd.BranchID
	x.SortOrder = upd.SortOrder
	x.IsActive = upd.IsActive

	if err := s.repo.Update(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *bonusOptionTypeService) Preview(id uint, scopeIDs []uint, baseValue float64) (float64, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("bonus option not found")
		}
		return 0, err
	}
	// TODO: replace with real calculation once BonusOptionType's fields
	// (e.g. percentage vs fixed amount, min/max caps) are known.
	_ = x
	return baseValue, nil
}

func (s *bonusOptionTypeService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("bonus option not found")
	}
	return s.repo.Delete(id, scopeIDs)
}
