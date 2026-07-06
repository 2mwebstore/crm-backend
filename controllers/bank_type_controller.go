package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type BankTypeController struct {
	svc     services.BankTypeService
	userSvc services.UserService
}

func NewBankTypeController(svc services.BankTypeService, userSvc services.UserService) *BankTypeController {
	return &BankTypeController{svc, userSvc}
}

func (ctrl *BankTypeController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *BankTypeController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	items, err := ctrl.svc.ListForUser(middlewares.GetUserID(c), showAll)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", items)
}

func (ctrl *BankTypeController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "bank type")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *BankTypeController) Create(c *gin.Context) {
	var x models.BankType
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), &x)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "bank type created", item)
}

func (ctrl *BankTypeController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var x models.BankType
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), &x)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "bank type updated", item)
}

func (ctrl *BankTypeController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "bank type deleted", nil)
}
