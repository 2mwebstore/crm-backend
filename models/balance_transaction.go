package models

import "time"

// BalanceTxType distinguishes a deposit into a balance from a withdrawal out of it.
type BalanceTxType string

const (
	BalanceTxTopUp      BalanceTxType = "topup"
	BalanceTxWithdrawal BalanceTxType = "withdrawal"
)

// BalanceEntityType identifies which table/entity a BalanceTransaction row
// belongs to, since this ledger is shared across multiple balance fields
// (CompanyBank.Cash, ProductType.Credit, and any future ones) rather than
// having a separate transaction table per entity.
type BalanceEntityType string

const (
	BalanceEntityCompanyBank BalanceEntityType = "company_bank"
	BalanceEntityProductType BalanceEntityType = "product_type"
)

// BalanceTransaction is an immutable audit record of a single top-up or
// withdrawal against a balance field (e.g. CompanyBank.Cash or
// ProductType.Credit). Rows are never updated or deleted — always insert a
// new one, so old_amount/new_amount always describe an actual DB state
// transition and the full history is fully reconstructable.
type BalanceTransaction struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Which entity + field this transaction affected.
	EntityType BalanceEntityType `gorm:"type:varchar(50);not null;index:idx_balance_entity" json:"entity_type"`
	EntityID   uint              `gorm:"not null;index:idx_balance_entity" json:"entity_id"`
	Field      string            `gorm:"type:varchar(50);not null" json:"field"` // e.g. "cash", "credit"

	Type BalanceTxType `gorm:"type:varchar(20);not null" json:"type"` // "topup" | "withdrawal"

	// OldAmount/NewAmount are the balance immediately before/after this
	// transaction; Amount is always the positive magnitude of the change —
	// direction is expressed by Type, not by the sign of Amount.
	OldAmount float64 `gorm:"type:decimal(18,2);not null" json:"old_amount"`
	Amount    float64 `gorm:"type:decimal(18,2);not null" json:"amount"`
	NewAmount float64 `gorm:"type:decimal(18,2);not null" json:"new_amount"`

	Remark string `gorm:"type:varchar(255)" json:"remark,omitempty"`

	CreatedByID uint  `gorm:"index;not null" json:"created_by_id"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}
