package models

import "time"

// Role is a named collection of permissions.
// Roles are created per-user-tree — a user can only assign roles
// they themselves own (created_by_id).
type Role struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"type:varchar(191);not null" json:"name"`
	Description string       `gorm:"type:varchar(500)" json:"description,omitempty"`
	IsSystem    bool         `gorm:"default:false" json:"is_system"` // true = seeded system roles, cannot be deleted
	CreatedByID *uint        `gorm:"index" json:"created_by_id,omitempty"`
	CreatedBy   *User        `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// HasPermission returns true if the role contains the given permission name.
func (r *Role) HasPermission(perm string) bool {
	for _, p := range r.Permissions {
		if p.Name == perm {
			return true
		}
	}
	return false
}

// PermissionNames returns just the permission name strings from the role.
func (r *Role) PermissionNames() []string {
	names := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		names = append(names, p.Name)
	}
	return names
}
