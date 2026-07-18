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

	// ShiftCheckInTime/ShiftCheckOutTime configure attendance timeliness
	// checking — each "HH:MM" (24-hour) time is compared against the
	// matching check-in/check-out action to label it early/good/late (see
	// AttendanceService.CheckIn/CheckOut). Both optional; a user with no
	// shift time set for that side gets no timeliness label at all
	// (nothing to compare against), not a default "late".
	ShiftCheckInTime  *string `gorm:"type:varchar(5)" json:"shift_check_in_time,omitempty"`
	ShiftCheckOutTime *string `gorm:"type:varchar(5)" json:"shift_check_out_time,omitempty"`

	// ShiftType distinguishes a Normal Day schedule from a Cross Day
	// (Night Shift) one, where check-in and check-out naturally fall on
	// two different calendar dates (e.g. in at 22:00, out at 06:00 the
	// next day). AttendanceService and the Leave/Overtime duplicate-date
	// checks both branch on this — see ShiftTypeCrossDay's own comment.
	ShiftType string `gorm:"type:varchar(20);not null;default:normal" json:"shift_type"`

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

const (
	ShiftTypeNormal = "normal"
	// ShiftTypeCrossDay marks a Night Shift user whose check-in and
	// check-out naturally fall on two different calendar dates.
	// AttendanceService.CheckOut looks for an open check-in from
	// YESTERDAY (not just today) for these users, and Leave/Overtime
	// skip their normal "one request per date" duplicate check, since a
	// rigid same-calendar-date rule doesn't fit a schedule that's
	// designed to straddle midnight.
	ShiftTypeCrossDay = "cross_day"
)

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
