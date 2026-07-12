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

// FindByID, List, and Delete deliberately ignore scopeIDs — CurrencyType
// is a shared global reference list (USD/KHR), not a branch- or
// creator-scoped entity. Filtering by "created_by_id IN scopeIDs" here was
// a copy-paste bug from a branch-scoped lookup repository: it silently
// hid every currency from a user unless THEY personally created it, which
// makes no sense for a shared master list everyone needs to see the same
// way. The scopeIDs parameters stay in the signature so the service/
// controller layer above doesn't need to change, they're just unused here.
func (r *currencyTypeRepository) FindByID(id uint, scopeIDs []uint) (*models.CurrencyType, error) {
	var x models.CurrencyType
	err := r.db.Where("id = ?", id).First(&x).Error
	return &x, err
}

func (r *currencyTypeRepository) List(scopeIDs []uint, showAll bool) ([]models.CurrencyType, error) {
	var items []models.CurrencyType
	q := r.db.Model(&models.CurrencyType{})
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *currencyTypeRepository) Update(x *models.CurrencyType) error {
	return r.db.Save(x).Error
}

func (r *currencyTypeRepository) Delete(id uint, scopeIDs []uint) error {
	return r.db.Where("id = ?", id).Delete(&models.CurrencyType{}).Error
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
