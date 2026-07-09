package models

import "time"

// DailyStartBalanceEntityType distinguishes what a DailyStartBalanceDetail
// row is describing.
type DailyStartBalanceEntityType string

const (
	DailyBalanceEntityCompanyBank DailyStartBalanceEntityType = "company_bank"
	DailyBalanceEntityProductType DailyStartBalanceEntityType = "product_type"
)

// DailyBalancePhase distinguishes an opening ("Start Today") detail row
// from a closing ("Close Today") detail row on the same day.
type DailyBalancePhase string

const (
	DailyBalancePhaseOpen  DailyBalancePhase = "open"
	DailyBalancePhaseClose DailyBalancePhase = "close"
)

// DailyStartBalance is now a SHIFT record rather than a strictly-once-per-
// day snapshot: a branch can be opened and closed multiple times in one
// day (one per staff shift). At most one shift can be OPEN (ClosedAt IS
// NULL) per branch at any moment — starting a new one while one is
// already open is rejected, and closing requires an open one to exist.
// "Income" for the currently open shift is always computed relative to
// THAT shift's own CreatedAt/opening totals — never the day's very first
// shift — so rotating shifts through the day never mixes up whose income
// belongs to whom.
type DailyStartBalance struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	BranchID uint    `gorm:"not null;index" json:"branch_id"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	// Date is informational only now (the calendar day this shift opened
	// on, Asia/Phnom_Penh) — no longer a uniqueness key, since several
	// shifts can share the same branch+day.
	Date *time.Time `gorm:"type:date;index" json:"date"`

	// Opening totals — split by currency, since Company Bank Cash and
	// Product Credit can each be denominated in USD or KHR per record.
	CashUSD   float64 `gorm:"type:decimal(18,2);not null" json:"cash_usd"`
	CashKHR   float64 `gorm:"type:decimal(18,2);not null" json:"cash_khr"`
	CreditUSD float64 `gorm:"type:decimal(18,2);not null" json:"credit_usd"`
	CreditKHR float64 `gorm:"type:decimal(18,2);not null" json:"credit_khr"`

	CreatedByID uint      `gorm:"not null;index" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"` // Asia/Phnom_Penh

	// Closing totals — nil until someone closes the day out.
	CloseCashUSD   *float64 `gorm:"type:decimal(18,2)" json:"close_cash_usd,omitempty"`
	CloseCashKHR   *float64 `gorm:"type:decimal(18,2)" json:"close_cash_khr,omitempty"`
	CloseCreditUSD *float64 `gorm:"type:decimal(18,2)" json:"close_credit_usd,omitempty"`
	CloseCreditKHR *float64 `gorm:"type:decimal(18,2)" json:"close_credit_khr,omitempty"`

	ClosedByID *uint      `gorm:"index" json:"closed_by_id,omitempty"`
	ClosedBy   *User      `gorm:"foreignKey:ClosedByID" json:"closed_by,omitempty"`
	ClosedAt   *time.Time `gorm:"index" json:"closed_at,omitempty"` // Asia/Phnom_Penh — NULL means this shift is still open

	// Details is the persisted per-bank/per-product breakdown — both the
	// opening (Phase=open) and, once present, closing (Phase=close) rows.
	Details []DailyStartBalanceDetail `gorm:"foreignKey:DailyStartBalanceID" json:"details,omitempty"`
}

// DailyStartBalanceDetail is one line item behind a DailyStartBalance
// snapshot — the individual CompanyBank or ProductType and its amount at
// the exact moment that snapshot phase (open/close) was taken. Stored
// denormalized (the entity's ID *and* a captured label/currency) rather
// than just a live join, so history stays accurate even if the underlying
// bank/product is later renamed, has its currency changed, or is deleted
// entirely.
type DailyStartBalanceDetail struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	DailyStartBalanceID uint              `gorm:"not null;index" json:"daily_start_balance_id"`
	Phase               DailyBalancePhase `gorm:"type:varchar(10);not null;index" json:"phase"` // "open" | "close"

	EntityType DailyStartBalanceEntityType `gorm:"type:varchar(30);not null;index" json:"entity_type"`
	EntityID   uint                        `gorm:"not null;index" json:"entity_id"`

	// Label and Currency are captured at snapshot time — e.g. the account
	// name for a company bank, or the product name for a product type —
	// so this row still makes sense even if the source record changes or
	// disappears later.
	Label    string  `gorm:"type:varchar(191)" json:"label"`
	Currency string  `gorm:"type:varchar(10)" json:"currency"`
	Amount   float64 `gorm:"type:decimal(18,2);not null" json:"amount"`
}
