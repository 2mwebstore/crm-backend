package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"

	"crm-backend/config"
	"crm-backend/utils"
)

const (
	CtxUserID = "userID"
	CtxEmail  = "userEmail"
	CtxRole        = "userRole"
	CtxSuperAdmin  = "isSuperAdmin"
)

// Auth validates the Bearer JWT and injects claims into Gin context.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()

		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Unauthorized(c, "authorization must be 'Bearer <token>'")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1], cfg.JWT.Secret)
		if err != nil {
			utils.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxEmail, claims.Email)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxSuperAdmin, claims.IsSuperAdmin)
		c.Next()
	}
}

// RequireRoles restricts access to users whose role is in the allowlist.
// Super admins always bypass this check.
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		// Super admin bypasses all role checks
		if IsSuperAdmin(c) {
			c.Next()
			return
		}
		role := GetRole(c)
		if _, ok := allowed[role]; !ok {
			utils.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// IsSuperAdmin checks whether the current user is a super admin from JWT context.
func IsSuperAdmin(c *gin.Context) bool {
	if v, ok := c.Get(CtxSuperAdmin); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetUserID extracts the authenticated user's ID from context.
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetRole extracts the authenticated user's role from context.
func GetRole(c *gin.Context) string {
	if v, ok := c.Get(CtxRole); ok {
		if r, ok := v.(string); ok {
			return r
		}
	}
	return ""
}

// RequireSuperAdmin middleware — only super admins can proceed.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsSuperAdmin(c) {
			utils.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
