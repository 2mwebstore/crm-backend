package models

import "time"

// Permission is a single action string, e.g. "clients.view"
type Permission struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"type:varchar(191);not null" json:"display_name"`
	Group       string    `gorm:"type:varchar(100);not null" json:"group"`
	Description string    `gorm:"type:varchar(500)" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ── Permission constants ───────────────────────────────────────────────────────
const (
	// Branch
	PermBranchView   = "branch.view"
	PermBranchManage = "branch.manage"

	// Phone visibility
	PermPhoneView = "phone.view"

	// Clients
	PermClientView   = "clients.view"
	PermClientCreate = "clients.create"
	PermClientEdit   = "clients.edit"
	PermClientDelete = "clients.delete"
	PermClientExport = "clients.export"

	// Interesting Clients
	PermICView    = "interesting_clients.view"
	PermICCreate  = "interesting_clients.create"
	PermICEdit    = "interesting_clients.edit"
	PermICDelete  = "interesting_clients.delete"
	PermICConvert = "interesting_clients.convert"
	PermICExport  = "interesting_clients.export"

	// Users
	PermUserView   = "users.view"
	PermUserCreate = "users.create"
	PermUserEdit   = "users.edit"
	PermUserDelete = "users.delete"

	// Roles
	PermRoleView   = "roles.view"
	PermRoleCreate = "roles.create"
	PermRoleEdit   = "roles.edit"
	PermRoleDelete = "roles.delete"

	// Reports
	PermReportView = "reports.view"

	// Deposits
	PermDepositView   = "deposits.view"
	PermDepositCreate = "deposits.create"
	PermDepositEdit   = "deposits.edit"
	PermDepositDelete = "deposits.delete"

	// Withdrawals
	PermWithdrawalView   = "withdrawals.view"
	PermWithdrawalCreate = "withdrawals.create"
	PermWithdrawalEdit   = "withdrawals.edit"
	PermWithdrawalDelete = "withdrawals.delete"

	// Company Banks — controls the Top Up / Withdraw balance actions
	// specifically (moving real cash), kept separate from
	// configuration.manage/lookup.manage so a role can manage company bank
	// records (create/edit/delete accounts) without being able to move
	// money, or vice versa.
	PermCompanyBankTopup = "company_banks.topup"

	// Product Types — same idea, for the shared credit pool balance.
	PermProductTypeTopup = "product_types.topup"

	// Adjustment — a manual correction (Addition/Subtraction) against
	// Company Bank cash or Product Type credit, kept as its own
	// permission separate from the routine Top Up/Withdraw permission
	// above, since adjustments are a more sensitive "fix a mistake"
	// action that not every Top Up/Withdraw-capable role should
	// necessarily be trusted with.
	PermCompanyBankAdjustment = "company_banks.adjustment"
	PermProductTypeAdjustment = "product_types.adjustment"

	// ── Lookup tables, broken out into full View/Create/Edit/Delete ────────
	// Each lookup table now has its own dedicated CRUD permissions, instead
	// of every lookup entity sharing one blanket lookup.view/lookup.manage.
	// lookup.view/lookup.manage/configuration.view/configuration.manage
	// still work as broad "grant everything" fallbacks (checked via OR
	// alongside each specific permission), so existing roles aren't broken
	// by this change — they just gain the option of finer-grained roles
	// going forward.
	PermBankTypeView   = "bank_types.view"
	PermBankTypeCreate = "bank_types.create"
	PermBankTypeEdit   = "bank_types.edit"
	PermBankTypeDelete = "bank_types.delete"

	PermCompanyBankView   = "company_banks.view"
	PermCompanyBankCreate = "company_banks.create"
	PermCompanyBankEdit   = "company_banks.edit"
	PermCompanyBankDelete = "company_banks.delete"

	PermProductTypeView   = "product_types.view"
	PermProductTypeCreate = "product_types.create"
	PermProductTypeEdit   = "product_types.edit"
	PermProductTypeDelete = "product_types.delete"

	PermBonusOptionView   = "bonus_options.view"
	PermBonusOptionCreate = "bonus_options.create"
	PermBonusOptionEdit   = "bonus_options.edit"
	PermBonusOptionDelete = "bonus_options.delete"

	PermLevelView   = "levels.view"
	PermLevelCreate = "levels.create"
	PermLevelEdit   = "levels.edit"
	PermLevelDelete = "levels.delete"

	PermContactSourceView   = "contact_sources.view"
	PermContactSourceCreate = "contact_sources.create"
	PermContactSourceEdit   = "contact_sources.edit"
	PermContactSourceDelete = "contact_sources.delete"

	// Currencies — previously had NO permission model at all (gated only
	// by superAdminOnly at the frontend router level); now has the same
	// granular View/Create/Edit/Delete as every other lookup table.
	PermCurrencyView   = "currencies.view"
	PermCurrencyCreate = "currencies.create"
	PermCurrencyEdit   = "currencies.edit"
	PermCurrencyDelete = "currencies.delete"
)

