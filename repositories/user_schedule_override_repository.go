package repositories

import (
	"gorm.io/gorm"

	"crm-backend/models"
)

// UserScheduleOverrideFilter bundles List's optional filters.
type UserScheduleOverrideFilter struct {
	UserID   uint   // 0 = no filter
	BranchID uint   // 0 = no filter — matches users assigned to this branch
	Search   string // "" = no filter — matches the user's name or email
}

type UserScheduleOverrideRepository interface {
	Create(x *models.UserScheduleOverride) error
	FindByID(id uint) (*models.UserScheduleOverride, error)
	// List returns overrides across ALL users by default, narrowed by
	// whichever filters are set — branch (via the user's assigned
	// branches) and/or a name/email search.
	List(filter UserScheduleOverrideFilter, page, pageSize int) ([]models.UserScheduleOverride, int64, error)
	ListForUser(userID uint) ([]models.UserScheduleOverride, error)
	Update(x *models.UserScheduleOverride) error
	Delete(id uint) error
	// FindActiveForDate returns the override (if any) covering date for
	// userID. If more than one somehow overlaps the same date, the most
	// recently created one wins. Returns gorm.ErrRecordNotFound if none
	// applies — callers should treat that as "no override, use the
	// user's own default shift times", not an error.
	FindActiveForDate(userID uint, date string) (*models.UserScheduleOverride, error)
}

type userScheduleOverrideRepository struct{ db *gorm.DB }

func NewUserScheduleOverrideRepository(db *gorm.DB) UserScheduleOverrideRepository {
	return &userScheduleOverrideRepository{db}
}

func (r *userScheduleOverrideRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("User").Preload("User.Branches").Preload("CreatedBy")
}

func (r *userScheduleOverrideRepository) Create(x *models.UserScheduleOverride) error {
	return r.db.Create(x).Error
}

func (r *userScheduleOverrideRepository) FindByID(id uint) (*models.UserScheduleOverride, error) {
	var x models.UserScheduleOverride
	err := r.preload(r.db).First(&x, id).Error
	return &x, err
}

// List joins to users (always, for name/email/search — and so results
// come back with a stable, name-sortable order) and to user_branches only
// when a branch filter is actually requested.
func (r *userScheduleOverrideRepository) List(filter UserScheduleOverrideFilter, page, pageSize int) ([]models.UserScheduleOverride, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.db.Model(&models.UserScheduleOverride{}).
		Joins("JOIN users ON users.id = user_schedule_overrides.user_id")
	if filter.BranchID != 0 {
		q = q.Joins("JOIN user_branches ub ON ub.user_id = users.id AND ub.branch_id = ?", filter.BranchID)
	}
	if filter.UserID != 0 {
		q = q.Where("user_schedule_overrides.user_id = ?", filter.UserID)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("users.name LIKE ? OR users.email LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.UserScheduleOverride
	err := r.preload(q).
		Order("user_schedule_overrides.date_from DESC, user_schedule_overrides.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *userScheduleOverrideRepository) ListForUser(userID uint) ([]models.UserScheduleOverride, error) {
	var list []models.UserScheduleOverride
	err := r.preload(r.db).
		Where("user_id = ?", userID).
		Order("date_from DESC").
		Find(&list).Error
	return list, err
}

func (r *userScheduleOverrideRepository) Update(x *models.UserScheduleOverride) error {
	x.User = nil
	x.CreatedBy = nil
	return r.db.Save(x).Error
}

func (r *userScheduleOverrideRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&models.UserScheduleOverride{}).Error
}

func (r *userScheduleOverrideRepository) FindActiveForDate(userID uint, date string) (*models.UserScheduleOverride, error) {
	var x models.UserScheduleOverride
	err := r.db.
		Where("user_id = ? AND date_from <= ? AND date_to >= ?", userID, date, date).
		Order("created_at DESC").
		First(&x).Error
	return &x, err
}
