package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterUserRoutes(rg *gin.RouterGroup, ctrl *controllers.UserController, userRepo repositories.UserRepository) {
	users := rg.Group("/users")
	users.Use(middlewares.Auth())
	{
		// ── Super Admin only ──────────────────────────────────────────────────
		admin := users.Group("/admin")
		admin.Use(middlewares.RequireSuperAdmin())
		{
			admin.GET("", ctrl.AdminListUsers)
			admin.POST("", ctrl.AdminCreateUser)
			admin.PUT("/:id", ctrl.AdminUpdateUser)
			admin.DELETE("/:id", ctrl.AdminDeleteUser)
			admin.PUT("/:id/branches", ctrl.AssignBranches)
		}

		// ── Scope helpers ────────────────────────────────────────────────────
		users.GET("/in-scope", ctrl.ListUsersInScope)

		// ── Simple / Sub-user management (permission-gated) ───────────────────
		sub := users.Group("/sub-users")
		{
			sub.GET("", middlewares.RequirePermission(userRepo, models.PermUserView), ctrl.ListSubUsers)
			sub.POST("", middlewares.RequirePermission(userRepo, models.PermUserCreate), ctrl.CreateSubUser)
			sub.GET("/:id", middlewares.RequirePermission(userRepo, models.PermUserView), ctrl.GetSubUser)
			sub.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermUserEdit), ctrl.UpdateSubUser)
			sub.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermUserDelete), ctrl.DeleteSubUser)
		}
	}
}
