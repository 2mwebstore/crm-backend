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
	Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest) (*models.Withdrawal, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Withdrawal, int64, error)
	GetBalance(clientID, clientProductID uint) (*transactiondto.BalanceResponse, error)
	Approve(id uint, approvedByID uint, req transactiondto.ApproveRequest) (*models.Withdrawal, error)
}

type withdrawalService struct {
	repo repositories.WithdrawalRepository
	db   *gorm.DB
}

func NewWithdrawalService(repo repositories.WithdrawalRepository, db *gorm.DB) WithdrawalService {
	return &withdrawalService{repo, db}
}

func (s *withdrawalService) Create(createdByID uint, req transactiondto.CreateRequest) (*models.Withdrawal, error) {
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

	totalDep, err := s.repo.SumDeposits(req.ClientID, req.ClientProductID)
	if err != nil {
		return nil, err
	}
	totalWdr, err := s.repo.SumWithdrawals(req.ClientID, req.ClientProductID)
	if err != nil {
		return nil, err
	}

	bal := utils.RoundFloat(totalDep-totalWdr-req.Amount-bonusAmount, 2)
	if req.Bal != 0 {
		bal = req.Bal
	}
	os := utils.RoundFloat(bal-req.TO, 2)
	if req.OS != 0 {
		os = req.OS
	}

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

	if err := s.repo.Create(withdrawal); err != nil {
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

func (s *withdrawalService) Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest) (*models.Withdrawal, error) {
	w, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("withdrawal not found")
	}

	if req.Date != nil {
		w.Date = req.Date.Time
	}
	if req.ClientBankID != nil {
		w.ClientBankID = *req.ClientBankID
	}
	if req.CompanyBankID != nil {
		w.CompanyBankID = *req.CompanyBankID
	}
	if req.Amount != nil {
		w.Amount = *req.Amount
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

	if err := s.repo.Update(w); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *withdrawalService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("withdrawal not found")
	}
	return s.repo.Delete(id)
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
