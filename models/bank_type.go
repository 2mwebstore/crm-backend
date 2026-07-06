package models

import "time"

// BankType represents a bank or financial institution option
// e.g. ABA Bank, ACLEDA, Canadia Bank, Wing, etc.
type BankType struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(191);not null" json:"name"`
	Code        string `gorm:"type:varchar(50)" json:"code"` // e.g. "ABA", "ACLEDA"
	Logo        string `gorm:"type:varchar(500)" json:"logo,omitempty"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	BranchID     *uint  `gorm:"index" json:"branch_id,omitempty"`
	Branch       *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedByID uint   `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User  `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
