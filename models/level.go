package models

import "time"

// Level represents a classification tier for clients
// e.g. Bronze, Silver, Gold, Platinum
type Level struct {
	ID          uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `gorm:"type:text" json:"description,omitempty"`
	Color       string  `gorm:"type:varchar(20);default:'#6366f1'" json:"color"`
	SortOrder   int     `gorm:"default:0" json:"sort_order"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	BranchID    uint    `gorm:"index" json:"branch_id,omitempty"`
	Branch      *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedByID uint    `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User   `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
