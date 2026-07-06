package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterClientRoutes(rg *gin.RouterGroup, ctrl *controllers.ClientController, userRepo repositories.UserRepository) {
	clients := rg.Group("/clients")
	clients.Use(middlewares.Auth())
	{
		// Core CRUD
		clients.GET("", middlewares.RequirePermission(userRepo, models.PermClientView), ctrl.List)
		clients.GET("/check-code", middlewares.RequirePermission(userRepo, models.PermClientView), ctrl.CheckCode)
		clients.GET("/next-code", middlewares.RequirePermission(userRepo, models.PermClientView), ctrl.NextCode)
		clients.POST("", middlewares.RequirePermission(userRepo, models.PermClientCreate), ctrl.Create)
		clients.GET("/:id", middlewares.RequirePermission(userRepo, models.PermClientView), ctrl.GetByID)
		clients.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermClientEdit), ctrl.Update)
		clients.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermClientDelete), ctrl.Delete)

		// Picture upload
		clients.POST("/:id/picture", middlewares.RequirePermission(userRepo, models.PermClientEdit), ctrl.UploadPicture)

		// Bank Section
		banks := clients.Group("/:id/banks")
		banks.Use(middlewares.RequirePermission(userRepo, models.PermClientEdit))
		{
			banks.POST("", ctrl.AddBank)
			banks.PUT("/:sub_id", ctrl.UpdateBank)
			banks.DELETE("/:sub_id", ctrl.DeleteBank)
		}

		// Product (Player) Section
		products := clients.Group("/:id/products")
		products.Use(middlewares.RequirePermission(userRepo, models.PermClientEdit))
		{
			products.POST("", ctrl.AddProduct)
			products.PUT("/:sub_id", ctrl.UpdateProduct)
			products.DELETE("/:sub_id", ctrl.DeleteProduct)
		}

		// Follow Up Section
		followups := clients.Group("/:id/follow-ups")
		followups.Use(middlewares.RequirePermission(userRepo, models.PermClientEdit))
		{
			followups.GET("", middlewares.RequirePermission(userRepo, models.PermClientView), ctrl.ListFollowUps)
			followups.POST("", ctrl.AddFollowUp)
			followups.DELETE("/:sub_id", ctrl.DeleteFollowUp)
		}
	}
}
