package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, ctrl *controllers.AuthController) {
	auth := rg.Group("/auth")
	{
		// Public
		auth.POST("/register", ctrl.Register)
		auth.POST("/login", ctrl.Login)

		// Protected
		protected := auth.Group("")
		protected.Use(middlewares.Auth())
		{
			protected.GET("/me", ctrl.GetProfile)
			protected.PUT("/profile", ctrl.UpdateProfile)
			protected.POST("/change-password", ctrl.ChangePassword)
		}
	}
}
