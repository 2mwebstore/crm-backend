package models

import "time"

type Client struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Code       string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name       string     `gorm:"type:varchar(191);not null" json:"name"`
	DateJoined *time.Time `json:"date_joined,omitempty"`
	Remark     string     `gorm:"type:text" json:"remark,omitempty"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`

	// ── Foreign Keys ───────────────────────────────────────────────────────
	BranchID        *uint `gorm:"index" json:"branch_id,omitempty"`
	LevelID         *uint `gorm:"index" json:"level_id,omitempty"`
	ContactSourceID *uint `gorm:"index" json:"contact_source_id,omitempty"`
	CreatedByID     uint  `gorm:"not null;index" json:"created_by_id"`

	// ── Relations ──────────────────────────────────────────────────────────
	Branch        *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Level         *Level         `gorm:"foreignKey:LevelID" json:"level,omitempty"`
	ContactSource *ContactSource `gorm:"foreignKey:ContactSourceID" json:"contact_source,omitempty"`
	CreatedBy     *User          `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// ── Sections ───────────────────────────────────────────────────────────
	Phones    []ClientPhone    `gorm:"foreignKey:ClientID" json:"phones,omitempty"`
	Banks     []ClientBank     `gorm:"foreignKey:ClientID" json:"banks,omitempty"`
	Products  []ClientProduct  `gorm:"foreignKey:ClientID" json:"products,omitempty"`
	FollowUps []ClientFollowUp `gorm:"foreignKey:ClientID" json:"follow_ups,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
