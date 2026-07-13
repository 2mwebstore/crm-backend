package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterAuditLogRoutes(rg *gin.RouterGroup, ctrl *controllers.AuditLogController, userRepo repositories.UserRepository) {
	g := rg.Group("/audit-logs")
	g.Use(middlewares.Auth())
	{
		g.GET("", middlewares.RequirePermission(userRepo, models.PermAuditLogView), ctrl.List)
		g.DELETE("", middlewares.RequirePermission(userRepo, models.PermAuditLogDelete), ctrl.DeleteOld)
	}
}
