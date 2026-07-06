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

	// Configuration (system setup — Branches, Lookup tables, Exchange Rates, Roles, Users)
	PermConfigView   = "configuration.view"
	PermConfigManage = "configuration.manage" // without this, phone numbers are masked

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

	// Lookup
	PermLookupView   = "lookup.view"
	PermLookupManage = "lookup.manage"

	// Exchange Rates
	PermExchangeView   = "exchange_rates.view"
	PermExchangeManage = "exchange_rates.manage"

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
)

// AllPermissions is the master list used for seeding.
var AllPermissions = []Permission{
	// Configuration
	{Name: PermConfigView, DisplayName: "View Configuration", Group: "configuration", Description: "Access settings: branches, lookup tables, exchange rates, roles, users"},
	{Name: PermConfigManage, DisplayName: "Manage Configuration", Group: "configuration", Description: "Create/edit/delete branches, lookup tables, exchange rates, roles, users"},
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
	// Lookup
	{Name: PermLookupView, DisplayName: "View Lookup Data", Group: "lookup", Description: "View banks, products, bonus options, currencies"},
	{Name: PermLookupManage, DisplayName: "Manage Lookup Data", Group: "lookup", Description: "Create/edit/delete lookup tables"},
	// Exchange Rates
	{Name: PermExchangeView, DisplayName: "View Exchange Rates", Group: "exchange_rates", Description: "View and convert currency rates"},
	{Name: PermExchangeManage, DisplayName: "Manage Exchange Rates", Group: "exchange_rates", Description: "Create/edit/delete exchange rates"},
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
	)
}
