package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterScheduleOverrideRoutes(rg *gin.RouterGroup, ctrl *controllers.UserScheduleOverrideController, userRepo repositories.UserRepository) {
	g := rg.Group("/schedule-overrides")
	g.Use(middlewares.Auth())
	{
		g.GET("", middlewares.RequirePermission(userRepo, models.PermScheduleOverrideView), ctrl.List)
		g.GET("/for-user", middlewares.RequirePermission(userRepo, models.PermScheduleOverrideView), ctrl.ListForUser)
		g.POST("", middlewares.RequirePermission(userRepo, models.PermScheduleOverrideCreate), ctrl.Create)
		g.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermScheduleOverrideEdit), ctrl.Update)
		g.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermScheduleOverrideDelete), ctrl.Delete)
	}
}
