package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterDailyStartBalanceRoutes(rg *gin.RouterGroup, ctrl *controllers.DailyStartBalanceController, userRepo repositories.UserRepository) {
	// View also accepts Start/Close — being able to act on a shift
	// naturally implies being able to see it, matching the pattern used
	// for View/Create/Edit/Delete elsewhere in this app.
	view := middlewares.RequirePermission(userRepo,
		models.PermDailyBalanceView, models.PermDailyBalanceStart, models.PermDailyBalanceClose)
	start := middlewares.RequirePermission(userRepo, models.PermDailyBalanceStart)
	closePerm := middlewares.RequirePermission(userRepo, models.PermDailyBalanceClose)

	g := rg.Group("/daily-balances")
	g.Use(middlewares.Auth())
	{
		g.POST("/start", start, ctrl.StartToday)
		g.POST("/close", closePerm, ctrl.CloseToday)
		g.GET("/today", view, ctrl.Today)
		g.GET("/history", view, ctrl.History)
		g.GET("/:id/balance-transactions", view, ctrl.ShiftBalanceTransactions)
	}
}
