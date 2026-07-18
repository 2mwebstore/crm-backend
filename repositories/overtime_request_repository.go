package repositories

import (
	"gorm.io/gorm"

	"crm-backend/models"
)

type OvertimeRequestFilter struct {
	UserID   uint
	BranchID uint
	Status   models.OvertimeRequestStatus
	DateFrom string
	DateTo   string
}

type OvertimeRequestRepository interface {
	Create(x *models.OvertimeRequest) error
	FindByID(id uint) (*models.OvertimeRequest, error)
	List(filter OvertimeRequestFilter, page, pageSize int) ([]models.OvertimeRequest, int64, error)
	Update(x *models.OvertimeRequest) error
	// HasOverlapping returns whether userID already has a pending or
	// approved Overtime request for date — blocks a second request for a
	// date already covered by an earlier one. excludeID (0 = none) lets a
	// future "edit request" flow re-check without flagging itself.
	HasOverlapping(userID uint, date string, excludeID uint) (bool, error)
}

type overtimeRequestRepository struct{ db *gorm.DB }

func NewOvertimeRequestRepository(db *gorm.DB) OvertimeRequestRepository {
	return &overtimeRequestRepository{db}
}

func (r *overtimeRequestRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("User").Preload("Branch").Preload("ApprovedBy")
}

func (r *overtimeRequestRepository) Create(x *models.OvertimeRequest) error {
	return r.db.Create(x).Error
}

func (r *overtimeRequestRepository) FindByID(id uint) (*models.OvertimeRequest, error) {
	var x models.OvertimeRequest
	err := r.preload(r.db).First(&x, id).Error
	return &x, err
}

func (r *overtimeRequestRepository) List(filter OvertimeRequestFilter, page, pageSize int) ([]models.OvertimeRequest, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.OvertimeRequest{})
	if filter.UserID != 0 {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if filter.BranchID != 0 {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
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
	var list []models.OvertimeRequest
	err := r.preload(q).Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *overtimeRequestRepository) Update(x *models.OvertimeRequest) error {
	x.User = nil
	x.Branch = nil
	x.ApprovedBy = nil
	return r.db.Save(x).Error
}

func (r *overtimeRequestRepository) HasOverlapping(userID uint, date string, excludeID uint) (bool, error) {
	q := r.db.Model(&models.OvertimeRequest{}).
		Where("user_id = ? AND date = ? AND status IN ?",
			userID, date, []models.OvertimeRequestStatus{models.OvertimeRequestPending, models.OvertimeRequestApproved})
	if excludeID != 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}
