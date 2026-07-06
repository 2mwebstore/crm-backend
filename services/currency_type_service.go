package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type CurrencyTypeService interface {
	Create(createdByID uint, scopeIDs []uint, x *models.CurrencyType) (*models.CurrencyType, error)
	List(scopeIDs []uint, showAll bool) ([]models.CurrencyType, error)
	GetByID(id uint, scopeIDs []uint) (*models.CurrencyType, error)
	Update(id uint, scopeIDs []uint, x *models.CurrencyType) (*models.CurrencyType, error)
	Delete(id uint, scopeIDs []uint) error
}

type currencyTypeService struct {
	repo repositories.CurrencyTypeRepository
}

func NewCurrencyTypeService(repo repositories.CurrencyTypeRepository) CurrencyTypeService {
	return &currencyTypeService{repo}
}

func (s *currencyTypeService) Create(createdByID uint, scopeIDs []uint, x *models.CurrencyType) (*models.CurrencyType, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, scopeIDs)
}

func (s *currencyTypeService) List(scopeIDs []uint, showAll bool) ([]models.CurrencyType, error) {
	return s.repo.List(scopeIDs, showAll)
}

func (s *currencyTypeService) GetByID(id uint, scopeIDs []uint) (*models.CurrencyType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("currency not found")
		}
		return nil, err
	}
	return x, nil
}

func (s *currencyTypeService) Update(id uint, scopeIDs []uint, upd *models.CurrencyType) (*models.CurrencyType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("currency not found")
	}
	*x = *upd
	x.ID = uint(id)
	if err := s.repo.Update(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *currencyTypeService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("currency not found")
	}
	return s.repo.Delete(id, scopeIDs)
}
