package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

var ErrForbidden = errors.New("bank type not accessible")

type BankTypeService interface {
	Create(createdByID uint, x *models.BankType) (*models.BankType, error)
	ListForUser(userID uint, showAll bool) ([]models.BankType, error)
	List(showAll bool) ([]models.BankType, error)
	// GetByID, Update, and Delete take the caller's branch scopeIDs (nil
	// means unscoped / super admin access) so access can be enforced the
	// same way ListForUser does.
	GetByID(id uint, scopeIDs []uint) (*models.BankType, error)
	Update(id uint, scopeIDs []uint, x *models.BankType) (*models.BankType, error)
	Delete(id uint, scopeIDs []uint) error
}

type bankTypeService struct {
	repo repositories.BankTypeRepository
}

func NewBankTypeService(repo repositories.BankTypeRepository) BankTypeService {
	return &bankTypeService{repo}
}

func (s *bankTypeService) Create(createdByID uint, x *models.BankType) (*models.BankType, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, nil)
}

func (s *bankTypeService) List(showAll bool) ([]models.BankType, error) {
	return s.repo.List(showAll)
}

func (s *bankTypeService) GetByID(id uint, scopeIDs []uint) (*models.BankType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("bank type not found")
		}
		return nil, err
	}
	return x, nil
}

func (s *bankTypeService) Update(id uint, scopeIDs []uint, upd *models.BankType) (*models.BankType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("bank type not found")
	}

	// Only copy the fields a client is allowed to change. This preserves
	// ID, CreatedByID, CreatedAt, and any other bookkeeping fields that
	// were previously being wiped out by a full struct overwrite.
	x.Name = upd.Name
	x.BranchID = upd.BranchID
	x.SortOrder = upd.SortOrder
	x.IsActive = upd.IsActive

	if err := s.repo.Update(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *bankTypeService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("bank type not found")
	}
	return s.repo.Delete(id, scopeIDs)
}

func (s *bankTypeService) ListForUser(userID uint, showAll bool) ([]models.BankType, error) {
	return s.repo.ListForUser(userID, showAll)
}