// AllPermissions is the master list used for seeding.
var AllPermissions = []Permission{
	// Branch
	{Name: PermBranchView, DisplayName: "View Branches", Group: "branch", Description: "View branch list"},
	{Name: PermBranchManage, DisplayName: "Manage Branches", Group: "branch", Description: "Create/edit/delete branches"},
	// Phone visibility
	{Name: PermPhoneView, DisplayName: "View Phone Numbers", Group: "clients", Description: "See full phone numbers; without this they are masked as ***8755"},
	// Clients
	{Name: PermClientView, DisplayName: "View Clients", Group: "clients", Description: "View client list and detail"},
	{Name: PermClientCreate, DisplayName: "Create Clients", Group: "clients", Description: "Create new clients"},
	{Name: PermClientEdit, DisplayName: "Edit Clients", Group: "clients", Description: "Edit existing clients"},
	{Name: PermClientDelete, DisplayName: "Delete Clients", Group: "clients", Description: "Delete clients"},
	{Name: PermClientExport, DisplayName: "Export Clients", Group: "clients", Description: "Export client data"},
	// Interesting Clients
	{Name: PermICView, DisplayName: "View Interesting Clients", Group: "interesting_clients", Description: "View interesting client list and detail"},
	{Name: PermICCreate, DisplayName: "Create Interesting Clients", Group: "interesting_clients", Description: "Create new interesting clients"},
	{Name: PermICEdit, DisplayName: "Edit Interesting Clients", Group: "interesting_clients", Description: "Edit interesting clients"},
	{Name: PermICDelete, DisplayName: "Delete Interesting Clients", Group: "interesting_clients", Description: "Delete interesting clients"},
	{Name: PermICConvert, DisplayName: "Convert Interesting Clients", Group: "interesting_clients", Description: "Convert interesting client to real client"},
	{Name: PermICExport, DisplayName: "Export Interesting Clients", Group: "interesting_clients", Description: "Export interesting client data"},
	// Users
	{Name: PermUserView, DisplayName: "View Sub-Users", Group: "users", Description: "View own sub-users"},
	{Name: PermUserCreate, DisplayName: "Create Sub-Users", Group: "users", Description: "Create sub-users under own account"},
	{Name: PermUserEdit, DisplayName: "Edit Sub-Users", Group: "users", Description: "Edit own sub-users"},
	{Name: PermUserDelete, DisplayName: "Delete Sub-Users", Group: "users", Description: "Delete own sub-users"},
	// Roles
	{Name: PermRoleView, DisplayName: "View Roles", Group: "roles", Description: "View roles and permissions"},
	{Name: PermRoleCreate, DisplayName: "Create Roles", Group: "roles", Description: "Create new roles"},
	{Name: PermRoleEdit, DisplayName: "Edit Roles", Group: "roles", Description: "Edit existing roles"},
	{Name: PermRoleDelete, DisplayName: "Delete Roles", Group: "roles", Description: "Delete roles"},
	// Reports
	{Name: PermReportView, DisplayName: "View Reports", Group: "reports", Description: "Access dashboard and reports"},
	// Deposits
	{Name: PermDepositView, DisplayName: "View Deposits", Group: "deposits", Description: "View deposit transactions"},
	{Name: PermDepositCreate, DisplayName: "Create Deposits", Group: "deposits", Description: "Create new deposit records"},
	{Name: PermDepositEdit, DisplayName: "Edit Deposits", Group: "deposits", Description: "Edit existing deposit records"},
	{Name: PermDepositDelete, DisplayName: "Delete Deposits", Group: "deposits", Description: "Delete deposit records"},
	// Withdrawals
	{Name: PermWithdrawalView, DisplayName: "View Withdrawals", Group: "withdrawals", Description: "View withdrawal transactions"},
	{Name: PermWithdrawalCreate, DisplayName: "Create Withdrawals", Group: "withdrawals", Description: "Create new withdrawal records"},
	{Name: PermWithdrawalEdit, DisplayName: "Edit Withdrawals", Group: "withdrawals", Description: "Edit existing withdrawal records"},
	{Name: PermWithdrawalDelete, DisplayName: "Delete Withdrawals", Group: "withdrawals", Description: "Delete withdrawal records"},
	// Company Banks (balance control)
	{Name: PermCompanyBankTopup, DisplayName: "Top Up / Withdraw Company Bank Cash", Group: "company_banks", Description: "Add or remove cash on a company bank account, separate from managing the account record itself"},
	{Name: PermCompanyBankAdjustment, DisplayName: "Adjust Company Bank Cash", Group: "company_banks", Description: "Manually add or subtract cash on a company bank account as a correction"},
	// Product Types (balance control)
	{Name: PermProductTypeTopup, DisplayName: "Top Up / Withdraw Product Credit", Group: "product_types", Description: "Add or remove credit on a product type's shared credit pool, separate from managing the product record itself"},
	{Name: PermProductTypeAdjustment, DisplayName: "Adjust Product Credit", Group: "product_types", Description: "Manually add or subtract credit on a product type's shared credit pool as a correction"},
	// Lookup tables — per-function View/Create/Edit/Delete permissions
	{Name: PermBankTypeView, DisplayName: "View Bank Types", Group: "bank_types", Description: "View bank type records"},
	{Name: PermBankTypeCreate, DisplayName: "Create Bank Types", Group: "bank_types", Description: "Create new bank type records"},
	{Name: PermBankTypeEdit, DisplayName: "Edit Bank Types", Group: "bank_types", Description: "Edit existing bank type records"},
	{Name: PermBankTypeDelete, DisplayName: "Delete Bank Types", Group: "bank_types", Description: "Delete bank type records"},

	{Name: PermCompanyBankView, DisplayName: "View Company Banks", Group: "company_banks", Description: "View company bank records"},
	{Name: PermCompanyBankCreate, DisplayName: "Create Company Banks", Group: "company_banks", Description: "Create new company bank records"},
	{Name: PermCompanyBankEdit, DisplayName: "Edit Company Banks", Group: "company_banks", Description: "Edit existing company bank records"},
	{Name: PermCompanyBankDelete, DisplayName: "Delete Company Banks", Group: "company_banks", Description: "Delete company bank records"},

	{Name: PermProductTypeView, DisplayName: "View Product Types", Group: "product_types", Description: "View product type records"},
	{Name: PermProductTypeCreate, DisplayName: "Create Product Types", Group: "product_types", Description: "Create new product type records"},
	{Name: PermProductTypeEdit, DisplayName: "Edit Product Types", Group: "product_types", Description: "Edit existing product type records"},
	{Name: PermProductTypeDelete, DisplayName: "Delete Product Types", Group: "product_types", Description: "Delete product type records"},

	{Name: PermBonusOptionView, DisplayName: "View Bonus Options", Group: "bonus_options", Description: "View bonus option records"},
	{Name: PermBonusOptionCreate, DisplayName: "Create Bonus Options", Group: "bonus_options", Description: "Create new bonus option records"},
	{Name: PermBonusOptionEdit, DisplayName: "Edit Bonus Options", Group: "bonus_options", Description: "Edit existing bonus option records"},
	{Name: PermBonusOptionDelete, DisplayName: "Delete Bonus Options", Group: "bonus_options", Description: "Delete bonus option records"},

	{Name: PermLevelView, DisplayName: "View Levels", Group: "levels", Description: "View level records"},
	{Name: PermLevelCreate, DisplayName: "Create Levels", Group: "levels", Description: "Create new level records"},
	{Name: PermLevelEdit, DisplayName: "Edit Levels", Group: "levels", Description: "Edit existing level records"},
	{Name: PermLevelDelete, DisplayName: "Delete Levels", Group: "levels", Description: "Delete level records"},

	{Name: PermContactSourceView, DisplayName: "View Contact Sources", Group: "contact_sources", Description: "View contact source records"},
	{Name: PermContactSourceCreate, DisplayName: "Create Contact Sources", Group: "contact_sources", Description: "Create new contact source records"},
	{Name: PermContactSourceEdit, DisplayName: "Edit Contact Sources", Group: "contact_sources", Description: "Edit existing contact source records"},
	{Name: PermContactSourceDelete, DisplayName: "Delete Contact Sources", Group: "contact_sources", Description: "Delete contact source records"},

	{Name: PermCurrencyView, DisplayName: "View Currencies", Group: "currencies", Description: "View currency records"},
	{Name: PermCurrencyCreate, DisplayName: "Create Currencies", Group: "currencies", Description: "Create new currency records"},
	{Name: PermCurrencyEdit, DisplayName: "Edit Currencies", Group: "currencies", Description: "Edit existing currency records"},
	{Name: PermCurrencyDelete, DisplayName: "Delete Currencies", Group: "currencies", Description: "Delete currency records"},
}

