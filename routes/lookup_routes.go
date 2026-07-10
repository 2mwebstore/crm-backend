package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterLookupRoutes(
	rg *gin.RouterGroup,
	bankTypeCtrl *controllers.BankTypeController,
	CompanybankCtrl *controllers.CompanyBankController,
	productTypeCtrl *controllers.ProductTypeController,
	bonusOptionCtrl *controllers.BonusOptionTypeController,
	currencyTypeCtrl *controllers.CurrencyTypeController,
	userRepo repositories.UserRepository,
	BalanceTransactionCtrl *controllers.BalanceTransactionController,
) {
	auth := middlewares.Auth()

	// Separate, more granular permission for the Top Up / Withdraw balance
	// actions specifically — a role can be granted record CRUD without
	// this, or this without record CRUD, or both.
	companyBankBalance := middlewares.RequirePermission(userRepo, models.PermCompanyBankTopup)
	productTypeBalance := middlewares.RequirePermission(userRepo, models.PermProductTypeTopup)
	companyBankAdjustment := middlewares.RequirePermission(userRepo, models.PermCompanyBankAdjustment)
	productTypeAdjustment := middlewares.RequirePermission(userRepo, models.PermProductTypeAdjustment)

	// Per-entity View/Create/Edit/Delete permissions — each table is fully
	// self-contained now (no more blanket lookup.view/lookup.manage/
	// configuration.view/configuration.manage fallback). View additionally
	// accepts that same table's own Create/Edit/Delete — if you can manage
	// a table's records you can naturally see them too.
	bankTypeView := middlewares.RequirePermission(userRepo,
		models.PermBankTypeView, models.PermBankTypeCreate, models.PermBankTypeEdit, models.PermBankTypeDelete)
	bankTypeCreate := middlewares.RequirePermission(userRepo, models.PermBankTypeCreate)
	bankTypeEdit := middlewares.RequirePermission(userRepo, models.PermBankTypeEdit)
	bankTypeDelete := middlewares.RequirePermission(userRepo, models.PermBankTypeDelete)

	companyBankView := middlewares.RequirePermission(userRepo,
		models.PermCompanyBankView, models.PermCompanyBankCreate, models.PermCompanyBankEdit, models.PermCompanyBankDelete)
	companyBankCreate := middlewares.RequirePermission(userRepo, models.PermCompanyBankCreate)
	companyBankEdit := middlewares.RequirePermission(userRepo, models.PermCompanyBankEdit)
	companyBankDelete := middlewares.RequirePermission(userRepo, models.PermCompanyBankDelete)

	productTypeView := middlewares.RequirePermission(userRepo,
		models.PermProductTypeView, models.PermProductTypeCreate, models.PermProductTypeEdit, models.PermProductTypeDelete)
	productTypeCreate := middlewares.RequirePermission(userRepo, models.PermProductTypeCreate)
	productTypeEdit := middlewares.RequirePermission(userRepo, models.PermProductTypeEdit)
	productTypeDelete := middlewares.RequirePermission(userRepo, models.PermProductTypeDelete)

	bonusOptionView := middlewares.RequirePermission(userRepo,
		models.PermBonusOptionView, models.PermBonusOptionCreate, models.PermBonusOptionEdit, models.PermBonusOptionDelete)
	bonusOptionCreate := middlewares.RequirePermission(userRepo, models.PermBonusOptionCreate)
	bonusOptionEdit := middlewares.RequirePermission(userRepo, models.PermBonusOptionEdit)
	bonusOptionDelete := middlewares.RequirePermission(userRepo, models.PermBonusOptionDelete)

	currencyView := middlewares.RequirePermission(userRepo,
		models.PermCurrencyView, models.PermCurrencyCreate, models.PermCurrencyEdit, models.PermCurrencyDelete)
	currencyCreate := middlewares.RequirePermission(userRepo, models.PermCurrencyCreate)
	currencyEdit := middlewares.RequirePermission(userRepo, models.PermCurrencyEdit)
	currencyDelete := middlewares.RequirePermission(userRepo, models.PermCurrencyDelete)

	// RegisterBalanceTransactionRoutes exposes read-only access to the shared
	// balance ledger (CompanyBank.Cash and ProductType.Credit top-up/withdrawal
	// history). No create/update/delete routes — rows are only ever written
	// internally by TopUpCash/WithdrawCash/TopUpCredit/WithdrawCredit.
	txs := rg.Group("/balance-transactions")
	txs.Use(auth)
	{
		txs.GET("", BalanceTransactionCtrl.List)
	}

	// ── Bank Types ─────────────────────────────────────────────────────────
	bt := rg.Group("/bank-types")
	bt.Use(auth)
	{
		bt.GET("", bankTypeView, bankTypeCtrl.List)
		bt.GET("/:id", bankTypeView, bankTypeCtrl.GetByID)
		bt.POST("", bankTypeCreate, bankTypeCtrl.Create)
		bt.PUT("/:id", bankTypeEdit, bankTypeCtrl.Update)
		bt.DELETE("/:id", bankTypeDelete, bankTypeCtrl.Delete)
	}

	// ── Company Banks ──────────────────────────────────────────────────────
	cb := rg.Group("/company-banks")
	cb.Use(auth)
	{
		cb.GET("", companyBankView, CompanybankCtrl.List)
		cb.GET("/:id", companyBankView, CompanybankCtrl.GetByID)
		cb.POST("", companyBankCreate, CompanybankCtrl.Create)
		cb.PUT("/:id", companyBankEdit, CompanybankCtrl.Update)
		cb.DELETE("/:id", companyBankDelete, CompanybankCtrl.Delete)
		cb.POST("/:id/topup", companyBankBalance, CompanybankCtrl.TopUpCash)
		cb.POST("/:id/withdraw", companyBankBalance, CompanybankCtrl.WithdrawCash)
		cb.POST("/:id/adjust", companyBankAdjustment, CompanybankCtrl.Adjust)
	}

	// ── Product Types ──────────────────────────────────────────────────────
	pt := rg.Group("/product-types")
	pt.Use(auth)
	{
		pt.GET("", productTypeView, productTypeCtrl.List)
		pt.GET("/:id", productTypeView, productTypeCtrl.GetByID)
		pt.POST("", productTypeCreate, productTypeCtrl.Create)
		pt.PUT("/:id", productTypeEdit, productTypeCtrl.Update)
		pt.DELETE("/:id", productTypeDelete, productTypeCtrl.Delete)
		pt.POST("/:id/topup", productTypeBalance, productTypeCtrl.TopUpCredit)
		pt.POST("/:id/withdraw", productTypeBalance, productTypeCtrl.WithdrawCredit)
		pt.POST("/:id/adjust", productTypeAdjustment, productTypeCtrl.Adjust)
	}

	// ── Bonus Option Types ─────────────────────────────────────────────────
	bo := rg.Group("/bonus-option-types")
	bo.Use(auth)
	{
		bo.GET("", bonusOptionView, bonusOptionCtrl.List)
		bo.GET("/:id", bonusOptionView, bonusOptionCtrl.GetByID)
		bo.POST("", bonusOptionCreate, bonusOptionCtrl.Create)
		bo.PUT("/:id", bonusOptionEdit, bonusOptionCtrl.Update)
		bo.DELETE("/:id", bonusOptionDelete, bonusOptionCtrl.Delete)
	}

	// ── Currency Types ─────────────────────────────────────────────────────
	ct := rg.Group("/currency-types")
	ct.Use(auth)
	{
		ct.GET("", currencyView, currencyTypeCtrl.List)
		ct.GET("/:id", currencyView, currencyTypeCtrl.GetByID)
		ct.POST("", currencyCreate, currencyTypeCtrl.Create)
		ct.PUT("/:id", currencyEdit, currencyTypeCtrl.Update)
		ct.DELETE("/:id", currencyDelete, currencyTypeCtrl.Delete)
	}
}
