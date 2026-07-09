package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type BalanceTransactionRepository interface {
	// Create inserts a ledger row. Called from within the same DB
	// transaction that mutates the actual balance column, so the ledger
	// entry and the balance change commit or roll back together.
	Create(tx *gorm.DB, entry *models.BalanceTransaction) error
	// ListByEntity returns transactions for one entity, optionally filtered
	// by txType ("topup"/"withdrawal" — pass "" for no filter).
	ListByEntity(entityType models.BalanceEntityType, entityID uint, txType models.BalanceTxType, page, pageSize int) ([]models.BalanceTransaction, int64, error)
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

func (r *balanceTransactionRepository) ListByEntity(entityType models.BalanceEntityType, entityID uint, txType models.BalanceTxType, page, pageSize int) ([]models.BalanceTransaction, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.BalanceTransaction{}).Where("entity_type = ? AND entity_id = ?", entityType, entityID)
	if txType != "" {
		q = q.Where("type = ?", txType)
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
