package routes

import (
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterFollowUpRoutes(rg *gin.RouterGroup, ctrl *controllers.FollowUpController, userRepo repositories.UserRepository) {
	g := rg.Group("/follow-ups")
	g.Use(middlewares.Auth())
	{
		g.GET("", middlewares.RequirePermission(userRepo, models.PermFollowUpView), ctrl.List)
		g.POST("", middlewares.RequirePermission(userRepo, models.PermFollowUpCreate), ctrl.Create)
		g.GET("/:id", middlewares.RequirePermission(userRepo, models.PermFollowUpView), ctrl.GetByID)
		g.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermFollowUpDelete), ctrl.Delete)
	}
}
