package repositories

import (
	"time"

	"gorm.io/gorm"

	"crm-backend/models"
)

// AuditLogFilter bundles List's optional filters.
type AuditLogFilter struct {
	UserID   uint   // 0 = no filter
	BranchID uint   // 0 = no filter
	Method   string // "" = no filter
	DateFrom string // YYYY-MM-DD, "" = no filter
	DateTo   string // YYYY-MM-DD, "" = no filter
	Search   string // matches against Path, "" = no filter
}

type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	List(filter AuditLogFilter, page, pageSize int) ([]models.AuditLog, int64, error)
	// DeleteBefore removes every entry with CreatedAt strictly before
	// cutoff, returning how many rows were actually deleted.
	DeleteBefore(cutoff time.Time) (int64, error)
}

type auditLogRepository struct{ db *gorm.DB }

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db}
}

func (r *auditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) List(filter AuditLogFilter, page, pageSize int) ([]models.AuditLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.AuditLog{})
	if filter.UserID != 0 {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if filter.BranchID != 0 {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.Method != "" {
		q = q.Where("method = ?", filter.Method)
	}
	if filter.Search != "" {
		q = q.Where("path LIKE ?", "%"+filter.Search+"%")
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
	var list []models.AuditLog
	err := q.Preload("User").Preload("Branch").
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *auditLogRepository) DeleteBefore(cutoff time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", cutoff).Delete(&models.AuditLog{})
	return result.RowsAffected, result.Error
}
