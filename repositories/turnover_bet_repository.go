package repositories

import (
	"time"

	"gorm.io/gorm"

	turnoverbetdto "crm-backend/dto/turnover_bet"
	"crm-backend/models"
	"crm-backend/utils"
)

type TurnoverBetRepository interface {
	Create(t *models.TurnoverBet) error
	FindByID(id uint, scopeIDs []uint) (*models.TurnoverBet, error)
	FindByIDUnsafe(id uint) (*models.TurnoverBet, error)
	Update(t *models.TurnoverBet) error
	Delete(id uint) error
	List(filter turnoverbetdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.TurnoverBet, int64, error)
}

type turnoverBetRepository struct{ db *gorm.DB }

func NewTurnoverBetRepository(db *gorm.DB) TurnoverBetRepository {
	return &turnoverBetRepository{db}
}

func (r *turnoverBetRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("ProductType").
		Preload("Branch").
		Preload("CreatedBy").
		Preload("ApprovedBy")
}

func (r *turnoverBetRepository) Create(t *models.TurnoverBet) error {
	return r.db.Create(t).Error
}

func (r *turnoverBetRepository) FindByIDUnsafe(id uint) (*models.TurnoverBet, error) {
	var t models.TurnoverBet
	return &t, r.preload(r.db).First(&t, id).Error
}

func (r *turnoverBetRepository) FindByID(id uint, scopeIDs []uint) (*models.TurnoverBet, error) {
	var t models.TurnoverBet
	q := r.preload(r.db)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("turnover_bets.branch_id IN ?", scopeIDs)
	}
	return &t, q.First(&t, id).Error
}

func (r *turnoverBetRepository) Update(t *models.TurnoverBet) error {
	return r.db.Save(t).Error
}

func (r *turnoverBetRepository) Delete(id uint) error {
	return r.db.Delete(&models.TurnoverBet{}, id).Error
}

func (r *turnoverBetRepository) List(f turnoverbetdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.TurnoverBet, int64, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return nil, 0, nil
	}
	q := r.preload(r.db.Model(&models.TurnoverBet{})).Preload("Branch")
	if !isSA {
		q = q.Where("turnover_bets.branch_id IN ?", branchIDs)
	}
	if f.ProductTypeID != nil {
		q = q.Where("product_type_id = ?", *f.ProductTypeID)
	}
	if f.BranchID != nil {
		q = q.Where("turnover_bets.branch_id = ?", *f.BranchID)
	}
	if f.CreatedByID != nil {
		q = q.Where("turnover_bets.created_by_id = ?", *f.CreatedByID)
	}
	if f.ApprovedByID != nil {
		q = q.Where("turnover_bets.approved_by_id = ?", *f.ApprovedByID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Currency != "" {
		q = q.Where("currency = ?", f.Currency)
	}
	if f.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			q = q.Where("date >= ?", t)
		}
	}
	if f.DateTo != "" {
		if t, err := time.Parse("2006-01-02", f.DateTo); err == nil {
			q = q.Where("date <= ?", t.Add(24*time.Hour))
		}
	}

	sortBy := "date"
	sortDir := "desc"
	if f.SortBy != "" {
		sortBy = f.SortBy
	}
	if f.SortDir != "" {
		sortDir = f.SortDir
	}
	q = q.Order(sortBy + " " + sortDir)

	var total int64
	q.Count(&total)
	var items []models.TurnoverBet
	err := q.Offset((p.Page - 1) * p.PageSize).Limit(p.PageSize).Find(&items).Error
	return items, total, err
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *turnoverBetRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
