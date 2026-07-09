package routes

import (
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(rg *gin.RouterGroup, ctrl *controllers.ReportController, userRepo repositories.UserRepository) {
	reports := rg.Group("/reports")
	reports.Use(middlewares.Auth(), middlewares.RequirePermission(userRepo, models.PermReportView))
	{
		reports.GET("/summary", ctrl.Summary)
		reports.GET("/transactions", ctrl.AllTransactions)
		reports.GET("/bank-summary", ctrl.BankSummary)
	}

}
