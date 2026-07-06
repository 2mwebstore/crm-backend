package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type LevelController struct {
	svc     services.LevelService
	userSvc services.UserService
}

func NewLevelController(svc services.LevelService, userSvc services.UserService) *LevelController {
	return &LevelController{svc, userSvc}
}

func (ctrl *LevelController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *LevelController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	items, err := ctrl.svc.ListForUser(middlewares.GetUserID(c), showAll)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", items)
}

func (ctrl *LevelController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "level")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *LevelController) Create(c *gin.Context) {
	var body struct {
		BranchID    uint   `json:"branch_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), body.BranchID, body.Name, body.Description, body.Color, body.SortOrder)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "level created", item)
}

func (ctrl *LevelController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var body struct {
		BranchID    uint   `json:"branch_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		SortOrder   int    `json:"sort_order"`
		IsActive    bool   `json:"is_active"`
	}
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), body.BranchID, body.Name, body.Description, body.Color, body.SortOrder, body.IsActive)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "level updated", item)
}

func (ctrl *LevelController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "level deleted", nil)
}
