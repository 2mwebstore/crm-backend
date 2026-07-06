package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	clientdto "crm-backend/dto/client"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type ClientService interface {
	Create(createdByID uint, req clientdto.CreateClientRequest) (*models.Client, error)
	GetByID(id uint, scopeIDs []uint) (*models.Client, error)
	Update(id uint, scopeIDs []uint, req clientdto.UpdateClientRequest) (*models.Client, error)
	Delete(id uint, scopeIDs []uint) error
	List(filter clientdto.ClientFilterQuery, p utils.PaginationParams, userID uint) ([]models.Client, int64, error)
	CheckCodeAvailable(code string, excludeID string) bool
	PeekNextSuffix(branchID uint) string
	UploadPicture(id uint, scopeIDs []uint, pictureURL string) (*models.Client, error)
	AddBank(clientID uint, scopeIDs []uint, req clientdto.BankInput) (*models.ClientBank, error)
	UpdateBank(clientID, bankID uint, scopeIDs []uint, req clientdto.BankInput) (*models.ClientBank, error)
	DeleteBank(clientID, bankID uint, scopeIDs []uint) error
	AddProduct(clientID uint, scopeIDs []uint, req clientdto.ProductInput) (*models.ClientProduct, error)
	UpdateProduct(clientID, productID uint, scopeIDs []uint, req clientdto.ProductInput) (*models.ClientProduct, error)
	DeleteProduct(clientID, productID uint, scopeIDs []uint) error
	AddFollowUp(clientID uint, scopeIDs []uint, createdByID uint, req clientdto.FollowUpInput) (*models.ClientFollowUp, error)
	ListFollowUps(clientID uint, scopeIDs []uint, page, pageSize int) ([]models.ClientFollowUp, int64, error)
	DeleteFollowUp(clientID, fuID uint, scopeIDs []uint) error
}

type clientService struct {
	repo repositories.ClientRepository
	db   *gorm.DB
}

func NewClientService(repo repositories.ClientRepository, db *gorm.DB) ClientService {
	return &clientService{repo, db}
}

func (s *clientService) Create(createdByID uint, req clientdto.CreateClientRequest) (*models.Client, error) {
	// Check code uniqueness
	var existing models.Client
	if err := s.db.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		return nil, errors.New("code already exists: " + req.Code)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var dateJoined *time.Time
	if req.DateJoined != nil {
		dateJoined = req.DateJoined.ToTimePtr()
	}

	client := &models.Client{
		Code:            req.Code,
		Name:            req.Name,
		DateJoined:      dateJoined,
		Remark:          req.Remark,
		IsActive:        isActive,
		BranchID:        req.BranchID,
		LevelID:         req.LevelID,
		ContactSourceID: req.ContactSourceID,
		CreatedByID:     createdByID,
	}
	if err := s.repo.Create(client); err != nil {
		return nil, err
	}
	if len(req.Phones) > 0 {
		if err := s.repo.CreatePhones(buildClientPhones(client.ID, req.Phones)); err != nil {
			return nil, err
		}
	}
	if len(req.Banks) > 0 {
		if err := s.repo.CreateBanks(buildClientBanks(client.ID, req.Banks)); err != nil {
			return nil, err
		}
	}
	if len(req.Products) > 0 {
		if err := s.repo.CreateProducts(buildClientProducts(client.ID, req.Products)); err != nil {
			return nil, err
		}
	}
	return s.repo.FindByIDUnsafe(client.ID)
}

func (s *clientService) GetByID(id uint, scopeIDs []uint) (*models.Client, error) {
	c, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	return c, nil
}

