package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	transactiondto "crm-backend/dto/transaction"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type DepositService interface {
	Create(createdByID uint, req transactiondto.CreateRequest) (*models.Deposit, error)
	GetByID(id uint, scopeIDs []uint) (*models.Deposit, error)
	Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest, updatedByID uint) (*models.Deposit, error)
	Delete(id uint, scopeIDs []uint, deletedByID uint) error
	List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Deposit, int64, error)
	GetBalance(clientID, clientProductID uint) (*transactiondto.BalanceResponse, error)
	Approve(id uint, approvedByID uint, req transactiondto.ApproveRequest) (*models.Deposit, error)
}

type depositService struct {
	repo                  repositories.DepositRepository
	clientRepo            repositories.ClientRepository
	companyBankRepo       repositories.CompanyBankRepository
	productTypeRepo       repositories.ProductTypeRepository
	dailyStartBalanceRepo repositories.DailyStartBalanceRepository
	db                    *gorm.DB
}

func NewDepositService(
	repo repositories.DepositRepository,
	clientRepo repositories.ClientRepository,
	companyBankRepo repositories.CompanyBankRepository,
	productTypeRepo repositories.ProductTypeRepository,
	dailyStartBalanceRepo repositories.DailyStartBalanceRepository,
	db *gorm.DB,
) DepositService {
	return &depositService{repo, clientRepo, companyBankRepo, productTypeRepo, dailyStartBalanceRepo, db}
}

// productTypeIDFor resolves a client_product_id (a specific client's
// product/account instance) to its parent ProductType ID (the shared
// product category that actually holds the Credit pool).
func (s *depositService) productTypeIDFor(clientProductID uint) (uint, error) {
	cp, err := s.clientRepo.FindProduct(clientProductID)
	if err != nil {
		return 0, errors.New("client product not found")
	}
	return cp.ProductTypeID, nil
}

// productCurrency returns the ProductType's own currency code, defaulting
// to USD if the product type has no currency set. Used to convert a
// deposit/withdrawal amount into the product's own currency when they
// differ, via utils.ConvertCurrency.
func (s *depositService) productCurrency(productTypeID uint) string {
	pt, err := s.productTypeRepo.FindByID(productTypeID, nil)
	if err != nil || pt.CurrencyType == nil || pt.CurrencyType.Code == "" {
		return "USD"
	}
	return pt.CurrencyType.Code
}

// requireOpenShift blocks Create when there's no currently-open Daily
// Start Balance shift (Opening Cash/Opening Credit) for the given branch —
// staff must click "Start Shift" on the Daily Balance page before any
// deposit/withdrawal can be processed for that branch. Deposits/
// withdrawals with no branch set at all skip this check entirely, since
// there's no branch context to look a shift up against.
func requireOpenShift(repo repositories.DailyStartBalanceRepository, branchID *uint) error {
	if branchID == nil || *branchID == 0 {
		return nil
	}
	if _, err := repo.FindOpenByBranch(*branchID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("no shift is open for this branch yet — please Start Shift on the Daily Balance page before processing deposits or withdrawals")
		}
		return err
	}
	return nil
}

// requireNotLockedByClosedShift blocks Update when the transaction's own
// Date falls before the most recently closed shift's ClosedAt for its
// branch — once a shift is closed, everything dated before that close is
// considered reconciled and locked from further edits. The check uses the
// transaction's ORIGINAL stored date (not any new date being requested),
// since it's asking "was this transaction already accounted for in a
// closed shift?", not "would the new date fall in one". Deposits/
// withdrawals with no branch set at all skip this check entirely, matching
// requireOpenShift's behavior.
func requireNotLockedByClosedShift(repo repositories.DailyStartBalanceRepository, branchID *uint, txDate time.Time) error {
	if branchID == nil || *branchID == 0 {
		return nil
	}
	latest, err := repo.FindLatestClosedByBranch(*branchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // this branch has never closed a shift — nothing is locked yet
		}
		return err
	}
	if latest.ClosedAt != nil && txDate.Before(*latest.ClosedAt) {
		return errors.New("this transaction can't be edited — it's dated before the most recent shift close (" +
			latest.ClosedAt.Format("2006-01-02 15:04") + ") and is considered reconciled")
	}
	return nil
}

