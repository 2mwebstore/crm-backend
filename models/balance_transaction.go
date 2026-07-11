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

// BalanceTxSource identifies WHERE a BalanceTransaction originated —
// explicit and set at creation time, rather than inferred later by
// guessing from the Remark text (which is fragile: any future remark
// wording change would silently break that classification).
type BalanceTxSource string

const (
	// BalanceSourceTransaction = created automatically as a side effect of
	// a client Deposit or Withdrawal (Transactions module).
	BalanceSourceTransaction BalanceTxSource = "transaction"
	// BalanceSourceConfiguration = created by a direct manual Top Up /
	// Withdraw action on a Company Bank or Product Type record itself
	// (Configuration module), not tied to any client transaction.
	BalanceSourceConfiguration BalanceTxSource = "configuration"
	// BalanceSourceAdjustment = a manual correction — used when staff need
	// to fix a mistake or account for something outside the normal Top
	// Up/Withdraw or client Deposit/Withdrawal flows (e.g. reconciling
	// against a physical cash count). Kept as its own source rather than
	// folded into "configuration" so adjustments are always clearly
	// distinguishable in the ledger and reporting from routine top-ups.
	BalanceSourceAdjustment BalanceTxSource = "adjustment"
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

	Type   BalanceTxType   `gorm:"type:varchar(20);not null" json:"type"`         // "topup" | "withdrawal"
	Source BalanceTxSource `gorm:"type:varchar(20);not null;index" json:"source"` // "transaction" | "configuration"

	// OldAmount/NewAmount are the balance immediately before/after this
	// transaction; Amount is always the positive magnitude of the change —
	// direction is expressed by Type, not by the sign of Amount.
	OldAmount float64 `gorm:"type:decimal(18,2);not null" json:"old_amount"`
	Amount    float64 `gorm:"type:decimal(18,2);not null" json:"amount"`
	NewAmount float64 `gorm:"type:decimal(18,2);not null" json:"new_amount"`

	// BonusAmount is the portion of Amount that came from a Deposit's
	// bonus (always 0 for non-Deposit-originated entries — Withdrawals,
	// direct Configuration actions, and Adjustments never carry a bonus).
	// Purely descriptive: the balance change already reflects the full
	// Amount (principal + bonus combined) — this field just lets readers
	// (like the Daily Balance ledger view) see how much of that change was
	// bonus vs the client's actual deposited cash, without altering the
	// balance math itself.
	BonusAmount float64 `gorm:"type:decimal(18,2);not null;default:0" json:"bonus_amount"`

	Remark string `gorm:"type:varchar(255)" json:"remark,omitempty"`

	CreatedByID uint  `gorm:"index;not null" json:"created_by_id"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}
