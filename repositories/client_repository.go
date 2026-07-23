package repositories

import (
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	clientdto "crm-backend/dto/client"
	"crm-backend/models"
	"crm-backend/utils"
)

type ClientRepository interface {
	Create(client *models.Client) error
	FindByID(id uint, scopeIDs []uint) (*models.Client, error)
	FindByIDUnsafe(id uint) (*models.Client, error)
	Update(client *models.Client) error
	Delete(id uint, scopeIDs []uint) error
	List(filter clientdto.ClientFilterQuery, p utils.PaginationParams, userID uint) ([]models.Client, int64, error)

	// Phone section
	DeletePhones(clientID uint) error
	DeletePhone(id uint) error
	CreatePhones(phones []models.ClientPhone) error
	FindPhone(id uint) (*models.ClientPhone, error)
	UpdatePhone(phone *models.ClientPhone) error

	// Bank section
	DeleteBanks(clientID uint) error
	CreateBank(bank *models.ClientBank) error
	CreateBanks(banks []models.ClientBank) error
	ListBanks(clientID uint) ([]models.ClientBank, error)
	FindBank(id uint) (*models.ClientBank, error)
	UpdateBank(bank *models.ClientBank) error
	DeleteBank(id uint) error
	// HasTransactionsForBank reports whether any Deposit or Withdrawal
	// references this bank account — checked before allowing a delete,
	// since both tables store a non-nullable client_bank_id foreign key
	// that would otherwise point at a deleted row.
	HasTransactionsForBank(bankID uint) (bool, error)

	// Product (Player) section
	DeleteProducts(clientID uint) error
	CreateProduct(product *models.ClientProduct) error
	CreateProducts(products []models.ClientProduct) error
	ListProducts(clientID uint) ([]models.ClientProduct, error)
	FindProduct(id uint) (*models.ClientProduct, error)
	UpdateProduct(product *models.ClientProduct) error
	DeleteProduct(id uint) error
	// HasTransactionsForProduct mirrors HasTransactionsForBank above, for
	// the client_product_id foreign key instead.
	HasTransactionsForProduct(productID uint) (bool, error)

	// Follow Up section
	CreateFollowUp(fu *models.ClientFollowUp) error
	ListFollowUps(clientID uint, page, pageSize int) ([]models.ClientFollowUp, int64, error)
	FindFollowUp(id uint) (*models.ClientFollowUp, error)
	DeleteFollowUp(id uint) error
}

type clientRepository struct{ db *gorm.DB }

func NewClientRepository(db *gorm.DB) ClientRepository { return &clientRepository{db} }

func (r *clientRepository) preload(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Branch").Preload("Level").Preload("ContactSource").
		Preload("CreatedBy").
		Preload("Phones").
		Preload("Banks.BankType").
		Preload("Products.ProductType")
}

func (r *clientRepository) Create(c *models.Client) error { return r.db.Create(c).Error }
func (r *clientRepository) Update(c *models.Client) error { return r.db.Save(c).Error }

