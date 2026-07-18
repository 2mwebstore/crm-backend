package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type LeaveTypeRepository interface {
	Create(x *models.LeaveType, createdByID uint) error
	FindByID(id uint, callerID uint) (*models.LeaveType, error)
	List(callerID uint, showAll bool, branchID *uint) ([]models.LeaveType, error)
	Update(x *models.LeaveType) error
	Delete(id uint) error
}

type leaveTypeRepository struct{ db *gorm.DB }

func NewLeaveTypeRepository(db *gorm.DB) LeaveTypeRepository {
	return &leaveTypeRepository{db}
}

func (r *leaveTypeRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("Branch").Preload("CreatedBy")
}

func (r *leaveTypeRepository) Create(x *models.LeaveType, createdByID uint) error {
	x.CreatedByID = createdByID
	return r.db.Create(x).Error
}

// FindByID is scope-checked the same way List is — a non-SA caller can't
// look up a leave type that belongs to a branch they're not assigned to,
// even by guessing its ID directly. A global (BranchID=nil) leave type is
// always visible to everyone regardless of scope.
func (r *leaveTypeRepository) FindByID(id uint, callerID uint) (*models.LeaveType, error) {
	branchIDs, isSA := r.resolveUserBranches(callerID)
	var x models.LeaveType
	q := r.preload(r.db).Where("id = ?", id)
	if !isSA {
		q = q.Where("branch_id IS NULL OR branch_id IN ?", branchIDs)
	}
	err := q.First(&x).Error
	return &x, err
}

func (r *leaveTypeRepository) List(callerID uint, showAll bool, branchID *uint) ([]models.LeaveType, error) {
	branchIDs, isSA := r.resolveUserBranches(callerID)
	if !isSA && len(branchIDs) == 0 {
		// Non-SA with no branches assigned still sees global leave types —
		// unlike ProductType, a leave type with no branch is meant for
		// everyone, so this isn't a "see nothing" case.
		branchIDs = []uint{}
	}
	var items []models.LeaveType
	q := r.preload(r.db).Model(&models.LeaveType{})
	if !isSA {
		q = q.Where("branch_id IS NULL OR branch_id IN ?", branchIDs)
	}
	if branchID != nil {
		// Intersected with the scope filter above — narrows to "global OR
		// this specific branch", still within whatever the caller can see.
		q = q.Where("branch_id IS NULL OR branch_id = ?", *branchID)
	}
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *leaveTypeRepository) Update(x *models.LeaveType) error {
	x.Branch = nil
	x.CreatedBy = nil
	return r.db.Save(x).Error
}

func (r *leaveTypeRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&models.LeaveType{}).Error
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything), []uint = the caller's
// own directly-assigned branches (may be empty = no branch-specific
// access, but global leave types are still visible — see List above).
func (r *leaveTypeRepository) resolveUserBranches(userID uint) ([]uint, bool) {
	var row struct {
		IsSuperAdmin bool
	}
	if err := r.db.Raw("SELECT is_super_admin FROM users WHERE id = ?", userID).Scan(&row).Error; err != nil {
		return []uint{}, false
	}
	if row.IsSuperAdmin {
		return nil, true
	}
	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", userID).Scan(&branchIDs)
	return branchIDs, false
}
