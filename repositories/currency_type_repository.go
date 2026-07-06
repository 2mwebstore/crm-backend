package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type CurrencyTypeRepository interface {
	Create(x *models.CurrencyType, createdByID uint) error
	FindByID(id uint, scopeIDs []uint) (*models.CurrencyType, error)
	List(scopeIDs []uint, showAll bool) ([]models.CurrencyType, error)
	Update(x *models.CurrencyType) error
	Delete(id uint, scopeIDs []uint) error
}

type currencyTypeRepository struct{ db *gorm.DB }

func NewCurrencyTypeRepository(db *gorm.DB) CurrencyTypeRepository {
	return &currencyTypeRepository{db}
}

func (r *currencyTypeRepository) Create(x *models.CurrencyType, createdByID uint) error {
	x.CreatedByID = createdByID
	return r.db.Create(x).Error
}

func (r *currencyTypeRepository) FindByID(id uint, scopeIDs []uint) (*models.CurrencyType, error) {
	var x models.CurrencyType
	q := r.db.Where("id = ?", id)
	if len(scopeIDs) > 0 {
		q = q.Where("created_by_id IN ?", scopeIDs)
	}
	return &x, q.First(&x).Error
}

func (r *currencyTypeRepository) List(scopeIDs []uint, showAll bool) ([]models.CurrencyType, error) {
	var items []models.CurrencyType
	q := r.db.Model(&models.CurrencyType{})
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	if len(scopeIDs) > 0 {
		q = q.Where("created_by_id IN ?", scopeIDs)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *currencyTypeRepository) Update(x *models.CurrencyType) error {
	return r.db.Save(x).Error
}

func (r *currencyTypeRepository) Delete(id uint, scopeIDs []uint) error {
	q := r.db.Where("id = ?", id)
	if len(scopeIDs) > 0 {
		q = q.Where("created_by_id IN ?", scopeIDs)
	}
	return q.Delete(&models.CurrencyType{}).Error
}

func (r *currencyTypeRepository) ExistsByName(name string, excludeID uint) bool {
	var count int64
	q := r.db.Model(&models.CurrencyType{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	q.Count(&count)
	return count > 0
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *currencyTypeRepository) resolveUserBranches(userID uint) ([]uint, bool) {
	var row struct {
		IsSuperAdmin bool
	}
	if err := r.db.Raw("SELECT is_super_admin FROM users WHERE id = ?", userID).Scan(&row).Error; err != nil {
		return []uint{}, false
	}
	if row.IsSuperAdmin { return nil, true }
	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", userID).Scan(&branchIDs)
	return branchIDs, false
}
