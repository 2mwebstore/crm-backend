package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type BranchService interface {
	Create(createdByID uint, name, code, description string) (*models.Branch, error)
	List(showAll bool) ([]models.Branch, error)
	ListForUser(userID uint, showAll bool) ([]models.Branch, error)
	GetByID(id uint) (*models.Branch, error)
	Update(id uint, name, code, description string, isActive bool) (*models.Branch, error)
	Delete(id uint) error
}

type branchService struct{ repo repositories.BranchRepository }

func NewBranchService(repo repositories.BranchRepository) BranchService {
	return &branchService{repo}
}

func (s *branchService) Create(createdByID uint, name, code, description string) (*models.Branch, error) {
	b := &models.Branch{Name: name, Code: code, Description: description, IsActive: true, CreatedByID: createdByID}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	return s.repo.FindByID(b.ID)
}

func (s *branchService) List(showAll bool) ([]models.Branch, error) {
	return s.repo.List(showAll)
}

func (s *branchService) GetByID(id uint) (*models.Branch, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("branch not found")
		}
		return nil, err
	}
	return b, nil
}

func (s *branchService) Update(id uint, name, code, description string, isActive bool) (*models.Branch, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("branch not found")
	}
	if name != "" {
		b.Name = name
	}
	if code != "" {
		b.Code = code
	}
	b.Description = description
	b.IsActive = isActive
	if err := s.repo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("branch not found")
	}
	return s.repo.Delete(id)
}

func (s *branchService) ListForUser(userID uint, showAll bool) ([]models.Branch, error) {
	return s.repo.ListForUser(userID, showAll)
}
