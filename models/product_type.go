package models

import "time"

// ProductType represents a category of product or service offered
// e.g. Loan, Savings, Insurance, Investment, Credit Card, etc.
type ProductType struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(191);not null" json:"name"`
	Code        string `gorm:"type:varchar(50)" json:"code"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Icon        string `gorm:"type:varchar(100)" json:"icon,omitempty"`

	// Currency this product is denominated in (KHR/USD)
	CurrencyTypeID *uint         `gorm:"index" json:"currency_type_id,omitempty"`
	CurrencyType   *CurrencyType `gorm:"foreignKey:CurrencyTypeID" json:"currency_type,omitempty"`

	// Whether this product can be purchased/opened on credit
	Credit float64 `gorm:"type:decimal(18,2);not null;default:0" json:"credit"`

	IsActive    bool    `gorm:"default:true" json:"is_active"`
	SortOrder   int     `gorm:"default:0" json:"sort_order"`
	BranchID    *uint   `gorm:"index" json:"branch_id,omitempty"`
	Branch      *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedByID uint    `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User   `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