func (s *depositService) Create(createdByID uint, req transactiondto.CreateRequest) (*models.Deposit, error) {
	if err := requireOpenShift(s.dailyStartBalanceRepo, req.BranchID); err != nil {
		return nil, err
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	txNo := req.TransactionNo
	if txNo == "" {
		if req.BranchID != nil && *req.BranchID != 0 {
			txNo = utils.GenerateTxCodeForBranch(s.db, *req.BranchID, utils.EntityDeposit)
		} else {
			txNo = utils.GenerateCode(s.db, createdByID, utils.EntityDeposit)
		}
	}

	bonusAmount := req.BonusAmount

	// Bal and OS are stored exactly as given — no auto-calculation,
	// matching how Update already treats them (a simple direct input with
	// no cross-field side effects).
	bal := utils.RoundFloat(req.Bal, 2)
	os := utils.RoundFloat(req.OS, 2)

	productTypeID, err := s.productTypeIDFor(req.ClientProductID)
	if err != nil {
		return nil, err
	}
	productCurrency := s.productCurrency(productTypeID)
	// Bonus is a promotional credit funded entirely from the product's
	// shared credit pool — it draws down credit the same way the deposit
	// amount does, but (unlike the amount) is never real cash, so it must
	// NOT be added to the company bank's cash top-up below.
	creditDelta := utils.ConvertCurrency(req.Amount+bonusAmount, currency, productCurrency)
	// Converted separately (not just carved out of creditDelta) so the
	// ledger's BonusAmount is recorded in the SAME currency as Amount —
	// otherwise a deposit in one currency against a product priced in
	// another would record a bonus figure on the wrong currency scale.
	bonusCreditDelta := utils.ConvertCurrency(bonusAmount, currency, productCurrency)

	deposit := &models.Deposit{
		TransactionNo:   txNo,
		Date:            req.Date.Time,
		ClientID:        req.ClientID,
		ClientProductID: req.ClientProductID,
		ClientBankID:    req.ClientBankID,
		CompanyBankID:   req.CompanyBankID,
		Amount:          req.Amount,
		BonusAmount:     bonusAmount,
		BonusOptionID:   req.BonusOptionID,
		Bal:             bal,
		TO:              req.TO,
		OS:              os,
		Play:            req.Play,
		Currency:        currency,
		Remark:          req.Remark,
		BranchID:        req.BranchID,
		CreatedByID:     createdByID,
	}

	// Everything below runs in one DB transaction: the deposit row, the
	// company bank cash top-up, and the product credit draw-down all
	// commit together or all roll back together. TopUpCash/WithdrawCredit
	// each open their own internal transaction, but since we pass them a
	// repository bound to `tx` (not s.db), GORM nests those as SAVEPOINTs
	// under this outer transaction rather than as independent commits.
	//
	// TODO(business rule to confirm): this applies the cash/credit effect
	// immediately at Create time, not gated on Approve. If a deposit is
	// later rejected via Approve(status="rejected"), these balance changes
	// are NOT currently reversed. Let me know if rejection should trigger
	// an automatic reversal — it's a contained addition on top of this.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txDepositRepo := repositories.NewDepositRepository(tx)
		txCompanyBankRepo := repositories.NewCompanyBankRepository(tx)
		txProductTypeRepo := repositories.NewProductTypeRepository(tx)

		if err := txDepositRepo.Create(deposit); err != nil {
			return err
		}
		remark := "Deposit " + txNo
		// Deposit = client puts real money into the company's bank account.
		// Amount can now be 0 (a bonus-only deposit) — skip the cash
		// top-up entirely in that case, since TopUpCash rejects a
		// non-positive amount.
		if req.Amount > 0 {
			if _, err := txCompanyBankRepo.TopUpCash(req.CompanyBankID, req.Amount, remark, createdByID, models.BalanceSourceTransaction); err != nil {
				return err
			}
		}
		// Deposit (plus any bonus) draws down the product's shared credit
		// pool — skip if there's nothing to draw down (amount and bonus
		// both 0).
		if creditDelta > 0 {
			if _, err := txProductTypeRepo.WithdrawCredit(productTypeID, creditDelta, remark, createdByID, models.BalanceSourceTransaction, bonusCreditDelta); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(deposit.ID)
}

func (s *depositService) GetByID(id uint, scopeIDs []uint) (*models.Deposit, error) {
	d, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("deposit not found")
		}
		return nil, err
	}
	return d, nil
}

