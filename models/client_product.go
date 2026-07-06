package models

import "time"

// ClientProduct represents the Player Section — a product/account
// the client holds (e.g. a loan account, investment account, savings account).
type ClientProduct struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientID      uint      `gorm:"not null;index" json:"client_id"`
	ProductTypeID uint      `gorm:"not null;index" json:"product_type_id"`
	AccountID     string    `gorm:"type:varchar(100);not null" json:"account_id"` // the external account identifier
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	SortOrder     int       `gorm:"default:0" json:"sort_order"`

	ProductType *ProductType `gorm:"foreignKey:ProductTypeID" json:"product_type,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
