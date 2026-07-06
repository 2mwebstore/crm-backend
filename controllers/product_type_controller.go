package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type ProductTypeController struct {
	svc     services.ProductTypeService
	userSvc services.UserService
}

func NewProductTypeController(svc services.ProductTypeService, userSvc services.UserService) *ProductTypeController {
	return &ProductTypeController{svc, userSvc}
}

func (ctrl *ProductTypeController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *ProductTypeController) List(c *gin.Context) {
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
func (ctrl *ProductTypeController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "product type")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *ProductTypeController) Create(c *gin.Context) {
	var x models.ProductType
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), &x)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "product type created", item)
}

func (ctrl *ProductTypeController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var x models.ProductType
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), &x)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "product type updated", item)
}

func (ctrl *ProductTypeController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "product type deleted", nil)
}
