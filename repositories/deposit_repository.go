package repositories

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	transactiondto "crm-backend/dto/transaction"
	"crm-backend/models"
	"crm-backend/utils"
)

type DepositRepository interface {
	Create(d *models.Deposit) error
	FindByID(id uint, scopeIDs []uint) (*models.Deposit, error)
	FindByIDUnsafe(id uint) (*models.Deposit, error)
	Update(d *models.Deposit) error
	Delete(id uint) error
	List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Deposit, int64, error)
	SumDeposits(clientID, clientProductID uint) (float64, error)
	SumWithdrawals(clientID, clientProductID uint) (float64, error)
	RunningBalance(clientID, clientProductID uint, beforeID uint) (float64, error)
}

type depositRepository struct{ db *gorm.DB }

func NewDepositRepository(db *gorm.DB) DepositRepository {
	return &depositRepository{db}
}

func (r *depositRepository) preload(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Client").
		Preload("ClientProduct.ProductType").
		Preload("ClientBank.BankType").
		Preload("CompanyBank").
		Preload("BonusOption").
		Preload("CreatedBy").
		Preload("ApprovedBy")
}

func (r *depositRepository) Create(d *models.Deposit) error {
	return r.db.Create(d).Error
}

func (r *depositRepository) FindByID(id uint, scopeIDs []uint) (*models.Deposit, error) {
	var d models.Deposit
	q := r.preload(r.db)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("deposits.id = ? AND deposits.branch_id IN ?", id, scopeIDs)
	} else {
		q = q.Where("deposits.id = ?", id)
	}
	return &d, q.First(&d).Error
}

func (r *depositRepository) FindByIDUnsafe(id uint) (*models.Deposit, error) {
	var d models.Deposit
	return &d, r.preload(r.db).First(&d, id).Error
}

// Update saves the deposit's own columns only. Two GORM behaviors need
// explicit overriding here:
//  1. Select("*") - several numeric fields (Bal, OS, TO, Play, BonusAmount)
//     have a `default:0` GORM tag. Without Select("*"), GORM silently
//     drops any such field from the UPDATE's SET clause whenever its value
//     is exactly zero (deferring to the column's DB default instead) -
//     meaning setting Bal or OS to 0 would leave the old stored value
//     untouched. Select("*") forces every field to be included regardless.
//  2. Omit(clause.Associations) - FindByID preloads related structs (e.g.
//     BonusOption), and GORM's default Save() behavior re-derives
//     belongs-to foreign keys from those loaded association structs -
//     which would silently override a manually-cleared BonusOptionID back
//     to its old value.
func (r *depositRepository) Update(d *models.Deposit) error {
	return r.db.Select("*").Omit(clause.Associations).Save(d).Error
}
func (r *depositRepository) Delete(id uint) error { return r.db.Delete(&models.Deposit{}, id).Error }

func (r *depositRepository) List(f transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Deposit, int64, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return nil, 0, nil
	}

	q := r.db.Model(&models.Deposit{}).
		Preload("Client").
		Preload("Branch").
		Preload("ClientProduct.ProductType").
		Preload("ClientBank.BankType").
		Preload("CompanyBank").
		Preload("BonusOption").
		Preload("CreatedBy").
		Preload("ApprovedBy")

	if !isSA {
		q = q.Where("deposits.branch_id IN ?", branchIDs)
	}

	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		q = q.Joins("LEFT JOIN clients ON clients.id = deposits.client_id").
			Where("LOWER(deposits.transaction_no) LIKE ? OR LOWER(clients.name) LIKE ?", like, like)
	}
	if f.BranchID != nil {
		q = q.Where("deposits.branch_id = ?", *f.BranchID)
	}
	if f.CreatedByID != nil {
		q = q.Where("deposits.created_by_id = ?", *f.CreatedByID)
	}
	if f.ApprovedByID != nil {
		q = q.Where("deposits.approved_by_id = ?", *f.ApprovedByID)
	}
	if f.ClientID != nil {
		q = q.Where("deposits.client_id = ?", *f.ClientID)
	}
	if f.CompanyBankTypeID != nil {
		q = q.Where("deposits.company_bank_id = ?", *f.CompanyBankTypeID)
	}
	if f.ProductTypeID != nil {
		q = q.Joins("LEFT JOIN client_products ON client_products.id = deposits.client_product_id").Where("client_products.product_type_id = ?", *f.ProductTypeID)
	}
	if f.ClientProductID != nil {
		q = q.Where("deposits.client_product_id = ?", *f.ClientProductID)
	}
	if f.CompanyBankID != nil {
		q = q.Where("deposits.company_bank_id = ?", *f.CompanyBankID)
	}
	if f.Currency != "" {
		q = q.Where("deposits.currency = ?", f.Currency)
	}
	if f.Status != "" {
		q = q.Where("deposits.status = ?", f.Status)
	}
	if f.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			q = q.Where("deposits.date >= ?", t)
		}
	}
	if f.DateTo != "" {
		if t, err := time.Parse("2006-01-02", f.DateTo); err == nil {
			q = q.Where("deposits.date <= ?", t.Add(23*time.Hour+59*time.Minute+59*time.Second))
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowed := map[string]string{
		"date": "deposits.date", "amount": "deposits.amount",
		"created_at": "deposits.created_at", "transaction_no": "deposits.transaction_no", "bal": "deposits.bal",
	}
	q = q.Order(utils.SanitizeSort(f.SortBy, allowed, "deposits.date") + " " + utils.SortDir(f.SortDir))
	var list []models.Deposit
	return list, total, q.Scopes(utils.Paginate(p)).Find(&list).Error
}

func (r *depositRepository) SumDeposits(clientID, clientProductID uint) (float64, error) {
	var sum float64
	err := r.db.Model(&models.Deposit{}).Where("client_id = ? AND client_product_id = ?", clientID, clientProductID).Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *depositRepository) SumWithdrawals(clientID, clientProductID uint) (float64, error) {
	var sum float64
	err := r.db.Table("withdrawals").Where("client_id = ? AND client_product_id = ?", clientID, clientProductID).Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *depositRepository) RunningBalance(clientID, clientProductID uint, beforeID uint) (float64, error) {
	var sumDep, sumWdr float64
	q := r.db.Model(&models.Deposit{}).Where("client_id = ? AND client_product_id = ?", clientID, clientProductID)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	if err := q.Select("COALESCE(SUM(amount), 0)").Scan(&sumDep).Error; err != nil {
		return 0, err
	}
	// Withdrawals live in their own table with an independent auto-increment
	// sequence, so "beforeID" (a deposit ID) has no meaningful relationship
	// to a withdrawal's id - applying it here as a filter compared two
	// unrelated ID spaces and produced an effectively arbitrary sum
	// whenever this was called with a real ID (i.e. on Update; Create
	// always passed 0, which skipped the filter, so this bug was invisible
	// there). Withdrawals are summed in full, consistent with how Create
	// already computes prevBal.
	qw := r.db.Table("withdrawals").Where("client_id = ? AND client_product_id = ?", clientID, clientProductID)
	if err := qw.Select("COALESCE(SUM(amount), 0)").Scan(&sumWdr).Error; err != nil {
		return 0, err
	}
	return sumDep - sumWdr, nil
}

func (r *depositRepository) resolveUserBranches(userID uint) ([]uint, bool) {
	var row struct{ IsSuperAdmin bool }
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