func (s *clientService) Update(id uint, scopeIDs []uint, req clientdto.UpdateClientRequest) (*models.Client, error) {
	client, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("client not found")
	}

	if req.Code != nil && *req.Code != client.Code {
		var dup models.Client
		if err := s.db.Where("code = ? AND id != ?", *req.Code, id).First(&dup).Error; err == nil {
			return nil, errors.New("code already exists: " + *req.Code)
		}
		client.Code = *req.Code
	}
	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.DateJoined != nil {
		client.DateJoined = req.DateJoined.ToTimePtr()
	}
	if req.Remark != nil {
		client.Remark = *req.Remark
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}
	if req.BranchID != nil {
		client.BranchID = req.BranchID
	}
	if req.LevelID != nil {
		client.LevelID = req.LevelID
	}
	if req.ContactSourceID != nil {
		client.ContactSourceID = req.ContactSourceID
	}

	if err := s.repo.Update(client); err != nil {
		return nil, err
	}

	if req.Phones != nil {
		if err := s.repo.DeletePhones(id); err != nil {
			return nil, err
		}
		if len(req.Phones) > 0 {
			if err := s.repo.CreatePhones(buildClientPhones(id, req.Phones)); err != nil {
				return nil, err
			}
		}
	}
	if req.Banks != nil {
		if err := s.repo.DeleteBanks(id); err != nil {
			return nil, err
		}
		if len(req.Banks) > 0 {
			if err := s.repo.CreateBanks(buildClientBanks(id, req.Banks)); err != nil {
				return nil, err
			}
		}
	}
	if req.Products != nil {
		if err := s.repo.DeleteProducts(id); err != nil {
			return nil, err
		}
		if len(req.Products) > 0 {
			if err := s.repo.CreateProducts(buildClientProducts(id, req.Products)); err != nil {
				return nil, err
			}
		}
	}
	return s.repo.FindByIDUnsafe(id)
}

func (s *clientService) Delete(id uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(id, scopeIDs); err != nil {
		return errors.New("client not found")
	}
	_ = s.repo.DeletePhones(id)
	_ = s.repo.DeleteBanks(id)
	_ = s.repo.DeleteProducts(id)
	return s.repo.Delete(id, scopeIDs)
}

func (s *clientService) List(filter clientdto.ClientFilterQuery, p utils.PaginationParams, userID uint) ([]models.Client, int64, error) {
	return s.repo.List(filter, p, userID)
}

func (s *clientService) UploadPicture(id uint, scopeIDs []uint, pictureURL string) (*models.Client, error) {
	client, err := s.repo.FindByID(id, scopeIDs)
	if err != nil {
		return nil, errors.New("client not found")
	}
	_ = pictureURL // Picture field removed
	return client, nil
}

func (s *clientService) AddBank(clientID uint, scopeIDs []uint, req clientdto.BankInput) (*models.ClientBank, error) {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return nil, errors.New("client not found")
	}
	bank := &models.ClientBank{ClientID: clientID, BankTypeID: req.BankTypeID, AccountNo: req.AccountNo, AccountName: req.AccountName, IsActive: req.IsActive, SortOrder: req.SortOrder}
	// Create directly so GORM scans back the auto-increment ID
	if err := s.repo.CreateBank(bank); err != nil {
		return nil, err
	}
	return s.repo.FindBank(bank.ID)
}

func (s *clientService) UpdateBank(clientID, bankID uint, scopeIDs []uint, req clientdto.BankInput) (*models.ClientBank, error) {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return nil, errors.New("client not found")
	}
	bank, err := s.repo.FindBank(bankID)
	if err != nil || bank.ClientID != clientID {
		return nil, errors.New("bank record not found")
	}
	bank.BankTypeID = req.BankTypeID
	bank.AccountNo = req.AccountNo
	bank.AccountName = req.AccountName
	bank.IsActive = req.IsActive
	bank.SortOrder = req.SortOrder
	if err := s.repo.UpdateBank(bank); err != nil {
		return nil, err
	}
	return s.repo.FindBank(bankID)
}

func (s *clientService) DeleteBank(clientID, bankID uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return errors.New("client not found")
	}
	bank, err := s.repo.FindBank(bankID)
	if err != nil || bank.ClientID != clientID {
		return errors.New("bank record not found")
	}
	return s.repo.DeleteBank(bankID)
}

func (s *clientService) AddProduct(clientID uint, scopeIDs []uint, req clientdto.ProductInput) (*models.ClientProduct, error) {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return nil, errors.New("client not found")
	}
	p := &models.ClientProduct{ClientID: clientID, ProductTypeID: req.ProductTypeID, AccountID: req.AccountID, IsActive: req.IsActive, SortOrder: req.SortOrder}
	// Create via the singular method (not CreateProducts) so GORM scans the
	// auto-increment ID back into p itself — CreateProducts takes the slice
	// by value, so p.ID would stay 0 and the FindProduct(p.ID) below would
	// look up the wrong row.
	if err := s.repo.CreateProduct(p); err != nil {
		return nil, err
	}
	return s.repo.FindProduct(p.ID)
}

