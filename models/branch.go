package models

import "time"

// Branch represents an organizational branch.
// Super Admin assigns branches to users (many-to-many via user_branches).
// The Code field is used as a prefix in document code generation:
//   Format: {ENTITY_PREFIX}-{YYYYMMDD}-{BRANCH_CODE}
//   e.g.   INT-20260701-CRNS
type Branch struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Code        string    `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"` // short code, e.g. "CRNS"
	Description string    `gorm:"type:varchar(500)" json:"description,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedByID uint      `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
