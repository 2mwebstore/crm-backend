package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type LevelService interface {
	Create(createdByID uint, branchID uint, name, description, color string, sortOrder int) (*models.Level, error)
	ListForUser(userID uint, showAll bool) ([]models.Level, error)
	List(showAll bool) ([]models.Level, error)
	// GetByID, Update, and Delete take the caller's branch scopeIDs (nil
	// means unscoped / super admin access) so access can be enforced the
	// same way ListForUser does.
	GetByID(id uint, scopeIDs []uint) (*models.Level, error)
	Update(id uint, scopeIDs []uint, branchID uint, name, description, color string, sortOrder int, isActive bool) (*models.Level, error)
	Delete(id uint, scopeIDs []uint) error
}

type levelService struct{ repo repositories.LevelRepository }

func NewLevelService(repo repositories.LevelRepository) LevelService {
	return &levelService{repo}
}

func (s *levelService) Create(createdByID uint, branchID uint, name, description, color string, sortOrder int) (*models.Level, error) {
	if color == "" {
		color = "#6366f1"
	}
	l := &models.Level{BranchID: branchID, Name: name, Description: description, Color: color, SortOrder: sortOrder, IsActive: true}
	if err := s.repo.Create(l, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(l.ID, nil)
}

func (s *levelService) List(showAll bool) ([]models.Level, error) {
	return s.repo.List(showAll)
}

func (s *levelService) GetByID(id uint, scopeIDs []uint) (*models.Level, error) {
	l, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("level not found")
		}
		return nil, err
	}
	return l, nil
}

func (s *levelService) Update(id uint, scopeIDs []uint, branchID uint, name, description, color string, sortOrder int, isActive bool) (*models.Level, error) {
	l, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("level not found")
	}
	if name != "" {
		l.Name = name
	}
	if color != "" {
		l.Color = color
	}
	if branchID != 0 {
		l.BranchID = branchID
	}
	l.Description = description
	l.SortOrder = sortOrder
	l.IsActive = isActive
	if err := s.repo.Update(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *levelService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("level not found")
	}
	return s.repo.Delete(id, scopeIDs)
}

func (s *levelService) ListForUser(userID uint, showAll bool) ([]models.Level, error) {
	return s.repo.ListForUser(userID, showAll)
}