// Delete removes a client by ID. If scopeIDs is non-nil, the record must
// belong to one of those branch IDs, otherwise nothing is deleted and
// ErrRecordNotFound is returned.
func (r *clientRepository) Delete(id uint, scopeIDs []uint) error {
	if scopeIDs != nil && len(scopeIDs) == 0 {
		return gorm.ErrRecordNotFound
	}
	q := r.db.Where("id = ?", id)
	if scopeIDs != nil {
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	res := q.Delete(&models.Client{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindByID loads a client by ID. If scopeIDs is non-nil, the record must
// belong to one of those branch IDs or ErrRecordNotFound is returned (so
// callers can't distinguish "doesn't exist" from "not in your scope").
func (r *clientRepository) FindByID(id uint, scopeIDs []uint) (*models.Client, error) {
	var c models.Client
	q := r.preload(r.db).Where("id = ?", id)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	if err := q.First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *clientRepository) FindByIDUnsafe(id uint) (*models.Client, error) {
	var c models.Client
	return &c, r.preload(r.db).First(&c, id).Error
}

func (r *clientRepository) List(f clientdto.ClientFilterQuery, p utils.PaginationParams, userID uint) ([]models.Client, int64, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return []models.Client{}, 0, nil
	}
	q := r.preload(r.db.Debug().Model(&models.Client{}))

	if !isSA {
		q = q.Where("clients.branch_id IN ?", branchIDs)
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		like := "%" + search + "%"
		q = q.Where("clients.name LIKE ? OR clients.code LIKE ?", like, like)
	}
	log.Printf("Search received: %q", f.Search)
	if f.IsActive != nil {
		q = q.Where("clients.is_active = ?", *f.IsActive)
	}
	if f.BranchID != nil {
		q = q.Where("clients.branch_id = ?", *f.BranchID)
	}
	if f.CreatedByID != nil {
		q = q.Where("clients.created_by_id = ?", *f.CreatedByID)
	}
	if f.LevelID != nil {
		q = q.Where("clients.level_id = ?", *f.LevelID)
	}
	if f.ContactSourceID != nil {
		q = q.Where("clients.contact_source_id = ?", *f.ContactSourceID)
	}
	if f.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			q = q.Where("clients.date_joined >= ?", t)
		}
	}
	if f.DateTo != "" {
		if t, err := time.Parse("2006-01-02", f.DateTo); err == nil {
			q = q.Where("clients.date_joined <= ?", t.Add(23*time.Hour+59*time.Minute+59*time.Second))
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowed := map[string]string{
		"name": "clients.name", "created_at": "clients.created_at",
		"updated_at": "clients.updated_at", "date_joined": "clients.date_joined",
	}
	q = q.Order(utils.SanitizeSort(f.SortBy, allowed, "clients.created_at") + " " + utils.SortDir(f.SortDir))

	var list []models.Client
	return list, total, q.Scopes(utils.Paginate(p)).Find(&list).Error
}

// ── Phone ─────────────────────────────────────────────────────────────────────

func (r *clientRepository) DeletePhones(clientID uint) error {
	return r.db.Exec("DELETE FROM client_phones WHERE client_id = ?", clientID).Error
}
func (r *clientRepository) DeletePhone(id uint) error {
	return r.db.Delete(&models.ClientPhone{}, id).Error
}
func (r *clientRepository) CreatePhones(phones []models.ClientPhone) error {
	if len(phones) == 0 {
		return nil
	}
	return r.db.Create(&phones).Error
}
func (r *clientRepository) FindPhone(id uint) (*models.ClientPhone, error) {
	var p models.ClientPhone
	return &p, r.db.First(&p, id).Error
}
func (r *clientRepository) UpdatePhone(p *models.ClientPhone) error { return r.db.Save(p).Error }

// ── Bank ──────────────────────────────────────────────────────────────────────

func (r *clientRepository) DeleteBanks(clientID uint) error {
	return r.db.Exec("DELETE FROM client_banks WHERE client_id = ?", clientID).Error
}
func (r *clientRepository) CreateBank(bank *models.ClientBank) error {
	return r.db.Create(bank).Error
}
func (r *clientRepository) CreateBanks(banks []models.ClientBank) error {
	if len(banks) == 0 {
		return nil
	}
	return r.db.Create(&banks).Error
}
func (r *clientRepository) ListBanks(clientID uint) ([]models.ClientBank, error) {
	var banks []models.ClientBank
	err := r.db.Preload("BankType").Where("client_id = ?", clientID).Order("sort_order ASC").Find(&banks).Error
	return banks, err
}
func (r *clientRepository) FindBank(id uint) (*models.ClientBank, error) {
	var b models.ClientBank
	return &b, r.db.Preload("BankType").First(&b, id).Error
}
func (r *clientRepository) UpdateBank(b *models.ClientBank) error { return r.db.Save(b).Error }
func (r *clientRepository) DeleteBank(id uint) error {
	return r.db.Delete(&models.ClientBank{}, id).Error
}

// HasTransactionsForBank checks both the deposits and withdrawals tables
// directly (raw table names, not the Deposit/Withdrawal Go models) to
// avoid this repository needing to import the transactions domain just
// for a single existence check.
func (r *clientRepository) HasTransactionsForBank(bankID uint) (bool, error) {
	var count int64
	if err := r.db.Table("deposits").Where("client_bank_id = ?", bankID).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	if err := r.db.Table("withdrawals").Where("client_bank_id = ?", bankID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ── Product ───────────────────────────────────────────────────────────────────

func (r *clientRepository) DeleteProducts(clientID uint) error {
	return r.db.Exec("DELETE FROM client_products WHERE client_id = ?", clientID).Error
}
func (r *clientRepository) CreateProduct(product *models.ClientProduct) error {
	return r.db.Create(product).Error
}
func (r *clientRepository) CreateProducts(products []models.ClientProduct) error {
	if len(products) == 0 {
		return nil
	}
	return r.db.Create(&products).Error
}
func (r *clientRepository) ListProducts(clientID uint) ([]models.ClientProduct, error) {
	var products []models.ClientProduct
	err := r.db.Preload("ProductType").Where("client_id = ?", clientID).Order("sort_order ASC").Find(&products).Error
	return products, err
}
func (r *clientRepository) FindProduct(id uint) (*models.ClientProduct, error) {
	var p models.ClientProduct
	return &p, r.db.Preload("ProductType").First(&p, id).Error
}
func (r *clientRepository) UpdateProduct(p *models.ClientProduct) error { return r.db.Save(p).Error }
func (r *clientRepository) DeleteProduct(id uint) error {
	return r.db.Delete(&models.ClientProduct{}, id).Error
}

// HasTransactionsForProduct mirrors HasTransactionsForBank, for
// client_product_id instead.
func (r *clientRepository) HasTransactionsForProduct(productID uint) (bool, error) {
	var count int64
	if err := r.db.Table("deposits").Where("client_product_id = ?", productID).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	if err := r.db.Table("withdrawals").Where("client_product_id = ?", productID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ── Follow Up ─────────────────────────────────────────────────────────────────

func (r *clientRepository) CreateFollowUp(fu *models.ClientFollowUp) error {
	return r.db.Create(fu).Error
}
func (r *clientRepository) ListFollowUps(clientID uint, page, pageSize int) ([]models.ClientFollowUp, int64, error) {
	var list []models.ClientFollowUp
	var total int64
	q := r.db.Model(&models.ClientFollowUp{}).Preload("CreatedBy").Where("client_id = ?", clientID)
	q.Count(&total)
	err := q.Order("follow_up_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
func (r *clientRepository) FindFollowUp(id uint) (*models.ClientFollowUp, error) {
	var fu models.ClientFollowUp
	return &fu, r.db.Preload("CreatedBy").First(&fu, id).Error
}
func (r *clientRepository) DeleteFollowUp(id uint) error {
	return r.db.Delete(&models.ClientFollowUp{}, id).Error
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *clientRepository) resolveUserBranches(userID uint) ([]uint, bool) {
	var row struct {
		IsSuperAdmin bool
	}
	if err := r.db.Raw("SELECT is_super_admin FROM users WHERE id = ?", userID).Scan(&row).Error; err != nil {
		return []uint{}, false
	}
	if row.IsSuperAdmin {
		return nil, true
	}
	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", userID).Scan(&branchIDs)
	return branchIDs, false
}
