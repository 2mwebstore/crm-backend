package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterRoleRoutes(rg *gin.RouterGroup, ctrl *controllers.RoleController, permCtrl *controllers.PermissionController, userRepo repositories.UserRepository) {
	// Permissions (read-only reference data — any authenticated user with role.view)
	perms := rg.Group("/permissions")
	perms.Use(middlewares.Auth())
	perms.Use(middlewares.RequirePermission(userRepo, models.PermRoleView))
	{
		perms.GET("", permCtrl.ListAll)
		perms.GET("/grouped", permCtrl.ListGrouped)
	}

	// Roles
	roles := rg.Group("/roles")
	roles.Use(middlewares.Auth())
	{
		// Anyone with role.view can list/get
		roles.GET("", middlewares.RequirePermission(userRepo, models.PermRoleView), ctrl.List)
		roles.GET("/:id", middlewares.RequirePermission(userRepo, models.PermRoleView), ctrl.GetByID)

		// Create: Super Admin → is_system=1 | Simple User → is_system=0, own role only
		roles.POST("", middlewares.RequirePermission(userRepo, models.PermRoleCreate), ctrl.Create)

		// Update/Delete: Super Admin can edit/delete any role
		//               Simple User can only edit/delete roles they created (service enforces this)
		roles.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermRoleEdit), ctrl.Update)
		roles.PUT("/:id/permissions", middlewares.RequirePermission(userRepo, models.PermRoleEdit), ctrl.AssignPermissions)
		roles.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermRoleDelete), ctrl.Delete)
	}
}
