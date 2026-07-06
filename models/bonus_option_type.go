package models

import "time"

// BonusOptionType defines a simple bonus label that can be selected
// on deposits and withdrawals. The actual bonus amount is entered manually.
type BonusOptionType struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(191);not null" json:"name"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	BranchID     *uint  `gorm:"index" json:"branch_id,omitempty"`
	Branch       *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedByID uint      `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
