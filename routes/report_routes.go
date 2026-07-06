package routes

import (
	"github.com/gin-gonic/gin"
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterReportRoutes(rg *gin.RouterGroup, ctrl *controllers.ReportController, userRepo repositories.UserRepository) {
	reports := rg.Group("/reports")
	reports.Use(middlewares.Auth(), middlewares.RequirePermission(userRepo, models.PermReportView))
	{
		reports.GET("/summary", ctrl.Summary)
	}
}
