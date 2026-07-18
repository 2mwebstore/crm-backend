package repositories

import (
	"gorm.io/gorm"

	"crm-backend/models"
)

type ActivityRequestFilter struct {
	UserID   uint
	BranchID uint
	DateFrom string
	DateTo   string
}

type ActivityRequestRepository interface {
	Create(x *models.ActivityRequest) error
	FindByID(id uint) (*models.ActivityRequest, error)
	List(filter ActivityRequestFilter, page, pageSize int) ([]models.ActivityRequest, int64, error)
	// HasForDate returns whether userID has an Activity request (always
	// effectively "approved" — see the model) covering date. Used by
	// AttendanceService.CheckIn to decide whether to skip the normal
	// distance/radius requirement for that check-in.
	HasForDate(userID uint, date string) (bool, error)
}

type activityRequestRepository struct{ db *gorm.DB }

func NewActivityRequestRepository(db *gorm.DB) ActivityRequestRepository {
	return &activityRequestRepository{db}
}

func (r *activityRequestRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("User").Preload("Branch")
}

func (r *activityRequestRepository) Create(x *models.ActivityRequest) error {
	return r.db.Create(x).Error
}

func (r *activityRequestRepository) FindByID(id uint) (*models.ActivityRequest, error) {
	var x models.ActivityRequest
	err := r.preload(r.db).First(&x, id).Error
	return &x, err
}

func (r *activityRequestRepository) List(filter ActivityRequestFilter, page, pageSize int) ([]models.ActivityRequest, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.ActivityRequest{})
	if filter.UserID != 0 {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if filter.BranchID != 0 {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.DateFrom != "" {
		q = q.Where("date >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		q = q.Where("date <= ?", filter.DateTo)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.ActivityRequest
	err := r.preload(q).Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *activityRequestRepository) HasForDate(userID uint, date string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ActivityRequest{}).
		Where("user_id = ? AND date = ?", userID, date).
		Count(&count).Error
	return count > 0, err
}
