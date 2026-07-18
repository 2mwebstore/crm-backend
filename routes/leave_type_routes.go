package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterLeaveTypeRoutes(rg *gin.RouterGroup, ctrl *controllers.LeaveTypeController, userRepo repositories.UserRepository) {
	g := rg.Group("/leave-types")
	g.Use(middlewares.Auth())
	{
		g.GET("", middlewares.RequirePermission(userRepo, models.PermLeaveTypeView), ctrl.List)
		g.GET("/:id", middlewares.RequirePermission(userRepo, models.PermLeaveTypeView), ctrl.GetByID)
		g.POST("", middlewares.RequirePermission(userRepo, models.PermLeaveTypeCreate), ctrl.Create)
		g.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermLeaveTypeEdit), ctrl.Update)
		g.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermLeaveTypeDelete), ctrl.Delete)
	}
}
