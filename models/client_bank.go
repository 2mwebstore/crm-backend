package models

import "time"

// ClientBank represents a bank account record in the Bank Section.
// A client can have multiple bank accounts across different banks.
type ClientBank struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientID    uint      `gorm:"not null;index" json:"client_id"`
	BankTypeID  uint      `gorm:"not null;index" json:"bank_type_id"`
	AccountNo   string    `gorm:"type:varchar(100);not null" json:"account_no"`
	AccountName string    `gorm:"type:varchar(191);not null" json:"account_name"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`

	BankType *BankType `gorm:"foreignKey:BankTypeID" json:"bank_type,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
