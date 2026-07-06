package controllers

import (
	"strconv"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	followupdto "crm-backend/dto/follow_up"
	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type FollowUpController struct {
	db      *gorm.DB
	svc     services.FollowUpService
	userSvc services.UserService
}

func NewFollowUpController(db *gorm.DB, svc services.FollowUpService, userSvc services.UserService) *FollowUpController {
	return &FollowUpController{db: db, svc: svc, userSvc: userSvc}
}

func (ctrl *FollowUpController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *FollowUpController) List(c *gin.Context) {
	var filter followupdto.FilterQuery
	if err := utils.BindQuery(c, &filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	p := utils.ParsePagination(c)
	items, total, err := ctrl.svc.List(filter, p, middlewares.GetUserID(c))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OKPaginated(c, items, utils.BuildMeta(p, total))
}

func (ctrl *FollowUpController) Create(c *gin.Context) {
	var req followupdto.CreateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	f, err := ctrl.svc.Create(middlewares.GetUserID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "follow-up created", f)
}

func (ctrl *FollowUpController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	f, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "follow-up")
		return
	}
	utils.OK(c, "success", f)
}

func (ctrl *FollowUpController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "deleted", nil)
}
