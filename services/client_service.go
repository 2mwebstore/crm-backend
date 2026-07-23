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
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var dateJoined *time.Time
	if req.DateJoined != nil {
		dateJoined = req.DateJoined.ToTimePtr()
	}

	// Generate code: if branch selected → use branch sequence, else user's assigned branch
	var code string
	if req.Code != "" {
		// Explicit code provided — check uniqueness
		var existing models.Client
		if err := s.db.Where("code = ?", req.Code).First(&existing).Error; err == nil {
			return nil, errors.New("code already exists: " + req.Code)
		}
		code = req.Code
	} else if req.BranchID != nil && *req.BranchID != 0 {
		// Branch selected → generate sequential code for that branch
		code = utils.GenerateICCodeForBranch(s.db, *req.BranchID, utils.EntityClient)
	} else {
		// Fall back to user's assigned branch
		code = utils.GenerateCode(s.db, createdByID, utils.EntityClient)
	}

	client := &models.Client{
		Code:            code,
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
		if err := s.checkDuplicateAccountIDs(req.Products, nil); err != nil {
			return nil, err
		}
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
		if err := s.syncPhones(id, client.Phones, req.Phones); err != nil {
			return nil, err
		}
	}
	if req.Banks != nil {
		if err := s.syncBanks(id, client.Banks, req.Banks); err != nil {
			return nil, err
		}
	}
	if req.Products != nil {
		if err := s.syncProducts(id, client.Products, req.Products); err != nil {
			return nil, err
		}
	}
	return s.repo.FindByIDUnsafe(id)
}

// syncPhones reconciles the client's phone list against the incoming
// request — an input with a matching ID updates that existing row in
// place (preserving its ID/created_at), an input with no ID (or an ID
// that doesn't belong to this client) creates a new row, and any
// existing row whose ID isn't present in the input is deleted. This
// replaces the previous delete-all-then-recreate approach, which
// destroyed and regenerated every row's ID on every single edit —
// harmless for phones specifically (nothing else references a phone by
// ID), but the same pattern on Banks/Products below is what actually
// mattered, since Deposit/Withdrawal transactions reference a specific
// client_bank_id/client_product_id.
func (s *clientService) syncPhones(clientID uint, existing []models.ClientPhone, inputs []clientdto.PhoneInput) error {
	existingByID := make(map[uint]models.ClientPhone, len(existing))
	for _, e := range existing {
		existingByID[e.ID] = e
	}
	keep := map[uint]bool{}
	var toCreate []models.ClientPhone

	for _, in := range inputs {
		label := in.Label
		if label == "" {
			label = "primary"
		}
		if in.ID != nil {
			if e, ok := existingByID[*in.ID]; ok {
				keep[*in.ID] = true
				e.Phone = in.Phone
				e.Label = label
				e.IsPrimary = in.IsPrimary
				e.IsActive = in.IsActive
				e.SortOrder = in.SortOrder
				if err := s.repo.UpdatePhone(&e); err != nil {
					return err
				}
				continue
			}
			// ID given but doesn't belong to this client (or doesn't
			// exist) — fall through and create fresh, rather than
			// silently dropping it or touching an unrelated record.
		}
		toCreate = append(toCreate, models.ClientPhone{
			ClientID: clientID, Phone: in.Phone, Label: label,
			IsPrimary: in.IsPrimary, Status: models.PhoneStatusActive, IsActive: in.IsActive, SortOrder: in.SortOrder,
		})
	}

	for _, e := range existing {
		if !keep[e.ID] {
			if err := s.repo.DeletePhone(e.ID); err != nil {
				return err
			}
		}
	}
	if len(toCreate) > 0 {
		if err := s.repo.CreatePhones(toCreate); err != nil {
			return err
		}
	}
	return nil
}

// syncBanks mirrors syncPhones — same reconcile-by-ID approach. This one
// matters concretely: a Deposit/Withdrawal transaction stores a specific
// client_bank_id, so replacing this bank's row wholesale on every client
// edit would silently point existing transactions at a deleted/wrong
// account.
func (s *clientService) syncBanks(clientID uint, existing []models.ClientBank, inputs []clientdto.BankInput) error {
	existingByID := make(map[uint]models.ClientBank, len(existing))
	for _, e := range existing {
		existingByID[e.ID] = e
	}
	keep := map[uint]bool{}
	for _, in := range inputs {
		if in.ID != nil {
			if _, ok := existingByID[*in.ID]; ok {
				keep[*in.ID] = true
			}
		}
	}

	// Validate every removal BEFORE any mutation happens — so if one bank
	// further down the list turns out to have transactions, the whole
	// update fails cleanly instead of leaving earlier banks already
	// updated/created while this one gets blocked.
	for _, e := range existing {
		if keep[e.ID] {
			continue
		}
		hasTx, err := s.repo.HasTransactionsForBank(e.ID)
		if err != nil {
			return err
		}
		if hasTx {
			return errors.New("cannot delete bank account " + e.AccountNo + " — it has deposit or withdrawal records. Please delete those transactions first")
		}
	}

	var toCreate []models.ClientBank
	for _, in := range inputs {
		if in.ID != nil {
			if e, ok := existingByID[*in.ID]; ok {
				e.BankTypeID = in.BankTypeID
				e.AccountNo = in.AccountNo
				e.AccountName = in.AccountName
				e.IsActive = in.IsActive
				e.SortOrder = in.SortOrder
				if err := s.repo.UpdateBank(&e); err != nil {
					return err
				}
				continue
			}
		}
		toCreate = append(toCreate, models.ClientBank{
			ClientID: clientID, BankTypeID: in.BankTypeID, AccountNo: in.AccountNo,
			AccountName: in.AccountName, IsActive: in.IsActive, SortOrder: in.SortOrder,
		})
	}

	for _, e := range existing {
		if !keep[e.ID] {
			if err := s.repo.DeleteBank(e.ID); err != nil {
				return err
			}
		}
	}
	if len(toCreate) > 0 {
		if err := s.repo.CreateBanks(toCreate); err != nil {
			return err
		}
	}
	return nil
}

