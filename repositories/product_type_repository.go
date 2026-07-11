package repositories

import (
	"errors"
	"fmt"

	"crm-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductTypeRepository interface {
	Create(x *models.ProductType, createdByID uint) error
	FindByID(id uint, scopeIDs []uint) (*models.ProductType, error)
	ListForUser(userID uint, showAll bool, branchID *uint) ([]models.ProductType, error)
	List(showAll bool) ([]models.ProductType, error)
	Update(x *models.ProductType) error
	Delete(id uint, scopeIDs []uint) error
	ExistsByName(name string, excludeID uint) bool
	// TopUpCredit and WithdrawCredit apply an atomic, row-locked balance
	// change and write a matching BalanceTransaction ledger row in the same
	// DB transaction (see CompanyBankRepository.applyCashDelta for the same
	// pattern/rationale). source records WHERE this call came from
	// (models.BalanceSourceTransaction for a client Deposit/Withdrawal side
	// effect, models.BalanceSourceConfiguration for a direct manual admin
	// action) — explicit at the call site rather than inferred later from
	// the remark text.
	TopUpCredit(id uint, amount float64, remark string, createdByID uint, source models.BalanceTxSource, bonusAmount float64) (*models.ProductType, error)
	WithdrawCredit(id uint, amount float64, remark string, createdByID uint, source models.BalanceTxSource, bonusAmount float64) (*models.ProductType, error)
}

type productTypeRepository struct{ db *gorm.DB }

func NewProductTypeRepository(db *gorm.DB) ProductTypeRepository {
	return &productTypeRepository{db}
}

func (r *productTypeRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("Branch").Preload("CreatedBy").Preload("CurrencyType")
}

func (r *productTypeRepository) Create(x *models.ProductType, createdByID uint) error {
	x.CreatedByID = createdByID
	return r.db.Create(x).Error
}

// FindByID loads a product type by ID. If scopeIDs is non-nil, the record
// must belong to one of those branch IDs or ErrRecordNotFound is returned
// (so callers can't distinguish "doesn't exist" from "not in your scope").
func (r *productTypeRepository) FindByID(id uint, scopeIDs []uint) (*models.ProductType, error) {
	var x models.ProductType
	q := r.preload(r.db).Where("id = ?", id)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	if err := q.First(&x).Error; err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *productTypeRepository) List(showAll bool) ([]models.ProductType, error) {
	var items []models.ProductType
	q := r.preload(r.db).Model(&models.ProductType{})
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *productTypeRepository) Update(x *models.ProductType) error {
	return r.db.Save(x).Error
}

// applyCreditDelta mirrors CompanyBankRepository.applyCashDelta: the row is
// locked with SELECT ... FOR UPDATE for the duration of the transaction, so
// concurrent top-ups/withdrawals on the same product serialize instead of
// racing, and the ledger row's old/new amounts always reflect real states.
// bonusAmount is purely descriptive metadata recorded on the ledger entry
// (how much of amount was a Deposit's bonus) — it does NOT get added on
// top of amount here, since amount already includes it when relevant
// (callers combine principal+bonus into amount before calling this).
func (r *productTypeRepository) applyCreditDelta(id uint, amount float64, txType models.BalanceTxType, remark string, createdByID uint, source models.BalanceTxSource, bonusAmount float64) (*models.ProductType, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var pt models.ProductType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pt, id).Error; err != nil {
			return err
		}
		oldAmount := pt.Credit
		var newAmount float64
		if txType == models.BalanceTxWithdrawal {
			newAmount = oldAmount - amount
			if newAmount < 0 {
				return fmt.Errorf(
					"insufficient credit balance: this product has %.2f credit but the deposit requires %.2f",
					oldAmount, amount,
				)
			}
		} else {
			newAmount = oldAmount + amount
		}
		if err := tx.Model(&models.ProductType{}).Where("id = ?", id).UpdateColumn("credit", newAmount).Error; err != nil {
			return err
		}
		entry := &models.BalanceTransaction{
			EntityType:  models.BalanceEntityProductType,
			EntityID:    id,
			Field:       "credit",
			Type:        txType,
			Source:      source,
			OldAmount:   oldAmount,
			Amount:      amount,
			BonusAmount: bonusAmount,
			NewAmount:   newAmount,
			Remark:      remark,
			CreatedByID: createdByID,
		}
		return tx.Create(entry).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id, nil)
}

func (r *productTypeRepository) TopUpCredit(id uint, amount float64, remark string, createdByID uint, source models.BalanceTxSource, bonusAmount float64) (*models.ProductType, error) {
	return r.applyCreditDelta(id, amount, models.BalanceTxTopUp, remark, createdByID, source, bonusAmount)
}

func (r *productTypeRepository) WithdrawCredit(id uint, amount float64, remark string, createdByID uint, source models.BalanceTxSource, bonusAmount float64) (*models.ProductType, error) {
	return r.applyCreditDelta(id, amount, models.BalanceTxWithdrawal, remark, createdByID, source, bonusAmount)
}

// Delete removes a product type by ID. If scopeIDs is non-nil, the record
// must belong to one of those branch IDs, otherwise nothing is deleted and
// ErrRecordNotFound is returned.
func (r *productTypeRepository) Delete(id uint, scopeIDs []uint) error {
	if scopeIDs != nil && len(scopeIDs) == 0 {
		return gorm.ErrRecordNotFound
	}
	q := r.db.Where("id = ?", id)
	if scopeIDs != nil {
		q = q.Where("branch_id IN ?", scopeIDs)
	}
	res := q.Delete(&models.ProductType{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *productTypeRepository) ExistsByName(name string, excludeID uint) bool {
	var count int64
	q := r.db.Model(&models.ProductType{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	q.Count(&count)
	return count > 0
}

func (r *productTypeRepository) ListForUser(userID uint, showAll bool, branchID *uint) ([]models.ProductType, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		return []models.ProductType{}, nil
	}
	var items []models.ProductType
	q := r.preload(r.db).Model(&models.ProductType{})
	if !isSA {
		q = q.Where("branch_id IN ?", branchIDs)
	}
	if branchID != nil {
		// Intersected with the scope filter above: a non-SA user still can't
		// pull product types for a branch outside their own assigned branches.
		q = q.Where("branch_id = ?", *branchID)
	}
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access). parent_id is not used here — simple/sub users never inherit a
// parent's branches, they only see what's assigned to them directly.
func (r *productTypeRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
