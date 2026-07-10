package repositories

import (
	"time"

	"crm-backend/models"

	"gorm.io/gorm"
)

// BalanceTransactionFilter bundles ListByEntity's optional filters — grown
// past the point where separate positional params stayed readable.
type BalanceTransactionFilter struct {
	Type        models.BalanceTxType   // "" = no filter
	Source      models.BalanceTxSource // "" = no filter
	DateFrom    string                 // YYYY-MM-DD, "" = no filter
	DateTo      string                 // YYYY-MM-DD, "" = no filter
	CreatedByID uint                   // 0 = no filter
}

type BalanceTransactionRepository interface {
	// Create inserts a ledger row. Called from within the same DB
	// transaction that mutates the actual balance column, so the ledger
	// entry and the balance change commit or roll back together.
	Create(tx *gorm.DB, entry *models.BalanceTransaction) error
	// ListByEntity returns transactions for one entity, filtered by
	// whichever fields of filter are set.
	ListByEntity(entityType models.BalanceEntityType, entityID uint, filter BalanceTransactionFilter, page, pageSize int) ([]models.BalanceTransaction, int64, error)
	// ListByEntitiesInRange returns every transaction for a SET of entities
	// of one type (e.g. every CompanyBank in a branch) whose CreatedAt
	// falls within [from, to]. Pass to=nil for "up to now" (an entity set
	// still active/open). Used to show exactly which top-ups/withdrawals
	// happened during a given Daily Start Balance shift.
	ListByEntitiesInRange(entityType models.BalanceEntityType, entityIDs []uint, from time.Time, to *time.Time) ([]models.BalanceTransaction, error)
}

type balanceTransactionRepository struct{ db *gorm.DB }

func NewBalanceTransactionRepository(db *gorm.DB) BalanceTransactionRepository {
	return &balanceTransactionRepository{db}
}

func (r *balanceTransactionRepository) Create(tx *gorm.DB, entry *models.BalanceTransaction) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(entry).Error
}

func (r *balanceTransactionRepository) ListByEntity(entityType models.BalanceEntityType, entityID uint, filter BalanceTransactionFilter, page, pageSize int) ([]models.BalanceTransaction, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.BalanceTransaction{}).Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.Source != "" {
		q = q.Where("source = ?", filter.Source)
	}
	if filter.CreatedByID != 0 {
		q = q.Where("created_by_id = ?", filter.CreatedByID)
	}
	if filter.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", filter.DateFrom); err == nil {
			q = q.Where("created_at >= ?", t)
		}
	}
	if filter.DateTo != "" {
		if t, err := time.Parse("2006-01-02", filter.DateTo); err == nil {
			q = q.Where("created_at <= ?", t.Add(23*time.Hour+59*time.Minute+59*time.Second))
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.BalanceTransaction
	err := q.Preload("CreatedBy").
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *balanceTransactionRepository) ListByEntitiesInRange(entityType models.BalanceEntityType, entityIDs []uint, from time.Time, to *time.Time) ([]models.BalanceTransaction, error) {
	if len(entityIDs) == 0 {
		return []models.BalanceTransaction{}, nil
	}
	q := r.db.Model(&models.BalanceTransaction{}).
		Where("entity_type = ? AND entity_id IN ? AND created_at >= ?", entityType, entityIDs, from)
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}
	var list []models.BalanceTransaction
	err := q.Preload("CreatedBy").Order("created_at ASC, id ASC").Find(&list).Error
	return list, err
}
