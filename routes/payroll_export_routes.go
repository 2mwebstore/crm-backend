package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterPayrollExportRoutes(rg *gin.RouterGroup, ctrl *controllers.PayrollExportController, userRepo repositories.UserRepository) {
	g := rg.Group("/attendance")
	g.Use(middlewares.Auth())
	{
		g.GET("/payroll-export", middlewares.RequirePermission(userRepo, models.PermAttendanceView), ctrl.Export)
	}
}
