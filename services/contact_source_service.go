package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type ContactSourceService interface {
	Create(createdByID uint, scopeIDs []uint, x *models.ContactSource) (*models.ContactSource, error)
	List(scopeIDs []uint, showAll bool) ([]models.ContactSource, error)
	ListForUser(userID uint, showAll bool) ([]models.ContactSource, error)
	GetByID(id uint, scopeIDs []uint) (*models.ContactSource, error)
	Update(id uint, scopeIDs []uint, x *models.ContactSource) (*models.ContactSource, error)
	Delete(id uint, scopeIDs []uint) error
}

type contactSourceService struct {
	repo repositories.ContactSourceRepository
}

func NewContactSourceService(repo repositories.ContactSourceRepository) ContactSourceService {
	return &contactSourceService{repo}
}

func (s *contactSourceService) Create(createdByID uint, scopeIDs []uint, x *models.ContactSource) (*models.ContactSource, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, scopeIDs)
}

func (s *contactSourceService) List(scopeIDs []uint, showAll bool) ([]models.ContactSource, error) {
	return s.repo.List(scopeIDs, showAll)
}

func (s *contactSourceService) ListForUser(userID uint, showAll bool) ([]models.ContactSource, error) {
	return s.repo.ListForUser(userID, showAll)
}

func (s *contactSourceService) GetByID(id uint, scopeIDs []uint) (*models.ContactSource, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact source not found")
		}
		return nil, err
	}
	return x, nil
}

func (s *contactSourceService) Update(id uint, scopeIDs []uint, upd *models.ContactSource) (*models.ContactSource, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("contact source not found")
	}

	// Only copy client-editable fields so CreatedByID/CreatedAt/ID aren't
	// clobbered by a full struct overwrite.
	x.Name = upd.Name
	x.Description = upd.Description
	x.Icon = upd.Icon
	x.BranchID = upd.BranchID
	x.IsActive = upd.IsActive

	if err := s.repo.Update(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *contactSourceService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("contact source not found")
	}
	return s.repo.Delete(id, scopeIDs)
}
