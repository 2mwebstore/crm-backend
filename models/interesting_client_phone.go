package models

import "time"

type PhoneStatus string

const (
	PhoneStatusActive   PhoneStatus = "active"
	PhoneStatusInactive PhoneStatus = "inactive"
)

type InterestingClientPhone struct {
	ID                  uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	InterestingClientID uint        `gorm:"not null;index" json:"interesting_client_id"`
	Phone               string      `gorm:"type:varchar(50);not null" json:"phone"`
	Label               string      `gorm:"type:varchar(50);default:'primary'" json:"label"`
	IsPrimary           bool        `gorm:"default:false" json:"is_primary"`
	Status              PhoneStatus `gorm:"type:enum('active','inactive');default:'active'" json:"status"`
	IsActive            bool        `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}
