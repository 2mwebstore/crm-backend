package controllers

import (
	"strconv"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	turnoverbetdto "crm-backend/dto/turnover_bet"
	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type TurnoverBetController struct {
	db      *gorm.DB
	svc     services.TurnoverBetService
	userSvc services.UserService
}

func NewTurnoverBetController(db *gorm.DB, svc services.TurnoverBetService, userSvc services.UserService) *TurnoverBetController {
	return &TurnoverBetController{db: db, svc: svc, userSvc: userSvc}
}

func (ctrl *TurnoverBetController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *TurnoverBetController) List(c *gin.Context) {
	var filter turnoverbetdto.FilterQuery
	c.ShouldBindQuery(&filter)
	p := utils.ParsePagination(c)
	items, total, err := ctrl.svc.List(filter, p, middlewares.GetUserID(c))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OKPaginated(c, items, utils.BuildMeta(p, total))
}

func (ctrl *TurnoverBetController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "turnover bet")
		return
	}
	utils.OK(c, "success", t)
}

func (ctrl *TurnoverBetController) Create(c *gin.Context) {
	var req turnoverbetdto.CreateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	t, err := ctrl.svc.Create(middlewares.GetUserID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "turnover bet created", t)
}

func (ctrl *TurnoverBetController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req turnoverbetdto.UpdateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	t, err := ctrl.svc.Update(uint(id), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "updated", t)
}

func (ctrl *TurnoverBetController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "deleted", nil)
}

func (ctrl *TurnoverBetController) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req turnoverbetdto.ApproveRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	t, err := ctrl.svc.Approve(uint(id), middlewares.GetUserID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "turnover bet "+req.Status, t)
}
