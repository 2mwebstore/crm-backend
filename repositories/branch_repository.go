package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type BranchRepository interface {
	Create(b *models.Branch) error
	FindByID(id uint) (*models.Branch, error)
	List(showAll bool) ([]models.Branch, error)
	ListForUser(userID uint, showAll bool) ([]models.Branch, error)
	Update(b *models.Branch) error
	Delete(id uint) error
}

type branchRepository struct{ db *gorm.DB }

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &branchRepository{db}
}

func (r *branchRepository) Create(b *models.Branch) error {
	return r.db.Create(b).Error
}

func (r *branchRepository) FindByID(id uint) (*models.Branch, error) {
	var b models.Branch
	return &b, r.db.Preload("CreatedBy").First(&b, id).Error
}

func (r *branchRepository) List(showAll bool) ([]models.Branch, error) {
	var items []models.Branch
	q := r.db.Preload("CreatedBy")
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *branchRepository) Update(b *models.Branch) error {
	return r.db.Save(b).Error
}

func (r *branchRepository) Delete(id uint) error {
	return r.db.Delete(&models.Branch{}, id).Error
}

// ListForUser returns branches scoped to a user:
//   - Super Admin or SA sub-user: all branches
//   - Simple User: branches assigned via user_branches (root ancestor's assignments)
func (r *branchRepository) ListForUser(userID uint, showAll bool) ([]models.Branch, error) {
	// Walk up to root ancestor
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
			// SA or SA sub → return all
			return r.List(showAll)
		}
		if parent.ParentID == nil {
			break // rootID is the simple user root
		}
		rootID = *parent.ParentID
	}

	// Get branch IDs assigned to root
	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", rootID).Scan(&branchIDs)

	if len(branchIDs) == 0 {
		return []models.Branch{}, nil
	}

	var items []models.Branch
	q := r.db.Preload("CreatedBy").Where("id IN ?", branchIDs)
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("name ASC").Find(&items).Error
	return items, err
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *branchRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
