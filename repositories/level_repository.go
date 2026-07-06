package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type LevelRepository interface {
	Create(x *models.Level, createdByID uint) error
	FindByID(id uint, scopeIDs []uint) (*models.Level, error)
	ListForUser(userID uint, showAll bool) ([]models.Level, error)
	List(showAll bool) ([]models.Level, error)
	Update(x *models.Level) error
	Delete(id uint, scopeIDs []uint) error
	ExistsByName(name string, excludeID uint) bool
}

type levelRepository struct{ db *gorm.DB }

func NewLevelRepository(db *gorm.DB) LevelRepository {
	return &levelRepository{db}
}

func (r *levelRepository) Create(x *models.Level, createdByID uint) error {
	x.CreatedByID = createdByID
	return r.db.Create(x).Error
}

// FindByID loads a level by ID. If scopeIDs is non-nil, the record must
// belong to one of those branch IDs or ErrRecordNotFound is returned (so
// callers can't distinguish "doesn't exist" from "not in your scope").
func (r *levelRepository) FindByID(id uint, scopeIDs []uint) (*models.Level, error) {
	var x models.Level
	q := r.db.Preload("Branch").Preload("CreatedBy").Where("id = ?", id)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	if err := q.First(&x).Error; err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *levelRepository) List(showAll bool) ([]models.Level, error) {
	var items []models.Level
	q := r.db.Preload("Branch").Preload("CreatedBy").Model(&models.Level{})
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *levelRepository) Update(x *models.Level) error {
	return r.db.Save(x).Error
}

// Delete removes a level by ID. If scopeIDs is non-nil, the record must
// belong to one of those branch IDs, otherwise nothing is deleted and
// ErrRecordNotFound is returned.
func (r *levelRepository) Delete(id uint, scopeIDs []uint) error {
	if scopeIDs != nil && len(scopeIDs) == 0 {
		return gorm.ErrRecordNotFound
	}
	q := r.db.Where("id = ?", id)
	if scopeIDs != nil {
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	res := q.Delete(&models.Level{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *levelRepository) ExistsByName(name string, excludeID uint) bool {
	var count int64
	q := r.db.Model(&models.Level{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	q.Count(&count)
	return count > 0
}


func (r *levelRepository) ListForUser(userID uint, showAll bool) ([]models.Level, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return []models.Level{}, nil
	}
	var items []models.Level
	q := r.db.Preload("Branch").Preload("CreatedBy").Model(&models.Level{})
	if !isSA {
		q = q.Where("branch_id IN ?", branchIDs)
	}
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *levelRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
