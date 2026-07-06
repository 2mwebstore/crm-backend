package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterTransactionRoutes(
	rg *gin.RouterGroup,
	depositCtrl *controllers.DepositController,
	withdrawalCtrl *controllers.WithdrawalController,
	userRepo repositories.UserRepository,
) {
	auth := middlewares.Auth()

	dep := rg.Group("/deposits")
	dep.Use(auth)
	{
		dep.GET("", middlewares.RequirePermission(userRepo, models.PermDepositView), depositCtrl.List)
		dep.GET("/balance", middlewares.RequirePermission(userRepo, models.PermDepositView), depositCtrl.GetBalance)
		dep.POST("", middlewares.RequirePermission(userRepo, models.PermDepositCreate), depositCtrl.Create)
		dep.GET("/:id", middlewares.RequirePermission(userRepo, models.PermDepositView), depositCtrl.GetByID)
		dep.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermDepositEdit), depositCtrl.Update)
		dep.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermDepositDelete), depositCtrl.Delete)
		dep.PUT("/:id/approve", middlewares.RequirePermission(userRepo, models.PermDepositApprove), depositCtrl.Approve)
	}

	wdr := rg.Group("/withdrawals")
	wdr.Use(auth)
	{
		wdr.GET("", middlewares.RequirePermission(userRepo, models.PermWithdrawalView), withdrawalCtrl.List)
		wdr.GET("/balance", middlewares.RequirePermission(userRepo, models.PermWithdrawalView), withdrawalCtrl.GetBalance)
		wdr.POST("", middlewares.RequirePermission(userRepo, models.PermWithdrawalCreate), withdrawalCtrl.Create)
		wdr.GET("/:id", middlewares.RequirePermission(userRepo, models.PermWithdrawalView), withdrawalCtrl.GetByID)
		wdr.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermWithdrawalEdit), withdrawalCtrl.Update)
		wdr.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermWithdrawalDelete), withdrawalCtrl.Delete)
		wdr.PUT("/:id/approve", middlewares.RequirePermission(userRepo, models.PermWithdrawalApprove), withdrawalCtrl.Approve)
	}
}
