package models

import "time"

// ClientPhone — phone numbers attached to a Client.
// Supports multiple phones, each with label, isPrimary, status, and isActive flags.
type ClientPhone struct {
	ID        uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientID  uint        `gorm:"not null;index" json:"client_id"`
	Phone     string      `gorm:"type:varchar(50);not null" json:"phone"`
	Label     string      `gorm:"type:varchar(50);default:'primary'" json:"label"` // e.g. primary, work, mobile, home
	IsPrimary bool        `gorm:"default:false" json:"is_primary"`
	Status    PhoneStatus `gorm:"type:enum('active','inactive');default:'active'" json:"status"`
	IsActive  bool        `gorm:"default:true" json:"is_active"`
	SortOrder int         `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
