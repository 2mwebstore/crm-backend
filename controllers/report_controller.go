package controllers

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	transactiondto "crm-backend/dto/transaction"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type ReportController struct {
	clientRepo      repositories.ClientRepository
	icRepo          repositories.InterestingClientRepository
	depositRepo     repositories.DepositRepository
	withdrawalRepo  repositories.WithdrawalRepository
	userRepo        repositories.UserRepository
	companyBankRepo repositories.CompanyBankRepository
	bankTypeRepo    repositories.BankTypeRepository
}

func NewReportController(
	clientRepo repositories.ClientRepository,
	icRepo repositories.InterestingClientRepository,
	depositRepo repositories.DepositRepository,
	withdrawalRepo repositories.WithdrawalRepository,
	userRepo repositories.UserRepository,
	companyBankRepo repositories.CompanyBankRepository,
	bankTypeRepo repositories.BankTypeRepository,
) *ReportController {
	return &ReportController{clientRepo, icRepo, depositRepo, withdrawalRepo, userRepo, companyBankRepo, bankTypeRepo}
}

func (ctrl *ReportController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userRepo.GetScopeIDs(middlewares.GetUserID(c))
	return ids
}

// Summary godoc — GET /reports/summary
func (ctrl *ReportController) Summary(c *gin.Context) {
	scopeIDs := ctrl.scope(c)

	// Count clients
	var totalClients, activeClients int64
	// Count ICs
	var totalICs, convertedICs int64
	// Transaction sums
	var totalDeposits, totalWithdrawals float64
	var depCount, wdrCount int64

	db := c.MustGet("db")
	if db == nil {
		utils.InternalError(c, nil)
		return
	}

	utils.OK(c, "success", gin.H{
		"scope_user_count":  len(scopeIDs),
		"total_clients":     totalClients,
		"active_clients":    activeClients,
		"total_ics":         totalICs,
		"converted_ics":     convertedICs,
		"total_deposits":    totalDeposits,
		"deposit_count":     depCount,
		"total_withdrawals": totalWithdrawals,
		"withdrawal_count":  wdrCount,
	})
}

// mergedTx is an intermediate holder used only to merge-sort deposits and
// withdrawals together before flattening — never serialized directly.
type mergedTx struct {
	Type string // "deposit" | "withdrawal"
	Date time.Time
	Data interface{}
}

// flattenWithType converts a Deposit/Withdrawal struct into a plain
// map[string]interface{} (via its own JSON marshaling, so every existing
// field/relation on the model comes through exactly as-is) and adds a
// "type" key alongside them — a true flat merge rather than a nested
// {type, data} wrapper.
func flattenWithType(record interface{}, txType string) (map[string]interface{}, error) {
	b, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	var row map[string]interface{}
	if err := json.Unmarshal(b, &row); err != nil {
		return nil, err
	}
	row["type"] = txType
	return row, nil
}

