package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
)

func RegisterBranchRoutes(rg *gin.RouterGroup, ctrl *controllers.BranchController) {
	g := rg.Group("/branches")
	g.Use(middlewares.Auth())
	{
		// All authenticated users can view branches (for dropdowns)
		g.GET("", ctrl.List)
		g.GET("/:id", ctrl.GetByID)
		// Only super admin can manage branches
		g.POST("", middlewares.RequireSuperAdmin(), ctrl.Create)
		g.PUT("/:id", middlewares.RequireSuperAdmin(), ctrl.Update)
		g.DELETE("/:id", middlewares.RequireSuperAdmin(), ctrl.Delete)
	}
}
