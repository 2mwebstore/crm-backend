package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	turnoverbetdto "crm-backend/dto/turnover_bet"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type TurnoverBetService interface {
	Create(createdByID uint, req turnoverbetdto.CreateRequest) (*models.TurnoverBet, error)
	GetByID(id uint, scopeIDs []uint) (*models.TurnoverBet, error)
	Update(id uint, scopeIDs []uint, req turnoverbetdto.UpdateRequest) (*models.TurnoverBet, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter turnoverbetdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.TurnoverBet, int64, error)
	Approve(id uint, approvedByID uint, req turnoverbetdto.ApproveRequest) (*models.TurnoverBet, error)
}

type turnoverBetService struct {
	repo repositories.TurnoverBetRepository
	db   *gorm.DB
}

func NewTurnoverBetService(repo repositories.TurnoverBetRepository, db *gorm.DB) TurnoverBetService {
	return &turnoverBetService{repo, db}
}

func (s *turnoverBetService) Create(createdByID uint, req turnoverbetdto.CreateRequest) (*models.TurnoverBet, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	t := &models.TurnoverBet{
		Date:          req.Date.Time,
		ProductTypeID: req.ProductTypeID,
		Amount:        req.Amount,
		Currency:      currency,
		Remark:        req.Remark,
		BranchID:      req.BranchID,
		Status:        models.TxStatusPending,
		CreatedByID:   createdByID,
	}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(t.ID)
}

func (s *turnoverBetService) GetByID(id uint, scopeIDs []uint) (*models.TurnoverBet, error) {
	t, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("turnover bet not found")
		}
		return nil, err
	}
	return t, nil
}

func (s *turnoverBetService) Update(id uint, scopeIDs []uint, req turnoverbetdto.UpdateRequest) (*models.TurnoverBet, error) {
	t, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("turnover bet not found")
	}
	if req.Date != nil {
		t.Date = req.Date.Time
	}
	if req.ProductTypeID != nil {
		t.ProductTypeID = *req.ProductTypeID
	}
	if req.Amount != nil {
		t.Amount = *req.Amount
	}
	if req.Currency != nil {
		t.Currency = *req.Currency
	}
	if req.Remark != nil {
		t.Remark = *req.Remark
	}
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *turnoverBetService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("turnover bet not found")
	}
	return s.repo.Delete(id)
}

func (s *turnoverBetService) List(filter turnoverbetdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.TurnoverBet, int64, error) {
	return s.repo.List(filter, p, userID)
}

func (s *turnoverBetService) Approve(id uint, approvedByID uint, req turnoverbetdto.ApproveRequest) (*models.TurnoverBet, error) {
	t, err := s.repo.FindByIDUnsafe(id)
	if err != nil {
		return nil, errors.New("turnover bet not found")
	}
	now := time.Now()
	t.Status = models.TransactionStatus(req.Status)
	t.ApprovedAt = &now
	t.ApprovedByID = &approvedByID
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return s.repo.FindByIDUnsafe(id)
}
