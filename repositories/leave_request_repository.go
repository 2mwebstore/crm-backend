package repositories

import (
	"gorm.io/gorm"

	"crm-backend/models"
)

type LeaveRequestFilter struct {
	UserID   uint
	BranchID uint
	Status   models.LeaveRequestStatus
	DateFrom string
	DateTo   string
}

type LeaveRequestRepository interface {
	Create(x *models.LeaveRequest) error
	FindByID(id uint) (*models.LeaveRequest, error)
	List(filter LeaveRequestFilter, page, pageSize int) ([]models.LeaveRequest, int64, error)
	Update(x *models.LeaveRequest) error
	// HasOverlapping returns whether userID already has a pending or
	// approved Leave request overlapping [dateFrom, dateTo] — blocks a
	// second request for a date already covered by an earlier one.
	HasOverlapping(userID uint, dateFrom, dateTo string, excludeID uint) (bool, error)
	// SumDaysInPeriod returns how many days of leaveTypeID this user has
	// already used (pending or approved) within [periodFrom, periodTo],
	// excluding excludeID (0 = none). A Half Day counts as 0.5.
	SumDaysInPeriod(userID, leaveTypeID uint, periodFrom, periodTo string, excludeID uint) (float64, error)
}

type leaveRequestRepository struct{ db *gorm.DB }

func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return &leaveRequestRepository{db}
}

func (r *leaveRequestRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("User").Preload("Branch").Preload("LeaveType").Preload("ApprovedBy")
}

func (r *leaveRequestRepository) Create(x *models.LeaveRequest) error {
	return r.db.Create(x).Error
}

func (r *leaveRequestRepository) FindByID(id uint) (*models.LeaveRequest, error) {
	var x models.LeaveRequest
	err := r.preload(r.db).First(&x, id).Error
	return &x, err
}

func (r *leaveRequestRepository) List(filter LeaveRequestFilter, page, pageSize int) ([]models.LeaveRequest, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.LeaveRequest{})
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
		q = q.Where("date_to >= ?", filter.DateFrom)
	}
	if filter.DateTo != "" {
		q = q.Where("date_from <= ?", filter.DateTo)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.LeaveRequest
	err := r.preload(q).Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *leaveRequestRepository) Update(x *models.LeaveRequest) error {
	x.User = nil
	x.Branch = nil
	x.LeaveType = nil
	x.ApprovedBy = nil
	return r.db.Save(x).Error
}

func (r *leaveRequestRepository) HasOverlapping(userID uint, dateFrom, dateTo string, excludeID uint) (bool, error) {
	q := r.db.Model(&models.LeaveRequest{}).
		Where("user_id = ? AND status IN ? AND date_from <= ? AND date_to >= ?",
			userID, []models.LeaveRequestStatus{models.LeaveRequestPending, models.LeaveRequestApproved}, dateTo, dateFrom)
	if excludeID != 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *leaveRequestRepository) SumDaysInPeriod(userID, leaveTypeID uint, periodFrom, periodTo string, excludeID uint) (float64, error) {
	q := r.db.Model(&models.LeaveRequest{}).
		Where("user_id = ? AND leave_type_id = ? AND status IN ? AND date_from <= ? AND date_to >= ?",
			userID, leaveTypeID, []models.LeaveRequestStatus{models.LeaveRequestPending, models.LeaveRequestApproved},
			periodTo, periodFrom)
	if excludeID != 0 {
		q = q.Where("id != ?", excludeID)
	}
	var total float64
	if err := q.Select("COALESCE(SUM(duration), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
