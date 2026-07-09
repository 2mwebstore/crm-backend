package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	interestingdto "crm-backend/dto/interesting_client"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type InterestingClientService interface {
	Create(createdByID uint, req interestingdto.CreateRequest) (*models.InterestingClient, error)
	GetByID(id uint, scopeIDs []uint) (*models.InterestingClient, error)
	Update(id uint, scopeIDs []uint, req interestingdto.UpdateRequest) (*models.InterestingClient, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter interestingdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.InterestingClient, int64, error)
	CheckCodeAvailable(code string, excludeID string) bool
	PeekNextSuffix(branchID uint) string
	Convert(id uint, scopeIDs []uint, req interestingdto.ConvertRequest, createdByID uint, clientRepo repositories.ClientRepository) (*models.Client, error)
	UpdatePhone(icID, phoneID uint, scopeIDs []uint, req interestingdto.PhoneInput) (*models.InterestingClientPhone, error)
	DeletePhone(icID, phoneID uint, scopeIDs []uint) error
}

type interestingClientService struct {
	repo repositories.InterestingClientRepository
	db   *gorm.DB
}

func NewInterestingClientService(repo repositories.InterestingClientRepository, db *gorm.DB) InterestingClientService {
	return &interestingClientService{repo, db}
}

func (s *interestingClientService) Create(createdByID uint, req interestingdto.CreateRequest) (*models.InterestingClient, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now()
	dateJoined := &now
	if req.DateJoined != nil {
		dateJoined = req.DateJoined.ToTimePtr()
	}

	// Generate code: if branch selected → use branch sequence, else user's assigned branch
	var code string
	if req.Code != "" {
		// Explicit code provided — check uniqueness
		var dup models.InterestingClient
		if s.db.Where("code = ?", req.Code).First(&dup).Error == nil {
			return nil, errors.New("code already exists: " + req.Code)
		}
		code = req.Code
	} else if req.BranchID != nil && *req.BranchID != 0 {
		// Branch selected → generate sequential code for that branch
		code = utils.GenerateICCodeForBranch(s.db, *req.BranchID, utils.EntityIC)
	} else {
		// Fall back to user's assigned branch
		code = utils.GenerateCode(s.db, createdByID, utils.EntityIC)
	}

	ic := &models.InterestingClient{
		Code:            code,
		FullName:        req.FullName,
		DateJoined:      dateJoined,
		Remark:          req.Remark,
		IsActive:        isActive,
		BranchID:        req.BranchID,
		ContactSourceID: req.ContactSourceID,
		CreatedByID:     createdByID,
	}

	if err := s.repo.Create(ic); err != nil {
		return nil, err
	}
	if len(req.Phones) > 0 {
		if err := s.repo.CreatePhones(buildICPhones(ic.ID, req.Phones)); err != nil {
			return nil, err
		}
	}
	return s.repo.FindByIDUnsafe(ic.ID)
}

func (s *interestingClientService) GetByID(id uint, scopeIDs []uint) (*models.InterestingClient, error) {
	ic, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("interesting client not found")
		}
		return nil, err
	}
	return ic, nil
}

func (s *interestingClientService) Update(id uint, scopeIDs []uint, req interestingdto.UpdateRequest) (*models.InterestingClient, error) {
	ic, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("interesting client not found")
	}

	if req.FullName != nil {
		ic.FullName = *req.FullName
	}
	if req.DateJoined != nil {
		ic.DateJoined = req.DateJoined.ToTimePtr()
	}
	if req.Remark != nil {
		ic.Remark = *req.Remark
	}
	if req.IsActive != nil {
		ic.IsActive = *req.IsActive
	}
	if req.BranchID != nil {
		ic.BranchID = req.BranchID
	}
	if req.ContactSourceID != nil {
		ic.ContactSourceID = req.ContactSourceID
	}
	if req.Code != nil && *req.Code != "" && *req.Code != ic.Code {
		var dup models.InterestingClient
		if s.db.Where("code = ? AND id != ?", *req.Code, id).First(&dup).Error == nil {
			return nil, errors.New("code already exists: " + *req.Code)
		}
		ic.Code = *req.Code
	}

	if err := s.repo.Update(ic); err != nil {
		return nil, err
	}

	if req.Phones != nil {
		_ = s.repo.DeletePhones(id)
		if len(req.Phones) > 0 {
			_ = s.repo.CreatePhones(buildICPhones(id, req.Phones))
		}
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *interestingClientService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("interesting client not found")
	}
	_ = s.repo.DeletePhones(id)
	return s.repo.Delete(id)
}

