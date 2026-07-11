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

type WithdrawalService interface {
	Create(createdByID uint, req transactiondto.CreateRequest) (*models.Withdrawal, error)
	GetByID(id uint, scopeIDs []uint) (*models.Withdrawal, error)
	Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest, updatedByID uint) (*models.Withdrawal, error)
	Delete(id uint, scopeIDs []uint, deletedByID uint) error
	List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Withdrawal, int64, error)
	GetBalance(clientID, clientProductID uint) (*transactiondto.BalanceResponse, error)
	Approve(id uint, approvedByID uint, req transactiondto.ApproveRequest) (*models.Withdrawal, error)
}

type withdrawalService struct {
	repo                  repositories.WithdrawalRepository
	clientRepo            repositories.ClientRepository
	companyBankRepo       repositories.CompanyBankRepository
	productTypeRepo       repositories.ProductTypeRepository
	dailyStartBalanceRepo repositories.DailyStartBalanceRepository
	db                    *gorm.DB
}

func NewWithdrawalService(
	repo repositories.WithdrawalRepository,
	clientRepo repositories.ClientRepository,
	companyBankRepo repositories.CompanyBankRepository,
	productTypeRepo repositories.ProductTypeRepository,
	dailyStartBalanceRepo repositories.DailyStartBalanceRepository,
	db *gorm.DB,
) WithdrawalService {
	return &withdrawalService{repo, clientRepo, companyBankRepo, productTypeRepo, dailyStartBalanceRepo, db}
}

// productTypeIDFor resolves a client_product_id to its parent ProductType ID
// (the shared product category that actually holds the Credit pool).
func (s *withdrawalService) productTypeIDFor(clientProductID uint) (uint, error) {
	cp, err := s.clientRepo.FindProduct(clientProductID)
	if err != nil {
		return 0, errors.New("client product not found")
	}
	return cp.ProductTypeID, nil
}

// productCurrency returns the ProductType's own currency code, defaulting
// to USD if unset. Used to convert a withdrawal amount into the product's
// own currency when they differ, via utils.ConvertCurrency.
func (s *withdrawalService) productCurrency(productTypeID uint) string {
	pt, err := s.productTypeRepo.FindByID(productTypeID, nil)
	if err != nil || pt.CurrencyType == nil || pt.CurrencyType.Code == "" {
		return "USD"
	}
	return pt.CurrencyType.Code
}

