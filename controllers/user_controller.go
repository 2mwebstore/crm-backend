package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	userdto "crm-backend/dto/user"
	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type UserController struct {
	svc services.UserService
}

func NewUserController(svc services.UserService) *UserController {
	return &UserController{svc}
}

func callerID(c *gin.Context) uint { return middlewares.GetUserID(c) }
func targetID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}

// ── Super Admin endpoints ─────────────────────────────────────────────────────

// AdminListUsers godoc
// @Summary List all users in the system (super admin only)
func (ctrl *UserController) AdminListUsers(c *gin.Context) {
	users, err := ctrl.svc.ListAllUsers()
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", users)
}

// AdminCreateUser godoc
// @Summary Create any user (super admin only) — can set parent_id to make sub-user
func (ctrl *UserController) AdminCreateUser(c *gin.Context) {
	var req userdto.AdminCreateUserRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	user, err := ctrl.svc.AdminCreateUser(req)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "user created", user)
}

// AdminUpdateUser godoc
// @Summary Update any user (super admin only)
func (ctrl *UserController) AdminUpdateUser(c *gin.Context) {
	var req userdto.AdminUpdateUserRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	// Treat empty password string as "no change"
	if req.Password != nil && *req.Password == "" {
		req.Password = nil
	}
	user, err := ctrl.svc.AdminUpdateUser(targetID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "user updated", user)
}

// AdminDeleteUser godoc
// @Summary Delete any user (super admin only)
func (ctrl *UserController) AdminDeleteUser(c *gin.Context) {
	if err := ctrl.svc.AdminDeleteUser(targetID(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "user deleted", nil)
}

// ── Simple / Sub-user endpoints ───────────────────────────────────────────────

// ListSubUsers godoc
// @Summary List all descendants of the caller
func (ctrl *UserController) ListSubUsers(c *gin.Context) {
	users, err := ctrl.svc.ListSubUsers(callerID(c))
	if err != nil {
		utils.Forbidden(c)
		return
	}
	utils.OK(c, "success", users)
}

// GetSubUser godoc
// @Summary Get a sub-user (must be a descendant of caller)
func (ctrl *UserController) GetSubUser(c *gin.Context) {
	user, err := ctrl.svc.GetSubUser(callerID(c), targetID(c))
	if err != nil {
		utils.NotFound(c, "user")
		return
	}
	utils.OK(c, "success", user)
}

// CreateSubUser godoc
// @Summary Create a sub-user under the caller's account
func (ctrl *UserController) CreateSubUser(c *gin.Context) {
	var req userdto.CreateSubUserRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	user, err := ctrl.svc.CreateSubUser(callerID(c), req)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "sub-user created", user)
}

// UpdateSubUser godoc
// @Summary Update a sub-user
func (ctrl *UserController) UpdateSubUser(c *gin.Context) {
	var req userdto.UpdateSubUserRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	// Treat empty password string as "no change"
	if req.Password != nil && *req.Password == "" {
		req.Password = nil
	}
	user, err := ctrl.svc.UpdateSubUser(callerID(c), targetID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "sub-user updated", user)
}

// DeleteSubUser godoc
// @Summary Delete a sub-user
func (ctrl *UserController) DeleteSubUser(c *gin.Context) {
	if err := ctrl.svc.DeleteSubUser(callerID(c), targetID(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "sub-user deleted", nil)
}

// AssignBranches godoc — PUT /users/admin/:id/branches
// Super Admin only: assigns multiple branches to a user.
func (ctrl *UserController) AssignBranches(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		BranchIDs []uint `json:"branch_ids"`
	}
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.svc.GetUserRepo().AssignBranches(uint(id), body.BranchIDs); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	user, _ := ctrl.svc.GetUserRepo().FindByID(uint(id))
	utils.OK(c, "branches assigned", user)
}

// ListUsersInScope returns users visible to the caller for filter dropdowns.
func (ctrl *UserController) ListUsersInScope(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	users, err := ctrl.svc.GetUsersInScope(userID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", users)
}