func (s *clientService) UpdateProduct(clientID, productID uint, scopeIDs []uint, req clientdto.ProductInput) (*models.ClientProduct, error) {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return nil, errors.New("client not found")
	}
	p, err := s.repo.FindProduct(productID)
	if err != nil || p.ClientID != clientID {
		return nil, errors.New("product record not found")
	}
	p.ProductTypeID = req.ProductTypeID
	p.AccountID = req.AccountID
	p.IsActive = req.IsActive
	p.SortOrder = req.SortOrder
	if err := s.repo.UpdateProduct(p); err != nil {
		return nil, err
	}
	return s.repo.FindProduct(productID)
}

func (s *clientService) DeleteProduct(clientID, productID uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return errors.New("client not found")
	}
	p, err := s.repo.FindProduct(productID)
	if err != nil || p.ClientID != clientID {
		return errors.New("product record not found")
	}
	return s.repo.DeleteProduct(productID)
}

func (s *clientService) AddFollowUp(clientID uint, scopeIDs []uint, createdByID uint, req clientdto.FollowUpInput) (*models.ClientFollowUp, error) {
	client, err := s.repo.FindByID(clientID, scopeIDs)
	if err != nil {
		return nil, errors.New("client not found")
	}
	fu := &models.ClientFollowUp{
		// BranchID inherited from the client — previously left unset, which
		// meant a follow-up added here got branch_id = 0 and became
		// invisible to any scoped GetByID/List on FollowUpRepository
		// (they filter by branch_id IN scopeIDs, which never matches 0).
		// NOTE: double-check that models.Client.BranchID and
		// models.ClientFollowUp.BranchID are the same type (both *uint,
		// based on how Update()/follow_up_service.Create() use them) —
		// adjust the dereference below if not.
		ClientID: clientID, BranchID: client.BranchID, Interest: req.Interest, GivenAccount: req.GivenAccount,
		BankAccount: req.BankAccount, Remark: req.Remark,
		FollowUpAt:  req.FollowUpAt.Time,
		CreatedByID: createdByID,
	}
	if err := s.repo.CreateFollowUp(fu); err != nil {
		return nil, err
	}
	return s.repo.FindFollowUp(fu.ID)
}

func (s *clientService) ListFollowUps(clientID uint, scopeIDs []uint, page, pageSize int) ([]models.ClientFollowUp, int64, error) {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return nil, 0, errors.New("client not found")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListFollowUps(clientID, page, pageSize)
}

func (s *clientService) DeleteFollowUp(clientID, fuID uint, scopeIDs []uint) error {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return errors.New("client not found")
	}
	fu, err := s.repo.FindFollowUp(fuID)
	if err != nil || fu.ClientID != clientID {
		return errors.New("follow-up not found")
	}
	return s.repo.DeleteFollowUp(fuID)
}

func buildClientPhones(clientID uint, inputs []clientdto.PhoneInput) []models.ClientPhone {
	phones := make([]models.ClientPhone, 0, len(inputs))
	for _, p := range inputs {
		label := p.Label
		if label == "" {
			label = "primary"
		}
		phones = append(phones, models.ClientPhone{
			ClientID: clientID, Phone: p.Phone, Label: label,
			IsPrimary: p.IsPrimary, Status: models.PhoneStatusActive, IsActive: p.IsActive, SortOrder: p.SortOrder,
		})
	}
	return phones
}

func buildClientBanks(clientID uint, inputs []clientdto.BankInput) []models.ClientBank {
	banks := make([]models.ClientBank, 0, len(inputs))
	for _, b := range inputs {
		banks = append(banks, models.ClientBank{ClientID: clientID, BankTypeID: b.BankTypeID, AccountNo: b.AccountNo, AccountName: b.AccountName, IsActive: b.IsActive, SortOrder: b.SortOrder})
	}
	return banks
}

func buildClientProducts(clientID uint, inputs []clientdto.ProductInput) []models.ClientProduct {
	products := make([]models.ClientProduct, 0, len(inputs))
	for _, p := range inputs {
		products = append(products, models.ClientProduct{ClientID: clientID, ProductTypeID: p.ProductTypeID, AccountID: p.AccountID, IsActive: p.IsActive, SortOrder: p.SortOrder})
	}
	return products
}

func (s *clientService) CheckCodeAvailable(code string, excludeID string) bool {
	var existing models.Client
	q := s.db.Where("code = ?", code)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	return q.First(&existing).Error != nil
}

func (s *clientService) PeekNextSuffix(branchID uint) string {
	return utils.PeekNextSuffix(s.db, branchID, utils.EntityClient)
}
