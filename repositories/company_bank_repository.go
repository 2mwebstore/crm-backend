package repositories

import (
	"errors"
	"fmt"

	"crm-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CompanyBankRepository interface {
	Create(x *models.CompanyBank, createdByID uint) error
	FindByID(id uint, scopeIDs []uint) (*models.CompanyBank, error)
	ListForUser(userID uint, showAll bool, branchID *uint) ([]models.CompanyBank, error)
	List(showAll bool) ([]models.CompanyBank, error)
	Update(x *models.CompanyBank) error
	Delete(id uint, scopeIDs []uint) error
	// TopUpCash and WithdrawCash apply an atomic, row-locked balance change
	// and write a matching BalanceTransaction ledger row in the same DB
	// transaction, so the audit trail and the actual balance can never
	// drift apart even under concurrent requests.
	TopUpCash(id uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error)
	WithdrawCash(id uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error)
}

type companyBankRepository struct{ db *gorm.DB }

func NewCompanyBankRepository(db *gorm.DB) CompanyBankRepository {
	return &companyBankRepository{db}
}

func (r *companyBankRepository) preload(q *gorm.DB) *gorm.DB {
	return q.Preload("BankType").Preload("CurrencyType").Preload("Branch").Preload("CreatedBy")
}

func (r *companyBankRepository) Create(x *models.CompanyBank, createdByID uint) error {
	x.CreatedByID = createdByID
	return r.db.Create(x).Error
}

// FindByID loads a company bank account by ID. If scopeIDs is non-nil, the
// record must belong to one of those branch IDs or ErrRecordNotFound is
// returned (so callers can't distinguish "doesn't exist" from "not in your
// scope"). Company-wide accounts (branch_id IS NULL) are always visible
// regardless of scope, since they aren't tied to any one branch.
func (r *companyBankRepository) FindByID(id uint, scopeIDs []uint) (*models.CompanyBank, error) {
	var x models.CompanyBank
	q := r.preload(r.db).Where("id = ?", id)
	if scopeIDs != nil {
		if len(scopeIDs) == 0 {
			q = q.Where("branch_id IS NULL")
		} else {
			q = q.Where("branch_id IN ? OR branch_id IS NULL", scopeIDs)
		}
	}
	if err := q.First(&x).Error; err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *companyBankRepository) List(showAll bool) ([]models.CompanyBank, error) {
	var items []models.CompanyBank
	q := r.preload(r.db).Model(&models.CompanyBank{})
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, account_name ASC").Find(&items).Error
	return items, err
}

func (r *companyBankRepository) Update(x *models.CompanyBank) error {
	return r.db.Save(x).Error
}

// applyCashDelta performs the balance change and ledger write inside a
// single DB transaction, with the CompanyBank row locked via SELECT ... FOR
// UPDATE (clause.Locking) for the transaction's duration. Locking the row
// means a second concurrent top-up/withdrawal on the same account has to
// wait for this one to commit, so old_amount always reflects a real
// pre-transaction state — no lost updates, no ledger drift.
func (r *companyBankRepository) applyCashDelta(id uint, amount float64, txType models.BalanceTxType, remark string, createdByID uint) (*models.CompanyBank, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var cb models.CompanyBank
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cb, id).Error; err != nil {
			return err
		}
		oldAmount := cb.Cash
		var newAmount float64
		if txType == models.BalanceTxWithdrawal {
			newAmount = oldAmount - amount
			if newAmount < 0 {
				return fmt.Errorf(
					"insufficient cash balance: this company bank has %.2f but the withdrawal requires %.2f",
					oldAmount, amount,
				)
			}
		} else {
			newAmount = oldAmount + amount
		}
		if err := tx.Model(&models.CompanyBank{}).Where("id = ?", id).UpdateColumn("cash", newAmount).Error; err != nil {
			return err
		}
		entry := &models.BalanceTransaction{
			EntityType:  models.BalanceEntityCompanyBank,
			EntityID:    id,
			Field:       "cash",
			Type:        txType,
			OldAmount:   oldAmount,
			Amount:      amount,
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

func (r *companyBankRepository) TopUpCash(id uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error) {
	return r.applyCashDelta(id, amount, models.BalanceTxTopUp, remark, createdByID)
}

func (r *companyBankRepository) WithdrawCash(id uint, amount float64, remark string, createdByID uint) (*models.CompanyBank, error) {
	return r.applyCashDelta(id, amount, models.BalanceTxWithdrawal, remark, createdByID)
}

// Delete removes a company bank account by ID. If scopeIDs is non-nil, the
// record must belong to one of those branch IDs, otherwise nothing is
// deleted and ErrRecordNotFound is returned.
func (r *companyBankRepository) Delete(id uint, scopeIDs []uint) error {
	if scopeIDs != nil && len(scopeIDs) == 0 {
		return gorm.ErrRecordNotFound
	}
	q := r.db.Where("id = ?", id)
	if scopeIDs != nil {
		q = q.Where("branch_id IN ? OR branch_id IS NULL", scopeIDs)
	}
	res := q.Delete(&models.CompanyBank{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *companyBankRepository) ListForUser(userID uint, showAll bool, branchID *uint) ([]models.CompanyBank, error) {
	branchIDs, isSA := r.resolveUserBranches(userID)
	if !isSA && len(branchIDs) == 0 {
		// Even with no branch of their own, a non-SA user can still see
		// company-wide accounts (branch_id IS NULL).
		var items []models.CompanyBank
		q := r.preload(r.db).Model(&models.CompanyBank{}).Where("branch_id IS NULL")
		if !showAll {
			q = q.Where("is_active = ?", true)
		}
		err := q.Order("sort_order ASC, account_name ASC").Find(&items).Error
		return items, err
	}

	var items []models.CompanyBank
	q := r.preload(r.db).Model(&models.CompanyBank{})
	if !isSA {
		q = q.Where("branch_id IN ? OR branch_id IS NULL", branchIDs)
	}
	if branchID != nil {
		// Intersected with the scope filter above: a non-SA user still can't
		// pull accounts for a branch outside their own assigned branches.
		q = q.Where("branch_id = ?", *branchID)
	}
	if !showAll {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order ASC, account_name ASC").Find(&items).Error
	return items, err
}

// resolveUserBranches returns the caller's own branch scope.
// Returns nil = SA (no filter, sees everything even with a parent_id set),
// []uint = the caller's own directly-assigned branches (may be empty = no
// access beyond company-wide accounts). parent_id is not used here —
// simple/sub users never inherit a parent's branches, they only see what's
// assigned to them directly.
func (r *companyBankRepository) resolveUserBranches(userID uint) ([]uint, bool) {
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
