package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	lookupdto "crm-backend/dto/lookup"
	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type RoleController struct {
	svc     services.RoleService
	userSvc services.UserService
}

func NewRoleController(svc services.RoleService, userSvc services.UserService) *RoleController {
	return &RoleController{svc, userSvc}
}

// List godoc
// @Summary      List roles accessible to the caller (system + own-subtree)
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /roles [get]
func (ctrl *RoleController) List(c *gin.Context) {
	callerID := middlewares.GetUserID(c)

	nameFilter := c.Query("name")

	var createdByID *uint
	if v := c.Query("created_by"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			utils.BadRequest(c, "invalid created_by")
			return
		}
		u := uint(id)
		createdByID = &u
	}

	roles, err := ctrl.svc.ListAccessible(callerID, nameFilter, createdByID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", roles)
}

// GetByID godoc
// @Summary      Get role details
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Role ID"
// @Success      200 {object} utils.Response
// @Router       /roles/{id} [get]
func (ctrl *RoleController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	callerID := middlewares.GetUserID(c)
	subtree, _ := ctrl.userSvc.GetScopeIDs(callerID)

	role, err := ctrl.svc.GetByID(uint(id), subtree)
	if err != nil {
		utils.NotFound(c, "role")
		return
	}
	utils.OK(c, "success", role)
}

// Create godoc
// @Summary      Create a custom role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body lookupdto.CreateRoleRequest true "Payload"
// @Success      201 {object} utils.Response
// @Router       /roles [post]
func (ctrl *RoleController) Create(c *gin.Context) {
	var req lookupdto.CreateRoleRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	callerID := middlewares.GetUserID(c)
	role, err := ctrl.svc.Create(callerID, req)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "role created", role)
}

// Update godoc
// @Summary      Update a custom role name/description
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Role ID"
// @Param        body body lookupdto.UpdateRoleRequest true "Payload"
// @Success      200 {object} utils.Response
// @Router       /roles/{id} [put]
func (ctrl *RoleController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req lookupdto.UpdateRoleRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	callerID := middlewares.GetUserID(c)
	subtree, _ := ctrl.userSvc.GetScopeIDs(callerID)
	role, err := ctrl.svc.Update(uint(id), callerID, subtree, req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "role updated", role)
}

// Delete godoc
// @Summary      Delete a custom role
// @Tags         Roles
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Role ID"
// @Success      200 {object} utils.Response
// @Router       /roles/{id} [delete]
func (ctrl *RoleController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	callerID := middlewares.GetUserID(c)
	subtree, _ := ctrl.userSvc.GetScopeIDs(callerID)
	if err := ctrl.svc.Delete(uint(id), subtree); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "role deleted", nil)
}

// AssignPermissions godoc
// @Summary      Replace all permissions on a role
// @Tags         Roles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Role ID"
// @Param        body body lookupdto.AssignPermissionsRequest true "Permission IDs"
// @Success      200 {object} utils.Response
// @Router       /roles/{id}/permissions [put]
func (ctrl *RoleController) AssignPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req lookupdto.AssignPermissionsRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	callerID := middlewares.GetUserID(c)
	subtree, _ := ctrl.userSvc.GetScopeIDs(callerID)
	role, err := ctrl.svc.AssignPermissions(uint(id), req.PermissionIDs, subtree)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "permissions updated", role)
}
