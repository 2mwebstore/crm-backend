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
	productTypeCtrl *controllers.ProductTypeController,
	bonusOptionCtrl *controllers.BonusOptionTypeController,
	currencyTypeCtrl *controllers.CurrencyTypeController,
	userRepo repositories.UserRepository,
) {
	auth := middlewares.Auth()
	manage := middlewares.RequirePermission(userRepo, models.PermLookupManage, models.PermConfigManage)

	// ── Bank Types ─────────────────────────────────────────────────────────
	bt := rg.Group("/bank-types")
	bt.Use(auth)
	{
		bt.GET("", bankTypeCtrl.List)
		bt.GET("/:id", bankTypeCtrl.GetByID)
		bt.POST("", manage, bankTypeCtrl.Create)
		bt.PUT("/:id", manage, bankTypeCtrl.Update)
		bt.DELETE("/:id", manage, bankTypeCtrl.Delete)
	}

	// ── Product Types ──────────────────────────────────────────────────────
	pt := rg.Group("/product-types")
	pt.Use(auth)
	{
		pt.GET("", productTypeCtrl.List)
		pt.GET("/:id", productTypeCtrl.GetByID)
		pt.POST("", manage, productTypeCtrl.Create)
		pt.PUT("/:id", manage, productTypeCtrl.Update)
		pt.DELETE("/:id", manage, productTypeCtrl.Delete)
	}

	// ── Bonus Option Types ─────────────────────────────────────────────────
	bo := rg.Group("/bonus-option-types")
	bo.Use(auth)
	{
		bo.GET("", bonusOptionCtrl.List)
		bo.GET("/:id", bonusOptionCtrl.GetByID)
		bo.POST("", manage, bonusOptionCtrl.Create)
		bo.PUT("/:id", manage, bonusOptionCtrl.Update)
		bo.DELETE("/:id", manage, bonusOptionCtrl.Delete)
	}

	// ── Currency Types ─────────────────────────────────────────────────────
	ct := rg.Group("/currency-types")
	ct.Use(auth)
	{
		ct.GET("", currencyTypeCtrl.List)
		ct.GET("/:id", currencyTypeCtrl.GetByID)
		ct.POST("", manage, currencyTypeCtrl.Create)
		ct.PUT("/:id", manage, currencyTypeCtrl.Update)
		ct.DELETE("/:id", manage, currencyTypeCtrl.Delete)
	}

}
