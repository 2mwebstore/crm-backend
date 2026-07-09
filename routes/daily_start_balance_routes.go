package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
)

func RegisterDailyStartBalanceRoutes(rg *gin.RouterGroup, ctrl *controllers.DailyStartBalanceController) {
	g := rg.Group("/daily-balances")
	g.Use(middlewares.Auth())
	{
		g.POST("/start", ctrl.StartToday)
		g.POST("/close", ctrl.CloseToday)
		g.GET("/today", ctrl.Today)
		g.GET("/history", ctrl.History)
	}
}
