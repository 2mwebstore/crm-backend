package controllers

import (
	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type PermissionController struct {
	svc      services.PermissionService
	userRepo repositories.UserRepository
}

func NewPermissionController(svc services.PermissionService, userRepo repositories.UserRepository) *PermissionController {
	return &PermissionController{svc, userRepo}
}

// ListAll godoc
func (ctrl *PermissionController) ListAll(c *gin.Context) {
	list, err := ctrl.svc.ListAll()
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	// Filter to caller's own permissions for non-SA users
	list = ctrl.filterToCallerPerms(c, list)
	utils.OK(c, "success", list)
}

// ListGrouped godoc — returns permissions grouped by resource.
// Super Admin gets ALL permissions.
// Simple/Sub users get ONLY permissions their own role has.
func (ctrl *PermissionController) ListGrouped(c *gin.Context) {
	grouped, err := ctrl.svc.ListGrouped()
	if err != nil {
		utils.InternalError(c, err)
		return
	}

	// Super Admin sees all permissions (for reference / super admin panel)
	if middlewares.IsSuperAdmin(c) {
		utils.OK(c, "success", grouped)
		return
	}

	// Regular users: filter each group to only include permissions they have
	callerPerms := ctrl.getCallerPermNames(c)
	if len(callerPerms) == 0 {
		// No permissions at all — return empty
		utils.OK(c, "success", map[string]interface{}{})
		return
	}

	filtered := make(map[string]interface{})
	for group, perms := range grouped {
		var allowed []interface{}
		for _, p := range perms {
			if callerPerms[p.Name] {
				allowed = append(allowed, p)
			}
		}
		if len(allowed) > 0 {
			filtered[group] = allowed
		}
	}
	utils.OK(c, "success", filtered)
}

// getCallerPermNames returns a set of permission names the caller has.
func (ctrl *PermissionController) getCallerPermNames(c *gin.Context) map[string]bool {
	userID := middlewares.GetUserID(c)
	if userID == 0 {
		return nil
	}
	user, err := ctrl.userRepo.FindByID(userID)
	if err != nil || user.Role == nil {
		return nil
	}
	set := make(map[string]bool, len(user.Role.Permissions))
	for _, p := range user.Role.Permissions {
		set[p.Name] = true
	}
	return set
}

// filterToCallerPerms filters a flat permission list to only what the caller has.
func (ctrl *PermissionController) filterToCallerPerms(c *gin.Context, all []models.Permission) []models.Permission {
	if middlewares.IsSuperAdmin(c) {
		return all
	}
	callerPerms := ctrl.getCallerPermNames(c)
	out := make([]models.Permission, 0)
	for _, p := range all {
		if callerPerms[p.Name] {
			out = append(out, p)
		}
	}
	return out
}
