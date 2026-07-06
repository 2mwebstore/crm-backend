package models

import "time"

// TurnoverBet records a turnover bet entry based on a product type.
// No client link — it's a product-level record only.
type TurnoverBet struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Date          time.Time `gorm:"not null" json:"date"`
	ProductTypeID uint      `gorm:"not null;index" json:"product_type_id"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Currency      string    `gorm:"type:varchar(10);default:'USD'" json:"currency"`
	Remark        string    `gorm:"type:text" json:"remark,omitempty"`

	Status       TransactionStatus `gorm:"type:enum('pending','approved','rejected');default:'pending'" json:"status"`
	ApprovedAt   *time.Time        `json:"approved_at,omitempty"`
	ApprovedByID *uint             `gorm:"index" json:"approved_by_id,omitempty"`
	BranchID     *uint             `gorm:"index" json:"branch_id,omitempty"`
	CreatedByID  uint              `gorm:"not null;index" json:"created_by_id"`

	// Relations
	ProductType *ProductType `gorm:"foreignKey:ProductTypeID" json:"product_type,omitempty"`
	Branch      *Branch      `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedBy   *User        `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	ApprovedBy  *User        `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