// TransactionPermissions is kept for backward-compatibility with seed code in routes.go
// It is now empty since all permissions are in AllPermissions above.
var TransactionPermissions = []Permission{}

const (
	PermDepositApprove    = "deposits.approve"
	PermWithdrawalApprove = "withdrawals.approve"
)

const (
	PermTurnoverView    = "turnover_bets.view"
	PermTurnoverCreate  = "turnover_bets.create"
	PermTurnoverEdit    = "turnover_bets.edit"
	PermTurnoverDelete  = "turnover_bets.delete"
	PermTurnoverApprove = "turnover_bets.approve"
	PermFollowUpView    = "follow_ups.view"
	PermFollowUpCreate  = "follow_ups.create"
	PermFollowUpDelete  = "follow_ups.delete"
)

const (
	// PermDailyBalanceView covers seeing the Daily Balance page itself —
	// current totals, income, and History — without necessarily being
	// able to act on it.
	PermDailyBalanceView = "daily_balance.view"
	// PermDailyBalanceStart/Close gate the two actions separately, so a
	// role can be granted view-only access, or the ability to open a
	// shift but not close it (or vice versa) — e.g. a supervisor who
	// closes out shifts opened by regular staff.
	PermDailyBalanceStart = "daily_balance.start"
	PermDailyBalanceClose = "daily_balance.close"
)

