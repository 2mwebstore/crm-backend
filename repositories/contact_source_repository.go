package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type ContactSourceRepository interface {
	Create(x *models.ContactSource, createdByID uint) error
	FindByID(id uint, scopeIDs []uint) (*models.ContactSource, error)
	List(scopeIDs []uint, showAll bool) ([]models.ContactSource, error)
	// ListForUser scopes by branch assignment (mirrors branch_repository.go
	// / bank_type_repository.go's ListForUser flow): Super Admin / SA
	// sub-user sees everything, a Simple User (or their sub-users) sees
	// contact sources whose branch_id is one of the root ancestor's
	// assigned branches, and a user with no branches assigned sees nothing.
	ListForUser(userID uint, showAll bool) ([]models.ContactSource, error)
	Update(x *models.ContactSource) error
	Delete(id uint, scopeIDs []uint) error
	ExistsByName(name string, excludeID uint) bool
}

type contactSourceRepository struct{ db *gorm.DB }

func NewContactSourceRepository(db *gorm.DB) ContactSourceRepository {
	return &contactSourceRepository{db}
}

func (r *contactSourceRepository) Create(x *models.ContactSource, createdByID uint) error {
	x.CreatedByID = createdByID
	return r.db.Create(x).Error
}

// FindByID loads a contact source by ID. When scopeIDs is non-empty, the
// record must have been created by one of those IDs (scopeIDs comes from
// userSvc.GetLookupScope: nil/empty means no restriction, and the sentinel
// []uint{0} means no access).
func (r *contactSourceRepository) FindByID(id uint, scopeIDs []uint) (*models.ContactSource, error) {
	var x models.ContactSource
	q := r.db.Preload("Branch").Preload("CreatedBy").Where("id = ?", id)
	if len(scopeIDs) > 0 {
		q = q.Where("created_by_id IN ?", scopeIDs)
	}
	return &x, q.First(&x).Error
}

func (r *contactSourceRepository) List(scopeIDs []uint, showAll bool) ([]models.ContactSource, error) {
	var items []models.ContactSource
	q := r.db.Preload("Branch").Preload("CreatedBy").Model(&models.ContactSource{})
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	if len(scopeIDs) > 0 {
		q = q.Where("created_by_id IN ?", scopeIDs)
	}
	err := q.Order("name ASC").Find(&items).Error
	return items, err
}

// ListForUser mirrors branch_repository.go / bank_type_repository.go's
// ListForUser: walk up to the root ancestor, let Super Admin (or their
// sub-users) see every contact source, otherwise filter to contact sources
// whose branch_id is one of the root's assigned branches.
func (r *contactSourceRepository) ListForUser(userID uint, showAll bool) ([]models.ContactSource, error) {
	rootID := userID
	for {
		var parent struct {
			ParentID     *uint
			IsSuperAdmin bool
		}
		if err := r.db.Raw("SELECT parent_id, is_super_admin FROM users WHERE id = ?", rootID).Scan(&parent).Error; err != nil {
			break
		}
		if parent.IsSuperAdmin {
			return r.List(nil, showAll)
		}
		if parent.ParentID == nil {
			break // rootID is the simple user root
		}
		rootID = *parent.ParentID
	}

	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", rootID).Scan(&branchIDs)

	if len(branchIDs) == 0 {
		return []models.ContactSource{}, nil
	}

	var items []models.ContactSource
	q := r.db.Preload("Branch").Preload("CreatedBy").Where("branch_id IN ?", branchIDs)
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *contactSourceRepository) Update(x *models.ContactSource) error {
	return r.db.Save(x).Error
}

func (r *contactSourceRepository) Delete(id uint, scopeIDs []uint) error {
	q := r.db.Where("id = ?", id)
	if len(scopeIDs) > 0 {
		q = q.Where("created_by_id IN ?", scopeIDs)
	}
	res := q.Delete(&models.ContactSource{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *contactSourceRepository) ExistsByName(name string, excludeID uint) bool {
	var count int64
	q := r.db.Model(&models.ContactSource{}).Where("LOWER(name) = LOWER(?)", name)
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
func (r *contactSourceRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
