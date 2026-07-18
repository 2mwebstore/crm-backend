package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterActivityRequestRoutes(rg *gin.RouterGroup, ctrl *controllers.ActivityRequestController, userRepo repositories.UserRepository) {
	g := rg.Group("/activity-requests")
	g.Use(middlewares.Auth())
	{
		// Submitting your own Activity requests requires
		// activity_requests.request — a sibling to View below, not a
		// prerequisite for it.
		g.POST("", middlewares.RequirePermission(userRepo, models.PermActivityRequestSubmit), ctrl.Create)
		g.GET("/mine", middlewares.RequirePermission(userRepo, models.PermActivityRequestSubmit), ctrl.Mine)
		g.GET("", middlewares.RequirePermission(userRepo, models.PermActivityRequestView), ctrl.List)
		g.GET("/:id", middlewares.RequirePermission(userRepo, models.PermActivityRequestView), ctrl.GetByID)
	}
}
