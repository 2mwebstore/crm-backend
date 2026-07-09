package models

import "time"

// CompanyBank represents one of the COMPANY's own bank accounts used to
// receive client deposits (e.g. a branch's ABA account number + QR code),
// as opposed to BankType which is just a lookup of bank brands a client can
// select from when entering their own bank info.
type CompanyBank struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Which bank brand this account belongs to (ABA, ACLEDA, Wing, etc.)
	BankTypeID uint      `gorm:"not null;index" json:"bank_type_id"`
	BankType   *BankType `gorm:"foreignKey:BankTypeID" json:"bank_type,omitempty"`

	AccountNumber string `gorm:"type:varchar(100);not null" json:"account_number"`
	AccountName   string `gorm:"type:varchar(191);not null" json:"account_name"`

	// Optional: currency this account is denominated in (KHR/USD)
	CurrencyTypeID *uint         `gorm:"index" json:"currency_type_id,omitempty"`
	CurrencyType   *CurrencyType `gorm:"foreignKey:CurrencyTypeID" json:"currency_type,omitempty"`

	// Optional: static KHQR / scan-to-pay QR image for this account
	QRCodeURL string `gorm:"type:varchar(500)" json:"qr_code_url,omitempty"`

	// Cash on hand tracked separately from the bank account itself
	// (e.g. a branch's physical cash box, distinct from what's actually
	// sitting in the bank).
	Cash float64 `gorm:"type:decimal(18,2);not null;default:0" json:"cash"`

	IsActive  bool `gorm:"default:true" json:"is_active"`
	SortOrder int  `gorm:"default:0" json:"sort_order"`

	// Which branch this account belongs to (nil = shared/company-wide account)
	BranchID *uint   `gorm:"index" json:"branch_id,omitempty"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	CreatedByID uint  `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