func (s *withdrawalService) Create(createdByID uint, req transactiondto.CreateRequest) (*models.Withdrawal, error) {
	if err := requireOpenShift(s.dailyStartBalanceRepo, req.BranchID); err != nil {
		return nil, err
	}
	// CreateRequest.Amount's own binding was loosened to allow 0 for
	// bonus-only Deposits — that doesn't apply to Withdrawals, which can
	// never withdraw a zero amount, so it's enforced here instead.
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	txNo := req.TransactionNo
	if txNo == "" {
		if req.BranchID != nil && *req.BranchID != 0 {
			txNo = utils.GenerateTxCodeForBranch(s.db, *req.BranchID, utils.EntityWithdrawal)
		} else {
			txNo = utils.GenerateCode(s.db, createdByID, utils.EntityWithdrawal)
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
	creditDelta := utils.ConvertCurrency(req.Amount, currency, productCurrency)

	withdrawal := &models.Withdrawal{
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

	// See depositService.Create for the transaction/nesting rationale — same
	// pattern here, just the reverse direction:
	//   Withdrawal = money paid out of the company bank (cash decreases,
	//   blocked if insufficient) and the product's credit pool is
	//   replenished (credit increases) since the credit that funded this
	//   client's balance is being settled/returned.
	//
	// TODO(business rule to confirm): same immediate-at-Create caveat as
	// deposits — a later-rejected withdrawal does not currently reverse
	// this cash/credit effect. See deposit_service.go's Create for details.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txWithdrawalRepo := repositories.NewWithdrawalRepository(tx)
		txCompanyBankRepo := repositories.NewCompanyBankRepository(tx)
		txProductTypeRepo := repositories.NewProductTypeRepository(tx)

		if err := txWithdrawalRepo.Create(withdrawal); err != nil {
			return err
		}
		remark := "Withdrawal " + txNo
		if _, err := txCompanyBankRepo.WithdrawCash(req.CompanyBankID, req.Amount, remark, createdByID, models.BalanceSourceTransaction); err != nil {
			return err
		}
		if _, err := txProductTypeRepo.TopUpCredit(productTypeID, creditDelta, remark, createdByID, models.BalanceSourceTransaction, 0); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(withdrawal.ID)
}

func (s *withdrawalService) GetByID(id uint, scopeIDs []uint) (*models.Withdrawal, error) {
	w, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("withdrawal not found")
		}
		return nil, err
	}
	return w, nil
}

func (s *withdrawalService) Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest, updatedByID uint) (*models.Withdrawal, error) {
	w, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("withdrawal not found")
	}

	if err := requireNotLockedByClosedShift(s.dailyStartBalanceRepo, w.BranchID, w.Date); err != nil {
		return nil, err
	}

	oldAmount := w.Amount
	oldCompanyBankID := w.CompanyBankID
	newAmount := oldAmount
	newCompanyBankID := oldCompanyBankID

	if req.Date != nil {
		w.Date = req.Date.Time
	}
	if req.ClientBankID != nil {
		w.ClientBankID = *req.ClientBankID
	}
	if req.CompanyBankID != nil {
		newCompanyBankID = *req.CompanyBankID
		w.CompanyBankID = newCompanyBankID
	}
	if req.Amount != nil {
		newAmount = *req.Amount
		w.Amount = newAmount
	}
	if req.TO != nil {
		w.TO = *req.TO
	}
	if req.Play != nil {
		w.Play = *req.Play
	}
	if req.Remark != nil {
		w.Remark = *req.Remark
	}

	// A real 0 means "clear the bonus option" (the frontend sends 0 instead
	// of null when the select is cleared, since a *uint can't otherwise
	// distinguish "field omitted" from "field sent as null" - both
	// unmarshal to nil). An omitted field (nil pointer) leaves the existing
	// value untouched. Bonus amount is a separate, independent manual field
	// with no cross-field side effects.
	if req.BonusOptionID != nil {
		if *req.BonusOptionID == 0 {
			w.BonusOptionID = nil
		} else {
			w.BonusOptionID = req.BonusOptionID
		}
	}
	if req.BonusAmount != nil {
		w.BonusAmount = *req.BonusAmount
	}

	// Bal and OS are stored exactly as given - simple direct input, same as
	// every other manual field. No auto-recalculation.
	if req.Bal != nil {
		w.Bal = *req.Bal
	}
	if req.OS != nil {
		w.OS = *req.OS
	}

	productTypeID, err := s.productTypeIDFor(w.ClientProductID)
	if err != nil {
		return nil, err
	}
	productCurrency := s.productCurrency(productTypeID)

	amountChanged := newAmount != oldAmount
	bankChanged := newCompanyBankID != oldCompanyBankID

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txWithdrawalRepo := repositories.NewWithdrawalRepository(tx)
		txCompanyBankRepo := repositories.NewCompanyBankRepository(tx)
		txProductTypeRepo := repositories.NewProductTypeRepository(tx)

		if amountChanged || bankChanged {
			// Post a reversal of the OLD effect, then apply the NEW effect
			// as its own entry — never mutate a past ledger row.
			remark := "Withdrawal edited " + w.TransactionNo
			oldCreditDelta := utils.ConvertCurrency(oldAmount, w.Currency, productCurrency)
			if _, err := txCompanyBankRepo.TopUpCash(oldCompanyBankID, oldAmount, remark+" (reversal)", updatedByID, models.BalanceSourceTransaction); err != nil {
				return err
			}
			if _, err := txProductTypeRepo.WithdrawCredit(productTypeID, oldCreditDelta, remark+" (reversal)", updatedByID, models.BalanceSourceTransaction, 0); err != nil {
				return err
			}

			newCreditDelta := utils.ConvertCurrency(newAmount, w.Currency, productCurrency)
			if _, err := txCompanyBankRepo.WithdrawCash(newCompanyBankID, newAmount, remark, updatedByID, models.BalanceSourceTransaction); err != nil {
				return err
			}
			if _, err := txProductTypeRepo.TopUpCredit(productTypeID, newCreditDelta, remark, updatedByID, models.BalanceSourceTransaction, 0); err != nil {
				return err
			}
		}
		return txWithdrawalRepo.Update(w)
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *withdrawalService) Delete(id uint, scopeIDs []uint, deletedByID uint) error {
	w, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return errors.New("withdrawal not found")
	}

	productTypeID, err := s.productTypeIDFor(w.ClientProductID)
	if err != nil {
		return err
	}
	productCurrency := s.productCurrency(productTypeID)
	creditDelta := utils.ConvertCurrency(w.Amount, w.Currency, productCurrency)

	return s.db.Transaction(func(tx *gorm.DB) error {
		txWithdrawalRepo := repositories.NewWithdrawalRepository(tx)
		txCompanyBankRepo := repositories.NewCompanyBankRepository(tx)
		txProductTypeRepo := repositories.NewProductTypeRepository(tx)

		remark := "Withdrawal deleted " + w.TransactionNo
		if _, err := txCompanyBankRepo.TopUpCash(w.CompanyBankID, w.Amount, remark, deletedByID, models.BalanceSourceTransaction); err != nil {
			return err
		}
		if _, err := txProductTypeRepo.WithdrawCredit(productTypeID, creditDelta, remark, deletedByID, models.BalanceSourceTransaction, 0); err != nil {
			return err
		}
		return txWithdrawalRepo.Delete(id)
	})
}

func (s *withdrawalService) List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Withdrawal, int64, error) {
	return s.repo.List(filter, p, userID)
}

func (s *withdrawalService) GetBalance(clientID, clientProductID uint) (*transactiondto.BalanceResponse, error) {
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

func (s *withdrawalService) Approve(id uint, approvedByID uint, req transactiondto.ApproveRequest) (*models.Withdrawal, error) {
	w, err := s.repo.FindByIDUnsafe(id)
	if err != nil {
		return nil, errors.New("withdrawal not found")
	}
	now := time.Now()
	w.Status = models.TransactionStatus(req.Status)
	w.ApprovedAt = &now
	w.ApprovedByID = &approvedByID
	if err := s.repo.Update(w); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}
