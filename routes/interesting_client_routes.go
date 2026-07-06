package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterInterestingClientRoutes(rg *gin.RouterGroup, ctrl *controllers.InterestingClientController, userRepo repositories.UserRepository) {
	ic := rg.Group("/interesting-clients")
	ic.Use(middlewares.Auth())
	{
		// Core CRUD
		ic.GET("", middlewares.RequirePermission(userRepo, models.PermICView), ctrl.List)
		ic.GET("/check-code", middlewares.RequirePermission(userRepo, models.PermICView), ctrl.CheckCode)
		ic.GET("/next-code", middlewares.RequirePermission(userRepo, models.PermICView), ctrl.PreviewCode)
		ic.GET("/preview-code", middlewares.RequirePermission(userRepo, models.PermICView), ctrl.PreviewCode)
		ic.POST("", middlewares.RequirePermission(userRepo, models.PermICCreate), ctrl.Create)
		ic.GET("/:id", middlewares.RequirePermission(userRepo, models.PermICView), ctrl.GetByID)
		ic.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermICEdit), ctrl.Update)
		ic.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermICDelete), ctrl.Delete)

		// Convert → Client
		ic.POST("/:id/convert", middlewares.RequirePermission(userRepo, models.PermICConvert), ctrl.Convert)

		// Phone sub-resource management
		phones := ic.Group("/:id/phones")
		phones.Use(middlewares.RequirePermission(userRepo, models.PermICEdit))
		{
			phones.PUT("/:sub_id", ctrl.UpdatePhone)
			phones.DELETE("/:sub_id", ctrl.DeletePhone)
		}
	}
}
