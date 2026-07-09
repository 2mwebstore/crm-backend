package routes

import (
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterLevelRoutes(rg *gin.RouterGroup, ctrl *controllers.LevelController, userRepo repositories.UserRepository) {
	view := middlewares.RequirePermission(userRepo,
		models.PermLevelView, models.PermLevelCreate, models.PermLevelEdit, models.PermLevelDelete)
	create := middlewares.RequirePermission(userRepo, models.PermLevelCreate)
	edit := middlewares.RequirePermission(userRepo, models.PermLevelEdit)
	del := middlewares.RequirePermission(userRepo, models.PermLevelDelete)

	g := rg.Group("/levels")
	g.Use(middlewares.Auth())
	{
		g.GET("", view, ctrl.List)
		g.GET("/:id", view, ctrl.GetByID)
		g.POST("", create, ctrl.Create)
		g.PUT("/:id", edit, ctrl.Update)
		g.DELETE("/:id", del, ctrl.Delete)
	}
}
