package routes

import (
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterLevelRoutes(rg *gin.RouterGroup, ctrl *controllers.LevelController, userRepo repositories.UserRepository) {
	g := rg.Group("/levels")
	g.Use(middlewares.Auth())
	{
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.GetByID)
		g.POST("", middlewares.RequirePermission(userRepo, models.PermLookupManage, models.PermConfigManage), ctrl.Create)
		g.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermLookupManage, models.PermConfigManage), ctrl.Update)
		g.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermLookupManage, models.PermConfigManage), ctrl.Delete)
	}
}
