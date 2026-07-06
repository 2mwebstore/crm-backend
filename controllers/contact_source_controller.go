package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type ContactSourceController struct {
	svc     services.ContactSourceService
	userSvc services.UserService
}

func NewContactSourceController(svc services.ContactSourceService, userSvc services.UserService) *ContactSourceController {
	return &ContactSourceController{svc, userSvc}
}

// scope resolves the caller's lookup scope for single-record operations via
// GetLookupScope: nil for Super Admins and users with branches assigned
// (full access to shared lookups), and the sentinel []uint{0} for users
// with no branches (no access).
func (ctrl *ContactSourceController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetLookupScope(middlewares.GetUserID(c))
	return ids
}

// List follows the same flow as BranchController.List / BankTypeController.List:
// it scopes by the caller's actual branch assignment (Super Admin sees all,
// otherwise only contact sources whose branch_id matches one of the
// caller's branches).
func (ctrl *ContactSourceController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	userID := middlewares.GetUserID(c)
	items, err := ctrl.svc.ListForUser(userID, showAll)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", items)
}

func (ctrl *ContactSourceController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "contact source")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *ContactSourceController) Create(c *gin.Context) {
	var x models.ContactSource
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), ctrl.scope(c), &x)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "contact source created", item)
}

func (ctrl *ContactSourceController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var x models.ContactSource
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), &x)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "contact source updated", item)
}

func (ctrl *ContactSourceController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "contact source deleted", nil)
}
