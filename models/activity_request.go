package models

import "time"

// ActivityRequest is a staff self-declaration of field/activity work for a
// given date — always a single date, and always auto-approved on
// submission (see ActivityRequestService.Create): the whole point is that
// it's effective immediately, not gated behind manual review like
// Leave/Overtime. BranchID is required here specifically (unlike Leave/
// Overtime, where it's optional) since it's what drives the automatic
// Attendance check-in/check-out — see ActivityRequestController.Create.
type ActivityRequest struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	UserID   uint    `gorm:"not null;index" json:"user_id"`
	User     *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BranchID uint    `gorm:"not null;index" json:"branch_id"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	Date string `gorm:"type:date;not null;index" json:"date"`

	Reason string `gorm:"type:varchar(500)" json:"reason,omitempty"`

	// Always "approved" — kept as a field (rather than assumed implicitly)
	// so the table's shape stays consistent with Leave/Overtime and
	// leaves room for a future admin override, without that being
	// something the current UI exposes.
	Status     string    `gorm:"type:varchar(20);not null;default:approved" json:"status"`
	ApprovedAt time.Time `json:"approved_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName pins this to the existing "activity_requests" table.
// Renamed the Go type to ActivityRequest for cleaner code, but NOT the
// underlying table — GORM's AutoMigrate doesn't rename tables, it creates
// a new one matching whatever the type would default to, which would
// silently orphan all existing data. This keeps the rename purely
// cosmetic at the Go/API level.
func (ActivityRequest) TableName() string {
	return "activity_requests"
}
