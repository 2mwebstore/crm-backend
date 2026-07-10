package services

import (
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

// cambodiaLoc is used for every date/time this feature records or
// computes ("today", CreatedAt, ClosedAt) — matching how Deposit dates are
// handled elsewhere in this system. Falls back to a fixed UTC+7 offset
// (Cambodia has no DST, so this is equivalent) if the IANA tzdata isn't
// available in the deployment environment for some reason.
var cambodiaLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Phnom_Penh")
	if err != nil {
		return time.FixedZone("ICT", 7*3600)
	}
	return loc
}()

func nowInCambodia() time.Time {
	return time.Now().In(cambodiaLoc)
}

// safeUserName reads a preloaded *models.User's Name without panicking if
// the relation wasn't loaded for some reason.
func safeUserName(u *models.User) string {
	if u == nil {
		return "someone"
	}
	return u.Name
}

// CurrencyAmount holds a USD/KHR pair for one total — used instead of a
// single blended float since Company Bank Cash and Product Credit can each
// be denominated in either currency per individual record.
type CurrencyAmount struct {
	USD float64 `json:"usd"`
	KHR float64 `json:"khr"`
}

// TodayBalanceResponse bundles today's live totals alongside this
// morning's snapshot (if one exists) and the resulting "income so far
// today" — the change since that snapshot was taken. Everything is split
// by currency.
type TodayBalanceResponse struct {
	Snapshot      *models.DailyStartBalance `json:"snapshot"`
	CurrentCash   CurrencyAmount            `json:"current_cash"`
	CurrentCredit CurrencyAmount            `json:"current_credit"`
	// nil when no snapshot has been started yet today.
	IncomeCash   *CurrencyAmount `json:"income_cash,omitempty"`
	IncomeCredit *CurrencyAmount `json:"income_credit,omitempty"`

	// Detail breakdown behind each total — every CompanyBank/ProductType
	// record that was actually summed, so the frontend can show exactly
	// what makes up the headline number instead of just a bare total.
	CompanyBanks []CompanyBankBalanceRow `json:"company_banks"`
	ProductTypes []ProductTypeBalanceRow `json:"product_types"`

	// IncomeTransactions is the actual list of Deposits/Withdrawals (date
	// >= this shift's own opening time) that make up Income Cash/Credit —
	// i.e. answers "where did this income actually come from", rather
	// than just showing the before/after balance diff. Empty/omitted
	// when no shift is currently open.
	IncomeTransactions []IncomeTransactionRow `json:"income_transactions,omitempty"`

	// BalanceTransactions is the raw ledger — every CompanyBank cash /
	// ProductType credit top-up or withdrawal recorded against this
	// branch's accounts since this shift opened. Distinct from
	// IncomeTransactions: this also captures manual admin top-ups/
	// withdrawals made directly on a Company Bank or Product Type record
	// (outside of any client deposit/withdrawal), so it's the complete
	// audit trail behind the Income totals, not just the client-facing
	// side of it.
	BalanceTransactions []models.BalanceTransaction `json:"balance_transactions,omitempty"`
}

// IncomeTransactionRow is one Deposit or Withdrawal that happened since
// the currently open shift's own CreatedAt.
type IncomeTransactionRow struct {
	Type          string    `json:"type"` // "deposit" | "withdrawal"
	TransactionNo string    `json:"transaction_no"`
	Date          time.Time `json:"date"`
	ClientName    string    `json:"client_name,omitempty"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
}

type CompanyBankBalanceRow struct {
	ID            uint    `json:"id"`
	AccountName   string  `json:"account_name"`
	AccountNumber string  `json:"account_number"`
	BankTypeName  string  `json:"bank_type_name,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	Cash          float64 `json:"cash"`
}

