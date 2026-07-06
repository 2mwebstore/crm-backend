package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type BranchController struct {
	svc services.BranchService
}

func NewBranchController(svc services.BranchService) *BranchController {
	return &BranchController{svc}
}

func (ctrl *BranchController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	userID := middlewares.GetUserID(c)
	items, err := ctrl.svc.ListForUser(userID, showAll)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", items)
}

func (ctrl *BranchController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "branch")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *BranchController) Create(c *gin.Context) {
	var body struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
	}
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), body.Name, body.Code, body.Description)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "branch created", item)
}

func (ctrl *BranchController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), body.Name, body.Code, body.Description, body.IsActive)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "branch updated", item)
}

func (ctrl *BranchController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "branch deleted", nil)
}
