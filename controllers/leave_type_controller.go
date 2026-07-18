package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type LeaveTypeController struct {
	svc services.LeaveTypeService
}

func NewLeaveTypeController(svc services.LeaveTypeService) *LeaveTypeController {
	return &LeaveTypeController{svc}
}

func (ctrl *LeaveTypeController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	var branchID *uint
	if bidStr := c.Query("branch_id"); bidStr != "" {
		if bid, err := strconv.ParseUint(bidStr, 10, 64); err == nil {
			b := uint(bid)
			branchID = &b
		}
	}
	items, err := ctrl.svc.List(middlewares.GetUserID(c), showAll, branchID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", items)
}

func (ctrl *LeaveTypeController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.GetByID(uint(id), middlewares.GetUserID(c))
	if err != nil {
		utils.NotFound(c, "leave type")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *LeaveTypeController) Create(c *gin.Context) {
	var body models.LeaveType
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), &body)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "leave type created", item)
}

func (ctrl *LeaveTypeController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body models.LeaveType
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), middlewares.GetUserID(c), &body)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "leave type updated", item)
}

func (ctrl *LeaveTypeController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "leave type deleted", nil)
}
