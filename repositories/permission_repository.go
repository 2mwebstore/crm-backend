package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	FindAll() ([]models.Permission, error)
	FindByIDs(ids []uint) ([]models.Permission, error)
	FindByNames(names []string) ([]models.Permission, error)
	FindByID(id uint) (*models.Permission, error)
	FindGrouped() (map[string][]models.Permission, error)
	Seed(perms []models.Permission) error
}

type permissionRepository struct{ db *gorm.DB }

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db}
}

func (r *permissionRepository) FindAll() ([]models.Permission, error) {
	var list []models.Permission
	err := r.db.Order("`group` ASC, name ASC").Find(&list).Error
	return list, err
}

func (r *permissionRepository) FindByIDs(ids []uint) ([]models.Permission, error) {
	var list []models.Permission
	err := r.db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *permissionRepository) FindByNames(names []string) ([]models.Permission, error) {
	var list []models.Permission
	err := r.db.Where("name IN ?", names).Find(&list).Error
	return list, err
}

func (r *permissionRepository) FindByID(id uint) (*models.Permission, error) {
	var p models.Permission
	return &p, r.db.First(&p, id).Error
}

func (r *permissionRepository) FindGrouped() (map[string][]models.Permission, error) {
	all, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]models.Permission)
	for _, p := range all {
		grouped[p.Group] = append(grouped[p.Group], p)
	}
	return grouped, nil
}

func (r *permissionRepository) Seed(perms []models.Permission) error {
	for _, p := range perms {
		var existing models.Permission
		err := r.db.Where("name = ?", p.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err2 := r.db.Create(&p).Error; err2 != nil {
				return err2
			}
		}
	}
	return nil
}
