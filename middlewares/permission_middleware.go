package middlewares

import (
	"crm-backend/repositories"
	"crm-backend/utils"

	"github.com/gin-gonic/gin"
)

// RequirePermission checks that the authenticated user's role contains
// the given permission string. Super admins always pass.
//
// Usage in routes:
//
//	clients.GET("", middlewares.RequirePermission(userRepo, models.PermClientView), ctrl.List)
func RequirePermission(userRepo repositories.UserRepository, perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Super admin bypasses all permission checks
		if IsSuperAdmin(c) {
			c.Next()
			return
		}
		userID := GetUserID(c)
		if userID == 0 {
			utils.Unauthorized(c, "not authenticated")
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(userID)
		if err != nil || !user.IsActive {
			utils.Unauthorized(c, "user not found or inactive")
			c.Abort()
			return
		}

		hasAny := false
		for _, perm := range perms {
			if user.HasPermission(perm) {
				hasAny = true
				break
			}
		}
		if !hasAny {
			utils.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyPermission passes if the user has AT LEAST ONE of the given permissions.
func RequireAnyPermission(userRepo repositories.UserRepository, perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			utils.Unauthorized(c, "not authenticated")
			c.Abort()
			return
		}
		user, err := userRepo.FindByID(userID)
		if err != nil || !user.IsActive {
			utils.Unauthorized(c, "user not found or inactive")
			c.Abort()
			return
		}
		for _, p := range perms {
			if user.HasPermission(p) {
				c.Next()
				return
			}
		}
		utils.Forbidden(c)
		c.Abort()
	}
}
