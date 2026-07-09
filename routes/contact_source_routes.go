package routes

import (
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterContactSourceRoutes(rg *gin.RouterGroup, ctrl *controllers.ContactSourceController, userRepo repositories.UserRepository) {
	view := middlewares.RequirePermission(userRepo,
		models.PermContactSourceView, models.PermContactSourceCreate, models.PermContactSourceEdit, models.PermContactSourceDelete)
	create := middlewares.RequirePermission(userRepo, models.PermContactSourceCreate)
	edit := middlewares.RequirePermission(userRepo, models.PermContactSourceEdit)
	del := middlewares.RequirePermission(userRepo, models.PermContactSourceDelete)

	g := rg.Group("/contact-sources")
	g.Use(middlewares.Auth())
	{
		g.GET("", view, ctrl.List)
		g.GET("/:id", view, ctrl.GetByID)
		g.POST("", create, ctrl.Create)
		g.PUT("/:id", edit, ctrl.Update)
		g.DELETE("/:id", del, ctrl.Delete)
	}
}