func init() {
	AllPermissions = append(AllPermissions,
		Permission{Name: PermDepositApprove, DisplayName: "Approve Deposits", Group: "deposits", Description: "Approve or reject deposit transactions"},
		Permission{Name: PermWithdrawalApprove, DisplayName: "Approve Withdrawals", Group: "withdrawals", Description: "Approve or reject withdrawal transactions"},
		Permission{Name: PermTurnoverView, DisplayName: "View Turnover Bets", Group: "turnover_bets", Description: "View turnover bet records"},
		Permission{Name: PermTurnoverCreate, DisplayName: "Create Turnover Bets", Group: "turnover_bets", Description: "Create new turnover bet records"},
		Permission{Name: PermTurnoverEdit, DisplayName: "Edit Turnover Bets", Group: "turnover_bets", Description: "Edit turnover bet records"},
		Permission{Name: PermTurnoverDelete, DisplayName: "Delete Turnover Bets", Group: "turnover_bets", Description: "Delete turnover bet records"},
		Permission{Name: PermTurnoverApprove, DisplayName: "Approve Turnover Bets", Group: "turnover_bets", Description: "Approve or reject turnover bets"},
		Permission{Name: PermFollowUpView, DisplayName: "View Follow Ups", Group: "follow_ups", Description: "View follow-up records"},
		Permission{Name: PermFollowUpCreate, DisplayName: "Create Follow Ups", Group: "follow_ups", Description: "Create follow-up records"},
		Permission{Name: PermFollowUpDelete, DisplayName: "Delete Follow Ups", Group: "follow_ups", Description: "Delete follow-up records"},
		Permission{Name: PermDailyBalanceView, DisplayName: "View Daily Balance", Group: "daily_balance", Description: "View the Daily Balance page — current totals, income, and shift History"},
		Permission{Name: PermDailyBalanceStart, DisplayName: "Start Shift", Group: "daily_balance", Description: "Open a new Daily Balance shift for a branch"},
		Permission{Name: PermDailyBalanceClose, DisplayName: "Close Shift", Group: "daily_balance", Description: "Close the currently open Daily Balance shift for a branch"},
	)
}