func (s *depositService) Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest, updatedByID uint) (*models.Deposit, error) {
	d, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("deposit not found")
	}

	if err := requireNotLockedByClosedShift(s.dailyStartBalanceRepo, d.BranchID, d.Date); err != nil {
		return nil, err
	}

	oldAmount := d.Amount
	oldBonusAmount := d.BonusAmount
	oldCompanyBankID := d.CompanyBankID
	newAmount := oldAmount
	newBonusAmount := oldBonusAmount
	newCompanyBankID := oldCompanyBankID

	if req.Date != nil {
		d.Date = req.Date.Time
	}
	if req.ClientBankID != nil {
		d.ClientBankID = *req.ClientBankID
	}
	if req.CompanyBankID != nil {
		newCompanyBankID = *req.CompanyBankID
		d.CompanyBankID = newCompanyBankID
	}
	if req.Amount != nil {
		newAmount = *req.Amount
		d.Amount = newAmount
	}
	if req.TO != nil {
		d.TO = *req.TO
	}
	if req.Play != nil {
		d.Play = *req.Play
	}
	if req.Remark != nil {
		d.Remark = *req.Remark
	}

	// A real 0 means "clear the bonus option" (the frontend now sends 0
	// instead of null when the select is cleared, since a *uint can't
	// otherwise distinguish "field omitted" from "field sent as null" -
	// both unmarshal to nil). An omitted field (nil pointer) leaves the
	// existing value untouched. Bonus amount is a separate, independent
	// manual field with no cross-field side effects.
	if req.BonusOptionID != nil {
		if *req.BonusOptionID == 0 {
			d.BonusOptionID = nil
		} else {
			d.BonusOptionID = req.BonusOptionID
		}
	}
	if req.BonusAmount != nil {
		newBonusAmount = *req.BonusAmount
		d.BonusAmount = newBonusAmount
	}

	// Bal and OS are stored exactly as given - simple direct input, same as
	// every other manual field. No auto-recalculation.
	if req.Bal != nil {
		d.Bal = *req.Bal
	}
	if req.OS != nil {
		d.OS = *req.OS
	}

	productTypeID, err := s.productTypeIDFor(d.ClientProductID)
	if err != nil {
		return nil, err
	}
	productCurrency := s.productCurrency(productTypeID)

	amountChanged := newAmount != oldAmount
	bankChanged := newCompanyBankID != oldCompanyBankID
	bonusChanged := newBonusAmount != oldBonusAmount

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txDepositRepo := repositories.NewDepositRepository(tx)
		txCompanyBankRepo := repositories.NewCompanyBankRepository(tx)
		txProductTypeRepo := repositories.NewProductTypeRepository(tx)

		if amountChanged || bankChanged || bonusChanged {
			// Never mutate a past ledger entry — post a reversal of the OLD
			// effect, then apply the NEW effect as its own entry. This keeps
			// the BalanceTransaction history an honest record of what
			// actually happened, rather than silently editing history.
			//
			// Cash only ever reflects the deposit AMOUNT (never bonus, since
			// bonus isn't real money). Credit reflects amount + bonus
			// together, since bonus draws down the product's credit pool
			// the same way the amount does. Both old and new amounts can
			// now be 0 (a bonus-only deposit), so every call below is
			// guarded — WithdrawCash/TopUpCash/TopUpCredit/WithdrawCredit
			// all reject a non-positive amount.
			remark := "Deposit edited " + d.TransactionNo
			oldCreditDelta := utils.ConvertCurrency(oldAmount+oldBonusAmount, d.Currency, productCurrency)
			// Same rationale as Create: convert the bonus portion
			// separately so it's recorded in the ledger on the same
			// currency scale as Amount, not the deposit's original
			// currency.
			oldBonusCreditDelta := utils.ConvertCurrency(oldBonusAmount, d.Currency, productCurrency)
			if oldAmount > 0 {
				if _, err := txCompanyBankRepo.WithdrawCash(oldCompanyBankID, oldAmount, remark+" (reversal)", updatedByID, models.BalanceSourceTransaction); err != nil {
					return err
				}
			}
			if oldCreditDelta > 0 {
				if _, err := txProductTypeRepo.TopUpCredit(productTypeID, oldCreditDelta, remark+" (reversal)", updatedByID, models.BalanceSourceTransaction, oldBonusCreditDelta); err != nil {
					return err
				}
			}

			newCreditDelta := utils.ConvertCurrency(newAmount+newBonusAmount, d.Currency, productCurrency)
			newBonusCreditDelta := utils.ConvertCurrency(newBonusAmount, d.Currency, productCurrency)
			if newAmount > 0 {
				if _, err := txCompanyBankRepo.TopUpCash(newCompanyBankID, newAmount, remark, updatedByID, models.BalanceSourceTransaction); err != nil {
					return err
				}
			}
			if newCreditDelta > 0 {
				if _, err := txProductTypeRepo.WithdrawCredit(productTypeID, newCreditDelta, remark, updatedByID, models.BalanceSourceTransaction, newBonusCreditDelta); err != nil {
					return err
				}
			}
		}
		return txDepositRepo.Update(d)
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *depositService) Delete(id uint, scopeIDs []uint, deletedByID uint) error {
	d, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return errors.New("deposit not found")
	}

	productTypeID, err := s.productTypeIDFor(d.ClientProductID)
	if err != nil {
		return err
	}
	productCurrency := s.productCurrency(productTypeID)
	creditDelta := utils.ConvertCurrency(d.Amount+d.BonusAmount, d.Currency, productCurrency)
	bonusCreditDelta := utils.ConvertCurrency(d.BonusAmount, d.Currency, productCurrency)

	return s.db.Transaction(func(tx *gorm.DB) error {
		txDepositRepo := repositories.NewDepositRepository(tx)
		txCompanyBankRepo := repositories.NewCompanyBankRepository(tx)
		txProductTypeRepo := repositories.NewProductTypeRepository(tx)

		remark := "Deposit deleted " + d.TransactionNo
		if d.Amount > 0 {
			if _, err := txCompanyBankRepo.WithdrawCash(d.CompanyBankID, d.Amount, remark, deletedByID, models.BalanceSourceTransaction); err != nil {
				return err
			}
		}
		if creditDelta > 0 {
			if _, err := txProductTypeRepo.TopUpCredit(productTypeID, creditDelta, remark, deletedByID, models.BalanceSourceTransaction, bonusCreditDelta); err != nil {
				return err
			}
		}
		return txDepositRepo.Delete(id)
	})
}

