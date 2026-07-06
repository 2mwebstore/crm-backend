package repositories

import (
	"strings"
	"time"

	"gorm.io/gorm"

	interestingdto "crm-backend/dto/interesting_client"
	"crm-backend/models"
	"crm-backend/utils"
)

type InterestingClientRepository interface {
	Create(ic *models.InterestingClient) error
	FindByID(id uint, scopeIDs []uint) (*models.InterestingClient, error)
	FindByIDUnsafe(id uint) (*models.InterestingClient, error)
	Update(ic *models.InterestingClient) error
	Delete(id uint) error
	List(filter interestingdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.InterestingClient, int64, error)
	// Phone helpers
	DeletePhones(icID uint) error
	CreatePhones(phones []models.InterestingClientPhone) error
	FindPhone(id uint) (*models.InterestingClientPhone, error)
	UpdatePhone(p *models.InterestingClientPhone) error
}

type interestingClientRepository struct{ db *gorm.DB }

func NewInterestingClientRepository(db *gorm.DB) InterestingClientRepository {
	return &interestingClientRepository{db}
}

func (r *interestingClientRepository) preload(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Branch").Preload("ContactSource").Preload("CreatedBy").Preload("Phones")
}

func (r *interestingClientRepository) Create(ic *models.InterestingClient) error {
	return r.db.Create(ic).Error
}
func (r *interestingClientRepository) Update(ic *models.InterestingClient) error {
	return r.db.Save(ic).Error
}
func (r *interestingClientRepository) Delete(id uint) error {
	return r.db.Delete(&models.InterestingClient{}, id).Error
}

func (r *interestingClientRepository) FindByID(id uint, scopeIDs []uint) (*models.InterestingClient, error) {
	return r.FindByIDWithBranch(id, scopeIDs, nil)
}

func (r *interestingClientRepository) FindByIDWithBranch(id uint, scopeIDs []uint, branchIDs []uint) (*models.InterestingClient, error) {
	var ic models.InterestingClient
	q := r.preload(r.db)
	if branchIDs == nil {
		q = q.Where("id = ?", id)
	} else if len(branchIDs) > 0 {
		q = q.Where("id = ? AND branch_id IN ?", id, branchIDs)
	} else if scopeIDs != nil && len(scopeIDs) > 0 {
		q = q.Where("id = ? AND created_by_id IN ?", id, scopeIDs)
	} else {
		q = q.Where("id = ?", id)
	}
	return &ic, q.First(&ic).Error
}

func (r *interestingClientRepository) FindByIDUnsafe(id uint) (*models.InterestingClient, error) {
	var ic models.InterestingClient
	return &ic, r.preload(r.db).First(&ic, id).Error
}

func (r *interestingClientRepository) List(f interestingdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.InterestingClient, int64, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return nil, 0, nil
	}
	q := r.db.Model(&models.InterestingClient{}).
		Preload("Branch").Preload("ContactSource").Preload("CreatedBy").Preload("Phones")

	if !isSA {
		q = q.Where("interesting_clients.branch_id IN ?", branchIDs)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		q = q.Where("LOWER(full_name) LIKE ? OR code LIKE ?", like, like)
	}
	if f.IsActive != nil {
		q = q.Where("is_active = ?", *f.IsActive)
	}
	if f.BranchID != nil {
		q = q.Where("interesting_clients.branch_id = ?", *f.BranchID)
	}
	if f.CreatedByID != nil {
		q = q.Where("interesting_clients.created_by_id = ?", *f.CreatedByID)
	}
	if f.ContactSourceID != nil {
		q = q.Where("contact_source_id = ?", *f.ContactSourceID)
	}
	if f.IsConverted != nil {
		q = q.Where("is_converted = ?", *f.IsConverted)
	}
	if f.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			q = q.Where("interesting_clients.date_joined >= ?", t)
		}
	}
	if f.DateTo != "" {
		if t, err := time.Parse("2006-01-02", f.DateTo); err == nil {
			q = q.Where("interesting_clients.date_joined <= ?", t.Add(23*time.Hour+59*time.Minute+59*time.Second))
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowed := map[string]string{
		"full_name": "full_name", "created_at": "created_at",
		"updated_at": "updated_at", "date_joined": "date_joined",
	}
	q = q.Order(utils.SanitizeSort(f.SortBy, allowed, "created_at") + " " + utils.SortDir(f.SortDir))

	var list []models.InterestingClient
	return list, total, q.Scopes(utils.Paginate(p)).Find(&list).Error
}

func (r *interestingClientRepository) DeletePhones(icID uint) error {
	return r.db.Where("interesting_client_id = ?", icID).Delete(&models.InterestingClientPhone{}).Error
}
func (r *interestingClientRepository) CreatePhones(phones []models.InterestingClientPhone) error {
	if len(phones) == 0 {
		return nil
	}
	return r.db.Create(&phones).Error
}
func (r *interestingClientRepository) FindPhone(id uint) (*models.InterestingClientPhone, error) {
	var p models.InterestingClientPhone
	return &p, r.db.First(&p, id).Error
}
func (r *interestingClientRepository) UpdatePhone(p *models.InterestingClientPhone) error {
	return r.db.Save(p).Error
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *interestingClientRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
