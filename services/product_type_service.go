package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type ProductTypeService interface {
	Create(createdByID uint, x *models.ProductType) (*models.ProductType, error)
	ListForUser(userID uint, showAll bool, branchID *uint) ([]models.ProductType, error)
	List(showAll bool) ([]models.ProductType, error)
	// GetByID, Update, and Delete take the caller's branch scopeIDs (nil
	// means unscoped / super admin access) so access can be enforced the
	// same way ListForUser does.
	GetByID(id uint, scopeIDs []uint) (*models.ProductType, error)
	Update(id uint, scopeIDs []uint, x *models.ProductType) (*models.ProductType, error)
	Delete(id uint, scopeIDs []uint) error
	TopUpCredit(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.ProductType, error)
	WithdrawCredit(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.ProductType, error)
}

type productTypeService struct {
	repo repositories.ProductTypeRepository
}

func NewProductTypeService(repo repositories.ProductTypeRepository) ProductTypeService {
	return &productTypeService{repo}
}

func (s *productTypeService) Create(createdByID uint, x *models.ProductType) (*models.ProductType, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, nil)
}

func (s *productTypeService) List(showAll bool) ([]models.ProductType, error) {
	return s.repo.List(showAll)
}

func (s *productTypeService) GetByID(id uint, scopeIDs []uint) (*models.ProductType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product type not found")
		}
		return nil, err
	}
	return x, nil
}

func (s *productTypeService) Update(id uint, scopeIDs []uint, upd *models.ProductType) (*models.ProductType, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("product type not found")
	}

	// Only copy the fields a client is allowed to change. This preserves
	// ID, CreatedByID, CreatedAt, and any other bookkeeping fields that a
	// full struct overwrite would otherwise wipe out.
	x.Name = upd.Name
	x.Code = upd.Code
	x.Description = upd.Description
	x.Icon = upd.Icon
	x.CurrencyTypeID = upd.CurrencyTypeID
	x.BranchID = upd.BranchID
	x.SortOrder = upd.SortOrder
	x.IsActive = upd.IsActive

	if err := s.repo.Update(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *productTypeService) TopUpCredit(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.ProductType, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return nil, errors.New("product type not found")
	}
	return s.repo.TopUpCredit(id, amount, remark, createdByID)
}

func (s *productTypeService) WithdrawCredit(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.ProductType, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return nil, errors.New("product type not found")
	}
	return s.repo.WithdrawCredit(id, amount, remark, createdByID)
}

func (s *productTypeService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("product type not found")
	}
	return s.repo.Delete(id, scopeIDs)
}

func (s *productTypeService) ListForUser(userID uint, showAll bool, branchID *uint) ([]models.ProductType, error) {
	return s.repo.ListForUser(userID, showAll, branchID)
}
