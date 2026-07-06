package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type BonusOptionTypeController struct {
	svc     services.BonusOptionTypeService
	userSvc services.UserService
}

func NewBonusOptionTypeController(svc services.BonusOptionTypeService, userSvc services.UserService) *BonusOptionTypeController {
	return &BonusOptionTypeController{svc, userSvc}
}

// scope resolves the caller's lookup scope for single-record operations via
// GetLookupScope: nil for Super Admins and users with branches assigned
// (full access to shared lookups), and the sentinel []uint{0} for users
// with no branches (no access).
func (ctrl *BonusOptionTypeController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetLookupScope(middlewares.GetUserID(c))
	return ids
}

// List follows the same flow as BranchController.List / BankTypeController.List:
// it scopes by the caller's actual branch assignment (Super Admin sees all,
// otherwise only bonus options whose branch_id matches one of the caller's
// branches).
func (ctrl *BonusOptionTypeController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	var branchID *uint
	if bidStr := c.Query("branch_id"); bidStr != "" {
		if bid, err := strconv.ParseUint(bidStr, 10, 64); err == nil {
			b := uint(bid)
			branchID = &b
		}
	}
	items, err := ctrl.svc.ListForUser(middlewares.GetUserID(c), showAll, branchID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", items)
}

func (ctrl *BonusOptionTypeController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "bonus option")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *BonusOptionTypeController) Create(c *gin.Context) {
	var x models.BonusOptionType
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), ctrl.scope(c), &x)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "bonus option created", item)
}

func (ctrl *BonusOptionTypeController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var x models.BonusOptionType
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), &x)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "bonus option updated", item)
}

func (ctrl *BonusOptionTypeController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "bonus option deleted", nil)
}
