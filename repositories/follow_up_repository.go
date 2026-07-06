package repositories

import (
	"time"

	"gorm.io/gorm"

	followupdto "crm-backend/dto/follow_up"
	"crm-backend/models"
	"crm-backend/utils"
)

type FollowUpRepository interface {
	Create(f *models.ClientFollowUp) error
	FindByID(id uint, scopeIDs []uint) (*models.ClientFollowUp, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter followupdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.ClientFollowUp, int64, error)
}

type followUpRepository struct{ db *gorm.DB }

func NewFollowUpRepository(db *gorm.DB) FollowUpRepository {
	return &followUpRepository{db}
}

func (r *followUpRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("Client").
		Preload("BonusOption").
		Preload("CreatedBy")
}

func (r *followUpRepository) Create(f *models.ClientFollowUp) error {
	return r.db.Create(f).Error
}

// FindByID loads a follow-up by ID. If scopeIDs is non-nil, the record must
// belong to one of those branch IDs or ErrRecordNotFound is returned (so
// callers can't distinguish "doesn't exist" from "not in your scope").
func (r *followUpRepository) FindByID(id uint, scopeIDs []uint) (*models.ClientFollowUp, error) {
	var f models.ClientFollowUp
	q := r.preload(r.db).Where("id = ?", id)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("client_follow_ups.branch_id IN ?", scopeIDs)
	}
	if err := q.First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// Delete removes a follow-up by ID. If scopeIDs is non-nil, the record must
// belong to one of those branch IDs, otherwise nothing is deleted and
// ErrRecordNotFound is returned.
func (r *followUpRepository) Delete(id uint, scopeIDs []uint) error {
	if scopeIDs != nil && len(scopeIDs) == 0 {
		return gorm.ErrRecordNotFound
	}
	q := r.db.Where("id = ?", id)
	if scopeIDs != nil {
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	res := q.Delete(&models.ClientFollowUp{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *followUpRepository) List(filter followupdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.ClientFollowUp, int64, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return []models.ClientFollowUp{}, 0, nil
	}

	q := r.preload(r.db.Model(&models.ClientFollowUp{})).Preload("Branch")
	if !isSA {
		q = q.Where("client_follow_ups.branch_id IN ?", branchIDs)
	}
	if filter.ClientID != nil {
		q = q.Where("client_follow_ups.client_id = ?", *filter.ClientID)
	}
	if filter.BranchID != nil {
		q = q.Where("client_follow_ups.branch_id = ?", *filter.BranchID)
	}
	if filter.CreatedByID != nil {
		q = q.Where("client_follow_ups.created_by_id = ?", *filter.CreatedByID)
	}
	if filter.BonusOptionID != nil {
		q = q.Where("client_follow_ups.bonus_option_id = ?", *filter.BonusOptionID)
	}
	if filter.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", filter.DateFrom); err == nil {
			q = q.Where("follow_up_at >= ?", t)
		}
	}
	if filter.DateTo != "" {
		if t, err := time.Parse("2006-01-02", filter.DateTo); err == nil {
			q = q.Where("follow_up_at <= ?", t.Add(24*time.Hour))
		}
	}

	// Sanitize sort column against an allow-list before it goes into a raw
	// ORDER BY clause — filter.SortBy previously went straight into the
	// query unescaped.
	allowed := map[string]string{
		"follow_up_at": "follow_up_at",
		"created_at":   "created_at",
	}
	q = q.Order(utils.SanitizeSort(filter.SortBy, allowed, "follow_up_at") + " " + utils.SortDir(filter.SortDir))

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.ClientFollowUp
	if err := q.Offset((p.Page - 1) * p.PageSize).Limit(p.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *followUpRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