type ProductTypeBalanceRow struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Code     string  `json:"code,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Credit   float64 `json:"credit"`
}

type DailyStartBalanceService interface {
	// StartToday captures the opening snapshot for the given branch —
	// fails if the caller doesn't have access to that branch, or if a
	// shift is already open for it (must be closed before a new one can
	// start).
	StartToday(callerID uint, branchID uint) (*models.DailyStartBalance, error)
	// CloseToday captures the closing snapshot for the given branch's
	// currently open shift — fails if no shift is currently open for it.
	CloseToday(callerID uint, branchID uint) (*models.DailyStartBalance, error)
	GetToday(callerID uint, branchID uint) (*TodayBalanceResponse, error)
	History(callerID uint, branchID uint, page, pageSize int) ([]models.DailyStartBalance, int64, error)
	// GetShiftBalanceTransactions returns the ledger entries (topups/
	// withdrawals) for any single shift — open or already closed — used
	// for an on-demand "View Transactions" lookup from the History table,
	// without bloating the paginated History response itself.
	GetShiftBalanceTransactions(callerID uint, shiftID uint) ([]models.BalanceTransaction, error)
}

type dailyStartBalanceService struct {
	repo                   repositories.DailyStartBalanceRepository
	userRepo               repositories.UserRepository
	companyBankRepo        repositories.CompanyBankRepository
	productTypeRepo        repositories.ProductTypeRepository
	depositRepo            repositories.DepositRepository
	withdrawalRepo         repositories.WithdrawalRepository
	balanceTransactionRepo repositories.BalanceTransactionRepository
}

func NewDailyStartBalanceService(
	repo repositories.DailyStartBalanceRepository,
	userRepo repositories.UserRepository,
	companyBankRepo repositories.CompanyBankRepository,
	productTypeRepo repositories.ProductTypeRepository,
	depositRepo repositories.DepositRepository,
	withdrawalRepo repositories.WithdrawalRepository,
	balanceTransactionRepo repositories.BalanceTransactionRepository,
) DailyStartBalanceService {
	return &dailyStartBalanceService{repo, userRepo, companyBankRepo, productTypeRepo, depositRepo, withdrawalRepo, balanceTransactionRepo}
}

// checkBranchAccess confirms the caller can act on the given branch —
// Super Admin (and SA sub-users) bypass this entirely (GetUserBranchIDs
// returns nil for them); everyone else must have that branch directly
// assigned to them.
func (s *dailyStartBalanceService) checkBranchAccess(callerID, branchID uint) error {
	branchIDs, err := s.userRepo.GetUserBranchIDs(callerID)
	if err != nil {
		return err
	}
	if branchIDs == nil {
		return nil
	}
	for _, id := range branchIDs {
		if id == branchID {
			return nil
		}
	}
	return errors.New("you do not have access to this branch")
}

// currentTotals groups Cash across every CompanyBank and Credit across
// every ProductType visible to the caller AND scoped to this one specific
// branch (via ListForUser's branchID filter) by currency — USD and KHR
// are never summed together. Also returns the individual per-record detail
// rows behind those totals, each tagged with its own currency code. Any
// record without a recognized KHR currency is bucketed into USD by default
// (matches this system's existing "USD is the default/base currency"
// convention elsewhere).
func (s *dailyStartBalanceService) currentTotals(callerID, branchID uint) (cash CurrencyAmount, credit CurrencyAmount, bankRows []CompanyBankBalanceRow, productRows []ProductTypeBalanceRow, err error) {
	banks, err := s.companyBankRepo.ListForUser(callerID, true, &branchID)
	if err != nil {
		return cash, credit, nil, nil, err
	}
	bankRows = make([]CompanyBankBalanceRow, 0, len(banks))
	for _, b := range banks {
		currencyCode := ""
		if b.CurrencyType != nil {
			currencyCode = b.CurrencyType.Code
		}
		if currencyCode == "KHR" {
			cash.KHR += b.Cash
		} else {
			cash.USD += b.Cash
		}
		row := CompanyBankBalanceRow{
			ID:            b.ID,
			AccountName:   b.AccountName,
			AccountNumber: b.AccountNumber,
			Currency:      currencyCode,
			Cash:          b.Cash,
		}
		if b.BankType != nil {
			row.BankTypeName = b.BankType.Name
		}
		bankRows = append(bankRows, row)
	}

	products, err := s.productTypeRepo.ListForUser(callerID, true, &branchID)
	if err != nil {
		return cash, credit, nil, nil, err
	}
	productRows = make([]ProductTypeBalanceRow, 0, len(products))
	for _, p := range products {
		currencyCode := ""
		if p.CurrencyType != nil {
			currencyCode = p.CurrencyType.Code
		}
		if currencyCode == "KHR" {
			credit.KHR += p.Credit
		} else {
			credit.USD += p.Credit
		}
		productRows = append(productRows, ProductTypeBalanceRow{
			ID:       p.ID,
			Name:     p.Name,
			Code:     p.Code,
			Currency: currencyCode,
			Credit:   p.Credit,
		})
	}
	return cash, credit, bankRows, productRows, nil
}

func buildDetailRows(phase models.DailyBalancePhase, bankRows []CompanyBankBalanceRow, productRows []ProductTypeBalanceRow) []models.DailyStartBalanceDetail {
	details := make([]models.DailyStartBalanceDetail, 0, len(bankRows)+len(productRows))
	for _, b := range bankRows {
		details = append(details, models.DailyStartBalanceDetail{
			Phase:      phase,
			EntityType: models.DailyBalanceEntityCompanyBank,
			EntityID:   b.ID,
			Label:      b.AccountName,
			Currency:   b.Currency,
			Amount:     b.Cash,
		})
	}
	for _, p := range productRows {
		details = append(details, models.DailyStartBalanceDetail{
			Phase:      phase,
			EntityType: models.DailyBalanceEntityProductType,
			EntityID:   p.ID,
			Label:      p.Name,
			Currency:   p.Currency,
			Amount:     p.Credit,
		})
	}
	return details
}

func (s *dailyStartBalanceService) StartToday(callerID uint, branchID uint) (*models.DailyStartBalance, error) {
	if err := s.checkBranchAccess(callerID, branchID); err != nil {
		return nil, err
	}
	if existing, err := s.repo.FindOpenByBranch(branchID); err == nil {
		return nil, errors.New("a shift is already open for this branch (started at " + existing.CreatedAt.Format("15:04") + " by " + safeUserName(existing.CreatedBy) + ") — close it before starting a new one")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cash, credit, bankRows, productRows, err := s.currentTotals(callerID, branchID)
	if err != nil {
		return nil, err
	}

	now := nowInCambodia()
	snap := &models.DailyStartBalance{
		BranchID:    branchID,
		Date:        &now,
		CashUSD:     cash.USD,
		CashKHR:     cash.KHR,
		CreditUSD:   credit.USD,
		CreditKHR:   credit.KHR,
		CreatedByID: callerID,
		CreatedAt:   now,
	}
	details := buildDetailRows(models.DailyBalancePhaseOpen, bankRows, productRows)

	if err := s.repo.CreateWithDetails(snap, details); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *dailyStartBalanceService) CloseToday(callerID uint, branchID uint) (*models.DailyStartBalance, error) {
	if err := s.checkBranchAccess(callerID, branchID); err != nil {
		return nil, err
	}

	snap, err := s.repo.FindOpenByBranch(branchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no shift is currently open for this branch — start one before closing it")
		}
		return nil, err
	}
	if snap.ClosedAt != nil {
		// Shouldn't happen — FindOpenByBranch only ever returns open
		// shifts — but guard against it regardless.
		return nil, errors.New("this shift was already closed at " + snap.ClosedAt.Format("15:04"))
	}

	cash, credit, bankRows, productRows, err := s.currentTotals(callerID, branchID)
	if err != nil {
		return nil, err
	}

	now := nowInCambodia()
	closedByID := callerID
	snap.CloseCashUSD = &cash.USD
	snap.CloseCashKHR = &cash.KHR
	snap.CloseCreditUSD = &credit.USD
	snap.CloseCreditKHR = &credit.KHR
	snap.ClosedByID = &closedByID
	snap.ClosedAt = &now

	details := buildDetailRows(models.DailyBalancePhaseClose, bankRows, productRows)

	if err := s.repo.UpdateCloseWithDetails(snap, details); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *dailyStartBalanceService) GetToday(callerID uint, branchID uint) (*TodayBalanceResponse, error) {
	if err := s.checkBranchAccess(callerID, branchID); err != nil {
		return nil, err
	}
	cash, credit, bankRows, productRows, err := s.currentTotals(callerID, branchID)
	if err != nil {
		return nil, err
	}

	resp := &TodayBalanceResponse{
		CurrentCash:   cash,
		CurrentCredit: credit,
		CompanyBanks:  bankRows,
		ProductTypes:  productRows,
	}

	snap, err := s.repo.FindOpenByBranch(branchID)
	if err == nil {
		resp.Snapshot = snap
		// Income is always computed against THIS specific open shift's own
		// opening totals — if a new shift was opened later in the day
		// (after an earlier one closed), this correctly reflects only
		// what's happened since THIS shift's own open time, never
		// bleeding in an earlier shift's activity.
		resp.IncomeCash = &CurrencyAmount{
			USD: cash.USD - snap.CashUSD,
			KHR: cash.KHR - snap.CashKHR,
		}
		resp.IncomeCredit = &CurrencyAmount{
			USD: credit.USD - snap.CreditUSD,
			KHR: credit.KHR - snap.CreditKHR,
		}

		// Pull the actual Deposit/Withdrawal records that happened since
		// this shift's own opening time — this is the real answer to
		// "where did this income come from", not just a balance diff.
		resp.IncomeTransactions = s.incomeTransactionsSince(branchID, snap.CreatedAt)

		// Pull the raw CompanyBank/ProductType ledger entries for this
		// branch's accounts since this shift opened — bankRows/productRows
		// already give us exactly which entities belong to this branch.
		resp.BalanceTransactions = s.balanceTransactionsInRange(bankRows, productRows, snap.CreatedAt, nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// gorm.ErrRecordNotFound just means "no shift currently open" — leave
	// Snapshot/IncomeCash/IncomeCredit nil, that's expected, not an error.

	return resp, nil
}

func (s *dailyStartBalanceService) History(callerID uint, branchID uint, page, pageSize int) ([]models.DailyStartBalance, int64, error) {
	if err := s.checkBranchAccess(callerID, branchID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListByBranch(branchID, page, pageSize)
}

// incomeTransactionsSince fetches every deposit and withdrawal for this
// branch with date >= since, merges them into one list sorted oldest-
// first, and quietly returns an empty slice on error rather than failing
// the whole GetToday call — this is supplementary detail, not something
// that should block showing the rest of the response if it can't be
// fetched for some reason.
func (s *dailyStartBalanceService) incomeTransactionsSince(branchID uint, since time.Time) []IncomeTransactionRow {
	var rows []IncomeTransactionRow

	deposits, err := s.depositRepo.ListSinceForBranch(branchID, since)
	if err == nil {
		for _, d := range deposits {
			clientName := ""
			if d.Client != nil {
				clientName = d.Client.Name
			}
			rows = append(rows, IncomeTransactionRow{
				Type:          "deposit",
				TransactionNo: d.TransactionNo,
				Date:          d.Date,
				ClientName:    clientName,
				Amount:        d.Amount,
				Currency:      d.Currency,
			})
		}
	}

	withdrawals, err := s.withdrawalRepo.ListSinceForBranch(branchID, since)
	if err == nil {
		for _, w := range withdrawals {
			clientName := ""
			if w.Client != nil {
				clientName = w.Client.Name
			}
			rows = append(rows, IncomeTransactionRow{
				Type:          "withdrawal",
				TransactionNo: w.TransactionNo,
				Date:          w.Date,
				ClientName:    clientName,
				Amount:        w.Amount,
				Currency:      w.Currency,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Date.Before(rows[j].Date) })
	return rows
}

// balanceTransactionsInRange fetches every ledger entry for the given
// company banks + product types whose CreatedAt falls within [from, to]
// (to=nil means "up to now"), merged into one time-ordered list. Quietly
// returns an empty slice on error, same rationale as
// incomeTransactionsSince — this is supplementary audit detail, not
// something that should block the rest of the response.
func (s *dailyStartBalanceService) balanceTransactionsInRange(bankRows []CompanyBankBalanceRow, productRows []ProductTypeBalanceRow, from time.Time, to *time.Time) []models.BalanceTransaction {
	bankIDs := make([]uint, len(bankRows))
	for i, b := range bankRows {
		bankIDs[i] = b.ID
	}
	productIDs := make([]uint, len(productRows))
	for i, p := range productRows {
		productIDs[i] = p.ID
	}

	var all []models.BalanceTransaction
	if bankTx, err := s.balanceTransactionRepo.ListByEntitiesInRange(models.BalanceEntityCompanyBank, bankIDs, from, to); err == nil {
		all = append(all, bankTx...)
	}
	if productTx, err := s.balanceTransactionRepo.ListByEntitiesInRange(models.BalanceEntityProductType, productIDs, from, to); err == nil {
		all = append(all, productTx...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.Before(all[j].CreatedAt) })
	return all
}

// GetShiftBalanceTransactions looks up any single shift (open or closed)
// by ID, confirms the caller has access to its branch, and returns the
// ledger entries recorded against that branch's company banks/product
// types between the shift's own open and close (or "now" if still open).
//
// NOTE: this scopes to the branch's CURRENTLY configured company banks/
// product types (via ListForUser), same as the rest of this feature — if
// an account was added, removed, or deactivated after this shift closed,
// this reflects today's account list, not necessarily the exact set that
// existed at the time. Fine for the common case; flagging in case it ever
// matters for a heavily-restructured branch's older history.
func (s *dailyStartBalanceService) GetShiftBalanceTransactions(callerID uint, shiftID uint) ([]models.BalanceTransaction, error) {
	shift, err := s.repo.FindByID(shiftID)
	if err != nil {
		return nil, err
	}
	if err := s.checkBranchAccess(callerID, shift.BranchID); err != nil {
		return nil, err
	}

	_, _, bankRows, productRows, err := s.currentTotals(callerID, shift.BranchID)
	if err != nil {
		return nil, err
	}

	return s.balanceTransactionsInRange(bankRows, productRows, shift.CreatedAt, shift.ClosedAt), nil
}
