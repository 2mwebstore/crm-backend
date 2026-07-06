package models

import "time"

// Withdrawal records a client withdrawal transaction.
type Withdrawal struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TransactionNo string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"transaction_no"`
	Date          time.Time `gorm:"not null" json:"date"`

	// ── Client / Product linkage ────────────────────────────────────────────
	ClientID        uint `gorm:"not null;index" json:"client_id"`
	ClientProductID uint `gorm:"not null;index" json:"client_product_id"`
	ClientBankID    uint `gorm:"not null;index" json:"client_bank_id"`
	CompanyBankID   uint `gorm:"not null;index" json:"company_bank_id"`

	// ── Amounts ────────────────────────────────────────────────────────────
	Amount      float64 `gorm:"type:decimal(15,2);not null" json:"amount"`
	BonusAmount float64 `gorm:"type:decimal(15,2);default:0" json:"bonus_amount"`
	Bal         float64 `gorm:"type:decimal(15,2);default:0" json:"bal"`
	TO          float64 `gorm:"type:decimal(15,2);default:0" json:"to"`
	OS          float64 `gorm:"type:decimal(15,2);default:0" json:"os"`
	Play        float64 `gorm:"type:decimal(15,2);default:0" json:"play"`
	Currency    string  `gorm:"type:varchar(10);default:'USD'" json:"currency"`

	// ── Bonus option ────────────────────────────────────────────────────────
	BonusOptionID *uint `gorm:"index" json:"bonus_option_id,omitempty"`

	// ── Approval workflow ───────────────────────────────────────────────────
	Status       TransactionStatus `gorm:"type:enum('pending','approved','rejected');default:'pending'" json:"status"`
	ApprovedAt   *time.Time        `json:"approved_at,omitempty"`
	ApprovedByID *uint             `gorm:"index" json:"approved_by_id,omitempty"`

	// ── Meta ────────────────────────────────────────────────────────────────
	Remark      string `gorm:"type:text" json:"remark,omitempty"`
	BranchID    *uint  `gorm:"index" json:"branch_id,omitempty"`
	CreatedByID uint   `gorm:"not null;index" json:"created_by_id"`

	// ── Relations ───────────────────────────────────────────────────────────
	Client        *Client          `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	ClientProduct *ClientProduct   `gorm:"foreignKey:ClientProductID" json:"client_product,omitempty"`
	ClientBank    *ClientBank      `gorm:"foreignKey:ClientBankID" json:"client_bank,omitempty"`
	CompanyBank   *BankType        `gorm:"foreignKey:CompanyBankID" json:"company_bank,omitempty"`
	BonusOption   *BonusOptionType `gorm:"foreignKey:BonusOptionID" json:"bonus_option,omitempty"`
	Branch        *Branch          `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedBy     *User            `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	ApprovedBy    *User            `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