func (s *interestingClientService) List(filter interestingdto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.InterestingClient, int64, error) {
	return s.repo.List(filter, p, userID)
}
func (s *interestingClientService) Convert(id uint, scopeIDs []uint, req interestingdto.ConvertRequest, createdByID uint, clientRepo repositories.ClientRepository) (*models.Client, error) {
	ic, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("interesting client not found")
	}
	if ic.IsConverted {
		return nil, errors.New("already converted to client")
	}

	var client *models.Client
	if req.ExistingClientID != nil {
		// Link to existing client
		client, err = clientRepo.FindByID(*req.ExistingClientID, scopeIDs)
		if err != nil {
			return nil, errors.New("existing client not found")
		}
	} else {
		// Generate code the same way clientService.Create does — never trust
		// a frontend-previewed code as final, since preview is a non-mutating
		// peek and never increments code_sequences.
		var code string
		if req.BranchID != nil && *req.BranchID != 0 {
			// Branch selected → generate + increment sequential code for that branch
			code = utils.GenerateICCodeForBranch(s.db, *req.BranchID, utils.EntityClient)
		} else {
			// Fall back to user's assigned branch
			code = utils.GenerateCode(s.db, createdByID, utils.EntityClient)
		}

		dateJoined := utils.NowInPhnomPenh()
		if ic.DateJoined != nil {
			dateJoined = ic.DateJoined
		}
		newClient := &models.Client{
			Code:            code,
			Name:            ic.FullName,
			DateJoined:      dateJoined,
			Remark:          ic.Remark,
			IsActive:        ic.IsActive,
			BranchID:        req.BranchID,
			ContactSourceID: ic.ContactSourceID,
			CreatedByID:     createdByID,
		}
		if err := clientRepo.Create(newClient); err != nil {
			return nil, err
		}
		// Copy phones from IC to new client
		if len(ic.Phones) > 0 {
			clientPhones := make([]models.ClientPhone, 0, len(ic.Phones))
			for _, p := range ic.Phones {
				clientPhones = append(clientPhones, models.ClientPhone{
					ClientID:  newClient.ID,
					Phone:     p.Phone,
					Label:     p.Label,
					IsPrimary: p.IsPrimary,
					Status:    models.PhoneStatusActive,
					IsActive:  p.IsActive,
				})
			}
			_ = clientRepo.CreatePhones(clientPhones)
		}
		client = newClient
	}

	// Mark IC as converted
	now := time.Now()
	ic.IsConverted = true
	ic.ConvertedAt = &now
	ic.ConvertedClientID = &client.ID
	_ = s.repo.Update(ic)
	return clientRepo.FindByIDUnsafe(client.ID)
}

func (s *interestingClientService) UpdatePhone(icID, phoneID uint, scopeIDs []uint, req interestingdto.PhoneInput) (*models.InterestingClientPhone, error) {
	if _, err := s.repo.FindByID(icID, scopeIDs); err != nil {
		return nil, errors.New("interesting client not found")
	}
	p, err := s.repo.FindPhone(phoneID)
	if err != nil || p.InterestingClientID != icID {
		return nil, errors.New("phone not found")
	}
	p.Phone = req.Phone
	p.Label = req.Label
	p.IsPrimary = req.IsPrimary
	if req.Status != "" {
		p.Status = models.PhoneStatus(req.Status)
	}
	p.IsActive = req.IsActive
	if err := s.repo.UpdatePhone(p); err != nil {
		return nil, err
	}
	return s.repo.FindPhone(phoneID)
}

func (s *interestingClientService) DeletePhone(icID, phoneID uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(icID, scopeIDs); err != nil {
		return errors.New("interesting client not found")
	}
	p, err := s.repo.FindPhone(phoneID)
	if err != nil || p.InterestingClientID != icID {
		return errors.New("phone not found")
	}
	return s.repo.UpdatePhone(&models.InterestingClientPhone{ID: phoneID, IsActive: false})
}

func buildICPhones(icID uint, inputs []interestingdto.PhoneInput) []models.InterestingClientPhone {
	phones := make([]models.InterestingClientPhone, 0, len(inputs))
	for _, p := range inputs {
		label := p.Label
		if label == "" {
			label = "primary"
		}
		status := models.PhoneStatusActive
		phones = append(phones, models.InterestingClientPhone{
			InterestingClientID: icID, Phone: p.Phone, Label: label,
			IsPrimary: p.IsPrimary, Status: status, IsActive: p.IsActive,
		})
	}
	return phones
}

func (s *interestingClientService) CheckCodeAvailable(code string, excludeID string) bool {
	var existing models.InterestingClient
	q := s.db.Where("code = ?", code)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	return q.First(&existing).Error != nil
}

func (s *interestingClientService) PeekNextSuffix(branchID uint) string {
	return utils.PeekNextSuffix(s.db, branchID, utils.EntityIC)
}
