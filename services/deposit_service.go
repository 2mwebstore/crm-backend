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
	Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest) (*models.Deposit, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Deposit, int64, error)
	GetBalance(clientID, clientProductID uint) (*transactiondto.BalanceResponse, error)
	Approve(id uint, approvedByID uint, req transactiondto.ApproveRequest) (*models.Deposit, error)
}

type depositService struct {
	repo       repositories.DepositRepository
	clientRepo repositories.ClientRepository
	db         *gorm.DB
}

func NewDepositService(repo repositories.DepositRepository, clientRepo repositories.ClientRepository, db *gorm.DB) DepositService {
	return &depositService{repo, clientRepo, db}
}

func (s *depositService) Create(createdByID uint, req transactiondto.CreateRequest) (*models.Deposit, error) {
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

	prevBal, err := s.repo.RunningBalance(req.ClientID, req.ClientProductID, 0)
	if err != nil {
		return nil, err
	}
	bal := utils.RoundFloat(prevBal+req.Amount+bonusAmount, 2)
	if req.Bal != 0 {
		bal = req.Bal
	}
	os := utils.RoundFloat(bal-req.TO, 2)
	if req.OS != 0 {
		os = req.OS
	}

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

	if err := s.repo.Create(deposit); err != nil {
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

func (s *depositService) Update(id uint, scopeIDs []uint, req transactiondto.UpdateRequest) (*models.Deposit, error) {
	d, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("deposit not found")
	}

	if req.Date != nil {
		d.Date = req.Date.Time
	}
	if req.ClientBankID != nil {
		d.ClientBankID = *req.ClientBankID
	}
	if req.CompanyBankID != nil {
		d.CompanyBankID = *req.CompanyBankID
	}
	if req.Amount != nil {
		d.Amount = *req.Amount
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
		d.BonusAmount = *req.BonusAmount
	}

	// Bal and OS are stored exactly as given - simple direct input, same as
	// every other manual field. No auto-recalculation.
	if req.Bal != nil {
		d.Bal = *req.Bal
	}
	if req.OS != nil {
		d.OS = *req.OS
	}

	if err := s.repo.Update(d); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *depositService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("deposit not found")
	}
	return s.repo.Delete(id)
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
