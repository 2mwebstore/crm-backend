package repositories

import (
	"gorm.io/gorm"

	"crm-backend/models"
)

// AttendanceFilter bundles List's optional filters.
type AttendanceFilter struct {
	UserID       uint   // 0 = no filter
	BranchID     uint   // 0 = no filter
	DateFrom     string // YYYY-MM-DD, "" = no filter
	DateTo       string // YYYY-MM-DD, "" = no filter
	ActivityOnly bool   // true = only rows checked in via an approved Activity request
}

type AttendanceRepository interface {
	Create(x *models.Attendance) error
	Update(x *models.Attendance) error
	FindByID(id uint) (*models.Attendance, error)
	// FindByUserAndDate returns gorm.ErrRecordNotFound if this user has no
	// attendance row yet for this date — callers should treat that as
	// "hasn't checked in today" rather than a real error.
	FindByUserAndDate(userID uint, date string) (*models.Attendance, error)
	List(filter AttendanceFilter, page, pageSize int) ([]models.Attendance, int64, error)
}

type attendanceRepository struct{ db *gorm.DB }

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db}
}

func (r *attendanceRepository) Create(x *models.Attendance) error {
	return r.db.Create(x).Error
}

func (r *attendanceRepository) Update(x *models.Attendance) error {
	x.User = nil
	x.Branch = nil
	return r.db.Save(x).Error
}

func (r *attendanceRepository) FindByID(id uint) (*models.Attendance, error) {
	var x models.Attendance
	err := r.db.Preload("User").Preload("Branch").First(&x, id).Error
	if err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *attendanceRepository) FindByUserAndDate(userID uint, date string) (*models.Attendance, error) {
	var x models.Attendance
	err := r.db.Preload("User").Preload("Branch").
		Where("user_id = ? AND date = ?", userID, date).
		First(&x).Error
	if err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *attendanceRepository) List(filter AttendanceFilter, page, pageSize int) ([]models.Attendance, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.Attendance{})
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
	if filter.ActivityOnly {
		q = q.Where("check_in_via_outdoor = ?", true)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Attendance
	err := q.Preload("User").Preload("Branch").
		Order("date DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}
