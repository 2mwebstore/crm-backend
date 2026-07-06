package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(role *models.Role) error
	FindByID(id uint) (*models.Role, error)
	FindByName(name string, createdByID *uint) (*models.Role, error)
	ListAccessible(callerID uint) ([]models.Role, error)
	ListForAssignment(callerID uint) ([]models.Role, error)
	Update(role *models.Role) error
	Delete(id uint) error
	SetPermissions(role *models.Role, perms []models.Permission) error
	SeedSystemRoles(perms []models.Permission) error
}

type roleRepository struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db}
}

func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) FindByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Preload("CreatedBy").First(&role, id).Error
	return &role, err
}

func (r *roleRepository) FindByName(name string, createdByID *uint) (*models.Role, error) {
	var role models.Role
	q := r.db.Where("name = ?", name)
	if createdByID != nil {
		q = q.Where("created_by_id = ?", *createdByID)
	} else {
		q = q.Where("created_by_id IS NULL")
	}
	return &role, q.First(&role).Error
}

// ListAccessible returns roles visible to the caller.
//
// Rules:
//   - callerID = 0  → SA or SA sub-user: only is_system = 1 roles
//   - callerID > 0  → Simple/Sub User: ONLY roles where created_by_id = callerID
//     (their own created roles, no system roles in the list)
func (r *roleRepository) ListAccessible(callerID uint) ([]models.Role, error) {
	var roles []models.Role
	q := r.db.Preload("Permissions").Order("is_system DESC, name ASC")
	if callerID == 0 {
		// SA / SA sub-user: system roles only
		q = q.Where("is_system = ?", true)
	} else {
		// Simple/Sub User: only own created roles
		q = q.Where("created_by_id = ?", callerID)
	}
	err := q.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

func (r *roleRepository) SetPermissions(role *models.Role, perms []models.Permission) error {
	return r.db.Model(role).Association("Permissions").Replace(perms)
}

// SeedSystemRoles creates the built-in Owner/Manager/Viewer roles if missing.
func (r *roleRepository) SeedSystemRoles(perms []models.Permission) error {
	// Build lookup map: name → Permission
	pm := make(map[string]models.Permission, len(perms))
	for _, p := range perms {
		pm[p.Name] = p
	}

	pick := func(names ...string) []models.Permission {
		out := make([]models.Permission, 0, len(names))
		for _, n := range names {
			if p, ok := pm[n]; ok {
				out = append(out, p)
			}
		}
		return out
	}

	type systemRole struct {
		name  string
		desc  string
		perms []models.Permission
	}

	roles := []systemRole{
		{
			name:  "Owner",
			desc:  "Full access — can do everything in the system",
			perms: perms, // ALL permissions
		},
		{
			name: "Manager",
			desc: "Full client/IC access, can manage sub-users but not roles",
			perms: pick(
				models.PermClientView, models.PermClientCreate, models.PermClientEdit, models.PermClientDelete, models.PermClientExport,
				models.PermICView, models.PermICCreate, models.PermICEdit, models.PermICDelete, models.PermICConvert, models.PermICExport,
				models.PermUserView, models.PermUserCreate, models.PermUserEdit,
				models.PermRoleView,
				models.PermLookupView,
				models.PermExchangeView,
				models.PermReportView,
				models.PermDepositView, models.PermDepositCreate, models.PermDepositEdit,
				models.PermWithdrawalView, models.PermWithdrawalCreate, models.PermWithdrawalEdit,
			),
		},
		{
			name: "Sales",
			desc: "Create and view own clients/ICs, no delete",
			perms: pick(
				models.PermClientView, models.PermClientCreate, models.PermClientEdit, models.PermClientExport,
				models.PermICView, models.PermICCreate, models.PermICEdit, models.PermICConvert, models.PermICExport,
				models.PermLookupView,
				models.PermExchangeView,
			),
		},
		{
			name: "Viewer",
			desc: "Read-only access to clients and ICs",
			perms: pick(
				models.PermClientView,
				models.PermICView,
				models.PermLookupView,
				models.PermExchangeView,
			),
		},
	}

	for _, sr := range roles {
		var existing models.Role
		err := r.db.Where("name = ? AND is_system = ?", sr.name, true).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			// Create new system role
			role := &models.Role{
				Name:        sr.name,
				Description: sr.desc,
				IsSystem:    true,
				Permissions: sr.perms,
			}
			if err2 := r.db.Create(role).Error; err2 != nil {
				return err2
			}
		} else if err == nil {
			// Update existing — always sync permissions to latest definition
			if err2 := r.db.Model(&existing).Association("Permissions").Replace(sr.perms); err2 != nil {
				return err2
			}
		}
	}
	return nil
}

// ListForAssignment returns roles for the role assignment dropdown.
// callerID = 0: system roles only.
// callerID > 0: system roles + own created roles.
func (r *roleRepository) ListForAssignment(callerID uint) ([]models.Role, error) {
	var roles []models.Role
	q := r.db.Preload("Permissions").Order("is_system DESC, name ASC")
	if callerID == 0 {
		q = q.Where("is_system = ?", true)
	} else {
		q = q.Where("is_system = ? OR created_by_id = ?", true, callerID)
	}
	err := q.Find(&roles).Error
	return roles, err
}