// AllTransactions godoc — GET /reports/transactions
// Merges deposits and withdrawals into a single flat feed — each item has
// all of its own Deposit/Withdrawal fields plus a "type": "deposit" |
// "withdrawal" key — sorted by date descending (newest first) by default.
// Pass sort_dir=asc to reverse it.
//
// NOTE: this fetches ALL matching deposit+withdrawal rows for the given
// filter/scope, merges + sorts them in memory, then paginates the merged
// slice — it does not paginate at the DB level per table (a true DB-level
// merge would need a SQL UNION across two distinct model types, which GORM
// doesn't do cleanly). Fine for moderate data volumes; if this grows very
// large, consider a UNION query or a dedicated transactions view instead.
func (ctrl *ReportController) AllTransactions(c *gin.Context) {
	var filter transactiondto.FilterQuery
	if err := utils.BindQuery(c, &filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	p := utils.ParsePagination(c)
	userID := middlewares.GetUserID(c)

	// Fetch everything matching the filter/scope from both tables — no
	// per-table pagination here, since we need the full set to merge-sort
	// correctly before applying pagination ourselves below.
	fetchAll := utils.PaginationParams{Page: 1, PageSize: 1000000}

	deposits, _, err := ctrl.depositRepo.List(filter, fetchAll, userID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	withdrawals, _, err := ctrl.withdrawalRepo.List(filter, fetchAll, userID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	merged := make([]mergedTx, 0, len(deposits)+len(withdrawals))
	for i := range deposits {
		merged = append(merged, mergedTx{Type: "deposit", Date: deposits[i].Date, Data: deposits[i]})
	}
	for i := range withdrawals {
		merged = append(merged, mergedTx{Type: "withdrawal", Date: withdrawals[i].Date, Data: withdrawals[i]})
	}

	// Descending (newest first) unless explicitly asked for ascending.
	asc := filter.SortDir == "asc"
	sort.Slice(merged, func(i, j int) bool {
		if asc {
			return merged[i].Date.Before(merged[j].Date)
		}
		return merged[i].Date.After(merged[j].Date)
	})

	total := int64(len(merged))
	start := (p.Page - 1) * p.PageSize
	if start > len(merged) {
		start = len(merged)
	}
	end := start + p.PageSize
	if end > len(merged) {
		end = len(merged)
	}

	// Flatten only this page's worth of records — no need to marshal
	// everything up front.
	pageOut := make([]map[string]interface{}, 0, end-start)
	for _, m := range merged[start:end] {
		row, err := flattenWithType(m.Data, m.Type)
		if err != nil {
			utils.InternalError(c, err)
			return
		}
		pageOut = append(pageOut, row)
	}

	utils.OKPaginated(c, pageOut, utils.BuildMeta(p, total))
}

// currencyAmounts holds a USD/KHR pair for one summary cell.
type currencyAmounts struct {
	USD float64 `json:"usd"`
	KHR float64 `json:"khr"`
}

// bankSummaryRow is one row of the Bank × Currency breakdown table
// (Bank | Deposit USD/KHR | Withdrawal USD/KHR | Total USD/KHR | Bonus USD/KHR).
type bankSummaryRow struct {
	Bank       string          `json:"bank"`
	Deposit    currencyAmounts `json:"deposit"`
	Withdrawal currencyAmounts `json:"withdrawal"`
	Total      currencyAmounts `json:"total"`
	Bonus      currencyAmounts `json:"bonus"`
}

// bankLabel resolves the display name for a company bank ID — grouped
// strictly by bank *type* (brand, e.g. "ABA", "Wing"), never by the
// individual account, by joining through companyBankByID -> bankTypeByID
// in plain Go. This is done manually rather than relying on
// Deposit/Withdrawal preloading CompanyBank.BankType, since CompanyBank as
// currently returned only exposes bank_type_id (a raw FK), never a nested
// bank_type object. Grouping strictly by type (not falling back to the
// account's own name) ensures multiple CompanyBank accounts under the same
// bank type always merge into a single row rather than fragmenting.
func bankLabel(companyBankID uint, companyBankByID map[uint]models.CompanyBank, bankTypeByID map[uint]models.BankType) string {
	cb, ok := companyBankByID[companyBankID]
	if !ok {
		return "Unknown"
	}
	if bt, ok := bankTypeByID[cb.BankTypeID]; ok {
		if bt.Code != "" {
			return bt.Code
		}
		if bt.Name != "" {
			return bt.Name
		}
	}
	return "Unknown"
}

// BankSummary godoc — GET /reports/bank-summary
// Breaks down deposits, withdrawals, net total (deposit - withdrawal), and
// bonus amounts by company bank (grouped by bank brand) and currency
// (USD/KHR) — the "Bank | Deposit | Withdrawal | Total | Bonus" table.
// Accepts the same filters as AllTransactions (client_id, branch_id,
// date_from/date_to, etc.) via transactiondto.FilterQuery.
func (ctrl *ReportController) BankSummary(c *gin.Context) {
	var filter transactiondto.FilterQuery
	if err := utils.BindQuery(c, &filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	userID := middlewares.GetUserID(c)
	fetchAll := utils.PaginationParams{Page: 1, PageSize: 1000000}

	deposits, _, err := ctrl.depositRepo.List(filter, fetchAll, userID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	withdrawals, _, err := ctrl.withdrawalRepo.List(filter, fetchAll, userID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	// Build companyBankID -> CompanyBank and bankTypeID -> BankType lookup
	// maps once, up front, so bankLabel() can resolve each transaction's
	// bank brand via plain ID lookups — no dependency on Deposit/Withdrawal
	// preloading any nested relation.
	companyBanks, err := ctrl.companyBankRepo.List(true)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	companyBankByID := make(map[uint]models.CompanyBank, len(companyBanks))
	for _, cb := range companyBanks {
		companyBankByID[cb.ID] = cb
	}

	bankTypes, err := ctrl.bankTypeRepo.List(true)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	bankTypeByID := make(map[uint]models.BankType, len(bankTypes))
	for _, bt := range bankTypes {
		bankTypeByID[bt.ID] = bt
	}

	rows := map[string]*bankSummaryRow{}
	var order []string
	getRow := func(bank string) *bankSummaryRow {
		if r, ok := rows[bank]; ok {
			return r
		}
		r := &bankSummaryRow{Bank: bank}
		rows[bank] = r
		order = append(order, bank)
		return r
	}

	// Pre-seed a zero row for every bank TYPE (brand) in the lookup table —
	// regardless of whether a CompanyBank account currently exists for it —
	// so every configured bank brand always shows a row, even at $0.00,
	// rather than only banks that happen to already have a CompanyBank
	// account or existing transaction activity.
	for _, bt := range bankTypes {
		label := bt.Code
		if label == "" {
			label = bt.Name
		}
		if label == "" {
			continue
		}
		getRow(label)
	}

	for i := range deposits {
		d := &deposits[i]
		row := getRow(bankLabel(d.CompanyBankID, companyBankByID, bankTypeByID))
		if d.Currency == "KHR" {
			row.Deposit.KHR += d.Amount
			row.Bonus.KHR += d.BonusAmount
		} else {
			row.Deposit.USD += d.Amount
			row.Bonus.USD += d.BonusAmount
		}
	}
	for i := range withdrawals {
		w := &withdrawals[i]
		row := getRow(bankLabel(w.CompanyBankID, companyBankByID, bankTypeByID))
		if w.Currency == "KHR" {
			row.Withdrawal.KHR += w.Amount
			row.Bonus.KHR += w.BonusAmount
		} else {
			row.Withdrawal.USD += w.Amount
			row.Bonus.USD += w.BonusAmount
		}
	}

	sort.Strings(order)

	resultRows := make([]bankSummaryRow, 0, len(order))
	grand := bankSummaryRow{Bank: "Grand Total"}
	for _, name := range order {
		r := rows[name]
		// NOTE: Total = Deposit - Withdrawal (net flow through this bank).
		// Swap to `r.Deposit.USD + r.Withdrawal.USD` etc. if you actually
		// want gross volume instead.
		r.Total.USD = utils.RoundFloat(r.Deposit.USD-r.Withdrawal.USD, 2)
		r.Total.KHR = utils.RoundFloat(r.Deposit.KHR-r.Withdrawal.KHR, 2)
		r.Deposit.USD = utils.RoundFloat(r.Deposit.USD, 2)
		r.Deposit.KHR = utils.RoundFloat(r.Deposit.KHR, 2)
		r.Withdrawal.USD = utils.RoundFloat(r.Withdrawal.USD, 2)
		r.Withdrawal.KHR = utils.RoundFloat(r.Withdrawal.KHR, 2)
		r.Bonus.USD = utils.RoundFloat(r.Bonus.USD, 2)
		r.Bonus.KHR = utils.RoundFloat(r.Bonus.KHR, 2)

		resultRows = append(resultRows, *r)
		grand.Deposit.USD += r.Deposit.USD
		grand.Deposit.KHR += r.Deposit.KHR
		grand.Withdrawal.USD += r.Withdrawal.USD
		grand.Withdrawal.KHR += r.Withdrawal.KHR
		grand.Bonus.USD += r.Bonus.USD
		grand.Bonus.KHR += r.Bonus.KHR
	}
	grand.Deposit.USD = utils.RoundFloat(grand.Deposit.USD, 2)
	grand.Deposit.KHR = utils.RoundFloat(grand.Deposit.KHR, 2)
	grand.Withdrawal.USD = utils.RoundFloat(grand.Withdrawal.USD, 2)
	grand.Withdrawal.KHR = utils.RoundFloat(grand.Withdrawal.KHR, 2)
	grand.Bonus.USD = utils.RoundFloat(grand.Bonus.USD, 2)
	grand.Bonus.KHR = utils.RoundFloat(grand.Bonus.KHR, 2)
	grand.Total.USD = utils.RoundFloat(grand.Deposit.USD-grand.Withdrawal.USD, 2)
	grand.Total.KHR = utils.RoundFloat(grand.Deposit.KHR-grand.Withdrawal.KHR, 2)

	utils.OK(c, "success", gin.H{
		"rows":        resultRows,
		"grand_total": grand,
		// Convenience alias matching the standalone "Total USD/KHR" block
		// at the bottom of the sheet — identical to grand_total.deposit.
		"overall_total": grand.Deposit,
	})
}
