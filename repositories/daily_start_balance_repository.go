package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type DailyStartBalanceRepository interface {
	// CreateWithDetails saves the opening snapshot header row and its
	// "open" phase detail rows together, in a single DB transaction —
	// both commit or neither does.
	CreateWithDetails(x *models.DailyStartBalance, details []models.DailyStartBalanceDetail) error
	// UpdateCloseWithDetails saves the closing totals onto an existing
	// snapshot row and inserts its "close" phase detail rows, in a single
	// DB transaction.
	UpdateCloseWithDetails(x *models.DailyStartBalance, details []models.DailyStartBalanceDetail) error
	FindOpenByBranch(branchID uint) (*models.DailyStartBalance, error)
	ListByBranch(branchID uint, page, pageSize int) ([]models.DailyStartBalance, int64, error)
}

type dailyStartBalanceRepository struct{ db *gorm.DB }

func NewDailyStartBalanceRepository(db *gorm.DB) DailyStartBalanceRepository {
	return &dailyStartBalanceRepository{db}
}

func (r *dailyStartBalanceRepository) CreateWithDetails(x *models.DailyStartBalance, details []models.DailyStartBalanceDetail) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(x).Error; err != nil {
			return err
		}
		if len(details) == 0 {
			return nil
		}
		for i := range details {
			details[i].DailyStartBalanceID = x.ID
		}
		return tx.Create(&details).Error
	})
}

func (r *dailyStartBalanceRepository) UpdateCloseWithDetails(x *models.DailyStartBalance, details []models.DailyStartBalanceDetail) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DailyStartBalance{}).Where("id = ?", x.ID).Updates(map[string]interface{}{
			"close_cash_usd":   x.CloseCashUSD,
			"close_cash_khr":   x.CloseCashKHR,
			"close_credit_usd": x.CloseCreditUSD,
			"close_credit_khr": x.CloseCreditKHR,
			"closed_by_id":     x.ClosedByID,
			"closed_at":        x.ClosedAt,
		}).Error; err != nil {
			return err
		}
		if len(details) == 0 {
			return nil
		}
		for i := range details {
			details[i].DailyStartBalanceID = x.ID
		}
		return tx.Create(&details).Error
	})
}

// FindOpenByBranch returns the currently-open shift (ClosedAt IS NULL) for
// this branch, or gorm.ErrRecordNotFound if none is open right now —
// callers should treat that as "no active shift" rather than a real error.
// If somehow more than one open row exists for a branch (shouldn't happen
// given Start/Close enforce the one-open-at-a-time rule), returns the most
// recently opened one.
func (r *dailyStartBalanceRepository) FindOpenByBranch(branchID uint) (*models.DailyStartBalance, error) {
	var x models.DailyStartBalance
	err := r.db.Preload("CreatedBy").Preload("ClosedBy").Preload("Details").
		Where("branch_id = ? AND closed_at IS NULL", branchID).
		Order("created_at DESC").
		First(&x).Error
	if err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *dailyStartBalanceRepository) ListByBranch(branchID uint, page, pageSize int) ([]models.DailyStartBalance, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.DailyStartBalance{}).Where("branch_id = ?", branchID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.DailyStartBalance
	err := q.Preload("CreatedBy").Preload("ClosedBy").Preload("Details").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}