func (s *depositService) List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Deposit, int64, error) {
	return s.repo.List(filter, p, userID)
}

func (s *depositService) GetBalance(clientID, clientProductID uint) (*transactiondto.BalanceResponse, error) {
	totalDep, err := s.repo.SumDeposits(clientID, clientProductID)
	if err != nil {
		return nil, err
	}
	totalWdr, err := s.repo.SumWithdrawals(clientID, clientProductID)
	if err != nil {
		return nil, err
	}
	return &transactiondto.BalanceResponse{
		ClientID: clientID, ClientProductID: clientProductID, Currency: "USD",
		TotalDeposits:    utils.RoundFloat(totalDep, 2),
		TotalWithdrawals: utils.RoundFloat(totalWdr, 2),
		CurrentBalance:   utils.RoundFloat(totalDep-totalWdr, 2),
	}, nil
}

func (s *depositService) Approve(id uint, approvedByID uint, req transactiondto.ApproveRequest) (*models.Deposit, error) {
	d, err := s.repo.FindByIDUnsafe(id)
	if err != nil {
		return nil, errors.New("deposit not found")
	}
	now := time.Now()
	d.Status = models.TransactionStatus(req.Status)
	d.ApprovedAt = &now
	d.ApprovedByID = &approvedByID
	if err := s.repo.Update(d); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}
