package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/models"
	"crm-backend/repositories"
)

// BranchInput bundles Create/Update's fields — grown past the point where
// separate positional params (several of the same string type) stayed
// safe to read/call without mixing an argument up.
type BranchInput struct {
	Name        string
	Code        string
	Description string
	IsActive    bool

	// Telegram notification target — see models.Branch for field meaning.
	// All optional; leaving TelegramBotToken/TelegramChatID empty simply
	// means this branch has no Telegram notifications configured.
	TelegramBotToken          string
	TelegramChatID            string
	TelegramDepositTopicID    *int
	TelegramWithdrawalTopicID *int
}

type BranchService interface {
	Create(createdByID uint, input BranchInput) (*models.Branch, error)
	List(showAll bool) ([]models.Branch, error)
	ListForUser(userID uint, showAll bool) ([]models.Branch, error)
	GetByID(id uint) (*models.Branch, error)
	Update(id uint, input BranchInput) (*models.Branch, error)
	Delete(id uint) error
}

type branchService struct{ repo repositories.BranchRepository }

func NewBranchService(repo repositories.BranchRepository) BranchService {
	return &branchService{repo}
}

func (s *branchService) Create(createdByID uint, input BranchInput) (*models.Branch, error) {
	b := &models.Branch{
		Name:                      input.Name,
		Code:                      input.Code,
		Description:               input.Description,
		IsActive:                  true,
		CreatedByID:               createdByID,
		TelegramBotToken:          input.TelegramBotToken,
		TelegramChatID:            input.TelegramChatID,
		TelegramDepositTopicID:    input.TelegramDepositTopicID,
		TelegramWithdrawalTopicID: input.TelegramWithdrawalTopicID,
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	return s.repo.FindByID(b.ID)
}

func (s *branchService) List(showAll bool) ([]models.Branch, error) {
	return s.repo.List(showAll)
}

func (s *branchService) GetByID(id uint) (*models.Branch, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("branch not found")
		}
		return nil, err
	}
	return b, nil
}

func (s *branchService) Update(id uint, input BranchInput) (*models.Branch, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("branch not found")
	}
	if input.Name != "" {
		b.Name = input.Name
	}
	if input.Code != "" {
		b.Code = input.Code
	}
	b.Description = input.Description
	b.IsActive = input.IsActive
	b.TelegramBotToken = input.TelegramBotToken
	b.TelegramChatID = input.TelegramChatID
	b.TelegramDepositTopicID = input.TelegramDepositTopicID
	b.TelegramWithdrawalTopicID = input.TelegramWithdrawalTopicID
	if err := s.repo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("branch not found")
	}
	return s.repo.Delete(id)
}

func (s *branchService) ListForUser(userID uint, showAll bool) ([]models.Branch, error) {
	return s.repo.ListForUser(userID, showAll)
}
