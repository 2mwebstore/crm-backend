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

type WithdrawalRepository interface {
	Create(w *models.Withdrawal) error
	FindByID(id uint, scopeIDs []uint) (*models.Withdrawal, error)
	FindByIDUnsafe(id uint) (*models.Withdrawal, error)
	Update(w *models.Withdrawal) error
	Delete(id uint) error
	List(filter transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Withdrawal, int64, error)
	SumDeposits(clientID, clientProductID uint) (float64, error)
	SumWithdrawals(clientID, clientProductID uint) (float64, error)
	// ListSinceForBranch returns every withdrawal for a branch with date >=
	// since (precise timestamp, not just calendar day) — used to show
	// exactly which withdrawals contributed to a shift's income since it
	// was opened.
	ListSinceForBranch(branchID uint, since time.Time) ([]models.Withdrawal, error)
}

type withdrawalRepository struct{ db *gorm.DB }

func NewWithdrawalRepository(db *gorm.DB) WithdrawalRepository {
	return &withdrawalRepository{db}
}

func (r *withdrawalRepository) preload(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Client").
		Preload("ClientProduct.ProductType").
		Preload("ClientBank.BankType").
		Preload("CompanyBank").
		Preload("CompanyBank.BankType").
		Preload("BonusOption").
		Preload("CreatedBy").
		Preload("ApprovedBy")
}

func (r *withdrawalRepository) Create(w *models.Withdrawal) error {
	return r.db.Create(w).Error
}

func (r *withdrawalRepository) FindByID(id uint, scopeIDs []uint) (*models.Withdrawal, error) {
	var w models.Withdrawal
	q := r.preload(r.db)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("withdrawals.id = ? AND withdrawals.branch_id IN ?", id, scopeIDs)
	} else {
		q = q.Where("withdrawals.id = ?", id)
	}
	return &w, q.First(&w).Error
}

func (r *withdrawalRepository) FindByIDUnsafe(id uint) (*models.Withdrawal, error) {
	var w models.Withdrawal
	return &w, r.preload(r.db).First(&w, id).Error
}

// Update saves the withdrawal's own columns only. Two GORM behaviors need
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
func (r *withdrawalRepository) Update(w *models.Withdrawal) error {
	return r.db.Select("*").Omit(clause.Associations).Save(w).Error
}

func (r *withdrawalRepository) Delete(id uint) error {
	return r.db.Delete(&models.Withdrawal{}, id).Error
}

func (r *withdrawalRepository) List(f transactiondto.FilterQuery, p utils.PaginationParams, userID uint) ([]models.Withdrawal, int64, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return nil, 0, nil
	}

	q := r.db.Model(&models.Withdrawal{}).
		Preload("Client").
		Preload("Branch").
		Preload("ClientProduct.ProductType").
		Preload("ClientBank.BankType").
		Preload("CompanyBank").
		Preload("CompanyBank.BankType").
		Preload("BonusOption").
		Preload("CreatedBy").
		Preload("ApprovedBy")

	if !isSA {
		q = q.Where("withdrawals.branch_id IN ?", branchIDs)
	}

	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		q = q.Joins("LEFT JOIN clients ON clients.id = withdrawals.client_id").
			Where("LOWER(withdrawals.transaction_no) LIKE ? OR LOWER(clients.name) LIKE ?", like, like)
	}
	if f.BranchID != nil {
		q = q.Where("withdrawals.branch_id = ?", *f.BranchID)
	}
	if f.CreatedByID != nil {
		q = q.Where("withdrawals.created_by_id = ?", *f.CreatedByID)
	}
	if f.ApprovedByID != nil {
		q = q.Where("withdrawals.approved_by_id = ?", *f.ApprovedByID)
	}
	if f.ClientID != nil {
		q = q.Where("withdrawals.client_id = ?", *f.ClientID)
	}
	if f.CompanyBankTypeID != nil {
		q = q.Where("withdrawals.company_bank_id = ?", *f.CompanyBankTypeID)
	}
	if f.ProductTypeID != nil {
		q = q.Joins("LEFT JOIN client_products ON client_products.id = withdrawals.client_product_id").Where("client_products.product_type_id = ?", *f.ProductTypeID)
	}
	if f.ClientProductID != nil {
		q = q.Where("withdrawals.client_product_id = ?", *f.ClientProductID)
	}
	if f.CompanyBankID != nil {
		q = q.Where("withdrawals.company_bank_id = ?", *f.CompanyBankID)
	}
	if f.Currency != "" {
		q = q.Where("withdrawals.currency = ?", f.Currency)
	}
	if f.Status != "" {
		q = q.Where("withdrawals.status = ?", f.Status)
	}
	if f.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", f.DateFrom); err == nil {
			q = q.Where("withdrawals.date >= ?", t)
		}
	}
	if f.DateTo != "" {
		if t, err := time.Parse("2006-01-02", f.DateTo); err == nil {
			q = q.Where("withdrawals.date <= ?", t.Add(23*time.Hour+59*time.Minute+59*time.Second))
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowed := map[string]string{
		"date": "withdrawals.date", "amount": "withdrawals.amount",
		"created_at": "withdrawals.created_at", "transaction_no": "withdrawals.transaction_no", "bal": "withdrawals.bal",
	}
	q = q.Order(utils.SanitizeSort(f.SortBy, allowed, "withdrawals.date") + " " + utils.SortDir(f.SortDir))
	var list []models.Withdrawal
	return list, total, q.Scopes(utils.Paginate(p)).Find(&list).Error
}

func (r *withdrawalRepository) SumDeposits(clientID, clientProductID uint) (float64, error) {
	var sum float64
	err := r.db.Table("deposits").Where("client_id = ? AND client_product_id = ?", clientID, clientProductID).Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *withdrawalRepository) SumWithdrawals(clientID, clientProductID uint) (float64, error) {
	var sum float64
	err := r.db.Model(&models.Withdrawal{}).Where("client_id = ? AND client_product_id = ?", clientID, clientProductID).Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

// ListSinceForBranch returns every withdrawal for a branch with date >=
// since (a precise timestamp, not just a calendar day) — used to show
// exactly which withdrawals happened during an open shift, since the
// shift's own opening time.
func (r *withdrawalRepository) ListSinceForBranch(branchID uint, since time.Time) ([]models.Withdrawal, error) {
	var list []models.Withdrawal
	err := r.preload(r.db).
		Where("withdrawals.branch_id = ? AND withdrawals.date >= ?", branchID, since).
		Order("withdrawals.date ASC").
		Find(&list).Error
	return list, err
}

func (r *withdrawalRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