// syncProducts mirrors syncBanks — same reconcile-by-ID approach, plus the
// duplicate account_id pre-check. That check now excludes the products
// being kept/updated from its own-DB-row comparison — otherwise a product
// keeping its unchanged account_id would falsely collide with its own
// pre-existing row, since (unlike the old delete-first approach) that row
// is no longer removed before the check runs.
func (s *clientService) syncProducts(clientID uint, existing []models.ClientProduct, inputs []clientdto.ProductInput) error {
	existingByID := make(map[uint]models.ClientProduct, len(existing))
	for _, e := range existing {
		existingByID[e.ID] = e
	}

	keepIDs := make([]uint, 0, len(inputs))
	keep := map[uint]bool{}
	for _, in := range inputs {
		if in.ID != nil {
			if _, ok := existingByID[*in.ID]; ok {
				keepIDs = append(keepIDs, *in.ID)
				keep[*in.ID] = true
			}
		}
	}
	if err := s.checkDuplicateAccountIDs(inputs, keepIDs); err != nil {
		return err
	}

	// Same "validate every removal before any mutation" reasoning as
	// syncBanks above.
	for _, e := range existing {
		if keep[e.ID] {
			continue
		}
		hasTx, err := s.repo.HasTransactionsForProduct(e.ID)
		if err != nil {
			return err
		}
		if hasTx {
			return errors.New("cannot delete product " + e.AccountID + " — it has deposit or withdrawal records. Please delete those transactions first")
		}
	}

	var toCreate []models.ClientProduct
	for _, in := range inputs {
		if in.ID != nil {
			if e, ok := existingByID[*in.ID]; ok {
				e.ProductTypeID = in.ProductTypeID
				e.AccountID = in.AccountID
				e.IsActive = in.IsActive
				e.SortOrder = in.SortOrder
				if err := s.repo.UpdateProduct(&e); err != nil {
					return err
				}
				continue
			}
		}
		toCreate = append(toCreate, models.ClientProduct{
			ClientID: clientID, ProductTypeID: in.ProductTypeID, AccountID: in.AccountID,
			IsActive: in.IsActive, SortOrder: in.SortOrder,
		})
	}

	for _, e := range existing {
		if !keep[e.ID] {
			if err := s.repo.DeleteProduct(e.ID); err != nil {
				return err
			}
		}
	}
	if len(toCreate) > 0 {
		if err := s.repo.CreateProducts(toCreate); err != nil {
			return err
		}
	}
	return nil
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
	hasTx, err := s.repo.HasTransactionsForBank(bankID)
	if err != nil {
		return err
	}
	if hasTx {
		return errors.New("cannot delete this bank account — it has deposit or withdrawal records. Please delete those transactions first")
	}
	return s.repo.DeleteBank(bankID)
}

func (s *clientService) AddProduct(clientID uint, scopeIDs []uint, req clientdto.ProductInput) (*models.ClientProduct, error) {
	if _, err := s.repo.FindByID(clientID, scopeIDs); err != nil {
		return nil, errors.New("client not found")
	}
	// account_id has a DB-level uniqueIndex — checking first gives a clean
	// error message instead of a raw MySQL "Duplicate entry" surfacing
	// straight from CreateProduct below.
	var existing models.ClientProduct
	if err := s.db.Where("account_id = ?", req.AccountID).First(&existing).Error; err == nil {
		return nil, errors.New("account ID already exists: " + req.AccountID)
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
	// Same duplicate pre-check as AddProduct — only matters if the account
	// ID is actually changing, and excludes this row itself from the check.
	if req.AccountID != p.AccountID {
		var existing models.ClientProduct
		if err := s.db.Where("account_id = ? AND id != ?", req.AccountID, productID).First(&existing).Error; err == nil {
			return nil, errors.New("account ID already exists: " + req.AccountID)
		}
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
	hasTx, err := s.repo.HasTransactionsForProduct(productID)
	if err != nil {
		return err
	}
	if hasTx {
		return errors.New("cannot delete this product — it has deposit or withdrawal records. Please delete those transactions first")
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

// checkDuplicateAccountIDs pre-validates a batch of ProductInput against the
// account_id uniqueIndex — both duplicates within the batch itself, and
// against what's already in the DB — so bulk product creation gets the
// same friendly error as the single-product AddProduct path, instead of a
// raw MySQL "Duplicate entry" surfacing straight from the batch insert.
// excludeIDs are existing ClientProduct IDs being kept/updated in this same
// request — their own DB row is excluded from the "already exists" check,
// since keeping an unchanged account_id on an existing row isn't actually a
// collision. Pass nil from Create (nothing pre-exists yet).
func (s *clientService) checkDuplicateAccountIDs(inputs []clientdto.ProductInput, excludeIDs []uint) error {
	seen := map[string]bool{}
	ids := make([]string, 0, len(inputs))
	for _, p := range inputs {
		if seen[p.AccountID] {
			return errors.New("account ID already exists: " + p.AccountID)
		}
		seen[p.AccountID] = true
		ids = append(ids, p.AccountID)
	}
	q := s.db.Where("account_id IN ?", ids)
	if len(excludeIDs) > 0 {
		q = q.Where("id NOT IN ?", excludeIDs)
	}
	var existing models.ClientProduct
	if err := q.First(&existing).Error; err == nil {
		return errors.New("account ID already exists: " + existing.AccountID)
	}
	return nil
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
