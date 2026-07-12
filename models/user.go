package models

import "time"

// User represents a CRM system user.
type User struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"type:varchar(191);not null" json:"name"`
	Email        string `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Avatar       string `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`

	// ── Hierarchy ──────────────────────────────────────────────────────────
	// ParentID tracks who created this user (nil = root/simple user,
	// set = sub-user created by that parent). Kept as a plain nullable FK
	// so callers can check user.ParentID directly instead of resolving it
	// through raw SQL every time.
	ParentID *uint `gorm:"index" json:"parent_id,omitempty"`

	// ── Role & Permissions ─────────────────────────────────────────────────
	RoleID *uint `gorm:"index" json:"role_id,omitempty"`
	Role   *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`

	// IsSuperAdmin bypasses all permission checks.
	IsSuperAdmin bool `gorm:"default:false" json:"is_super_admin"`

	// TokenVersion enforces "one active session per user" — every
	// successful login increments this and issues a JWT carrying the new
	// value; the auth middleware rejects any token whose embedded version
	// doesn't match this current value, so logging in on a new device
	// silently invalidates every other device's existing session.
	TokenVersion int `gorm:"default:0" json:"-"`

	// ── Branches (many-to-many) ─────────────────────────────────────────────
	// Super Admin assigns one or more branches to a user.
	Branches []Branch `gorm:"many2many:user_branches;" json:"branches,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RootID returns the user's own ID. Use UserRepository.GetRootAncestorID
// to walk the parent_id chain to the actual root ancestor.
func (u *User) RootID() uint { return u.ID }

// HasPermission checks if the user's role contains the given permission.
func (u *User) HasPermission(perm string) bool {
	if u.IsSuperAdmin {
		return true
	}
	if u.Role == nil {
		return false
	}
	return u.Role.HasPermission(perm)
}
