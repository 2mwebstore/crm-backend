package services

import (
	"crm-backend/models"
	"crm-backend/repositories"
)

type BalanceTransactionService interface {
	ListByEntity(entityType models.BalanceEntityType, entityID uint, txType models.BalanceTxType, page, pageSize int) ([]models.BalanceTransaction, int64, error)
}

type balanceTransactionService struct {
	repo repositories.BalanceTransactionRepository
}

func NewBalanceTransactionService(repo repositories.BalanceTransactionRepository) BalanceTransactionService {
	return &balanceTransactionService{repo}
}

func (s *balanceTransactionService) ListByEntity(entityType models.BalanceEntityType, entityID uint, txType models.BalanceTxType, page, pageSize int) ([]models.BalanceTransaction, int64, error) {
	return s.repo.ListByEntity(entityType, entityID, txType, page, pageSize)
}
