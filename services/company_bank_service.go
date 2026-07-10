package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

type CompanyBankService interface {
	Create(createdByID uint, x *models.CompanyBank) (*models.CompanyBank, error)
	ListForUser(userID uint, showAll bool, branchID *uint) ([]models.CompanyBank, error)
	List(showAll bool) ([]models.CompanyBank, error)
	// GetByID, Update, and Delete take the caller's branch scopeIDs (nil
	// means unscoped / super admin access) so access can be enforced the
	// same way ListForUser does.
	GetByID(id uint, scopeIDs []uint) (*models.CompanyBank, error)
	Update(id uint, scopeIDs []uint, x *models.CompanyBank) (*models.CompanyBank, error)
	Delete(id uint, scopeIDs []uint) error
	TopUpCash(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error)
	WithdrawCash(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error)
	// AdjustCash applies a manual correction — direction is "addition" or
	// "subtraction" — recorded under BalanceSourceAdjustment so it's
	// always clearly distinguishable from a routine Top Up/Withdraw in
	// the ledger.
	AdjustCash(id uint, scopeIDs []uint, direction string, amount float64, remark string, createdByID uint) (*models.CompanyBank, error)
}

type companyBankService struct {
	repo repositories.CompanyBankRepository
}

func NewCompanyBankService(repo repositories.CompanyBankRepository) CompanyBankService {
	return &companyBankService{repo}
}

func (s *companyBankService) Create(createdByID uint, x *models.CompanyBank) (*models.CompanyBank, error) {
	x.IsActive = true
	if err := s.repo.Create(x, createdByID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(x.ID, nil)
}

func (s *companyBankService) List(showAll bool) ([]models.CompanyBank, error) {
	return s.repo.List(showAll)
}

func (s *companyBankService) GetByID(id uint, scopeIDs []uint) (*models.CompanyBank, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("company bank not found")
		}
		return nil, err
	}
	return x, nil
}

func (s *companyBankService) Update(id uint, scopeIDs []uint, upd *models.CompanyBank) (*models.CompanyBank, error) {
	x, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("company bank not found")
	}

	// Only copy the fields a client is allowed to change. This preserves
	// ID, CreatedByID, CreatedAt, and any other bookkeeping fields that a
	// full struct overwrite would otherwise wipe out.
	//
	// NOTE: every field the frontend form sends must be listed here — the
	// BankType/ProductType Update() methods silently dropped Code and
	// Description for a while because they were left off this whitelist.
	x.BankTypeID = upd.BankTypeID
	x.AccountNumber = upd.AccountNumber
	x.AccountName = upd.AccountName
	x.CurrencyTypeID = upd.CurrencyTypeID
	x.QRCodeURL = upd.QRCodeURL
	x.BranchID = upd.BranchID
	x.SortOrder = upd.SortOrder
	x.IsActive = upd.IsActive

	if err := s.repo.Update(x); err != nil {
		return nil, err
	}
	return x, nil
}

func (s *companyBankService) TopUpCash(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	// Scope check first: confirms the caller can even see this account
	// before we touch the balance/ledger.
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return nil, errors.New("company bank not found")
	}
	// This service method is only ever reached via the direct admin Top
	// Up/Withdraw endpoint (Configuration module) — never as a side effect
	// of a client Deposit/Withdrawal, which calls the repository directly
	// from deposit_service.go/withdrawal_service.go instead.
	return s.repo.TopUpCash(id, amount, remark, createdByID, models.BalanceSourceConfiguration)
}

func (s *companyBankService) WithdrawCash(id uint, scopeIDs []uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return nil, errors.New("company bank not found")
	}
	return s.repo.WithdrawCash(id, amount, remark, createdByID, models.BalanceSourceConfiguration)
}

func (s *companyBankService) AdjustCash(id uint, scopeIDs []uint, direction string, amount float64, remark string, createdByID uint) (*models.CompanyBank, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return nil, errors.New("company bank not found")
	}
	switch direction {
	case "addition":
		return s.repo.TopUpCash(id, amount, remark, createdByID, models.BalanceSourceAdjustment)
	case "subtraction":
		return s.repo.WithdrawCash(id, amount, remark, createdByID, models.BalanceSourceAdjustment)
	default:
		return nil, errors.New("direction must be \"addition\" or \"subtraction\"")
	}
}

func (s *companyBankService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("company bank not found")
	}
	return s.repo.Delete(id, scopeIDs)
}

func (s *companyBankService) ListForUser(userID uint, showAll bool, branchID *uint) ([]models.CompanyBank, error) {
	return s.repo.ListForUser(userID, showAll, branchID)
}
