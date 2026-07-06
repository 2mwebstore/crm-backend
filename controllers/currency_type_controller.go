package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type CurrencyTypeController struct {
	svc     services.CurrencyTypeService
	userSvc services.UserService
}

func NewCurrencyTypeController(svc services.CurrencyTypeService, userSvc services.UserService) *CurrencyTypeController {
	return &CurrencyTypeController{svc, userSvc}
}

func (ctrl *CurrencyTypeController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *CurrencyTypeController) List(c *gin.Context) {
	showAll := c.Query("show_all") == "true"
	items, err := ctrl.svc.List(ctrl.scope(c), showAll)
	if err != nil { utils.InternalError(c, err); return }
	utils.OK(c, "success", items)
}

func (ctrl *CurrencyTypeController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil { utils.NotFound(c, "currency"); return }
	utils.OK(c, "success", item)
}

func (ctrl *CurrencyTypeController) Create(c *gin.Context) {
	var x models.CurrencyType
	if err := utils.BindJSON(c, &x); err != nil { utils.BadRequest(c, err.Error()); return }
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), ctrl.scope(c), &x)
	if err != nil { utils.Conflict(c, err.Error()); return }
	utils.Created(c, "currency created", item)
}

func (ctrl *CurrencyTypeController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var x models.CurrencyType
	if err := utils.BindJSON(c, &x); err != nil { utils.BadRequest(c, err.Error()); return }
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), &x)
	if err != nil { utils.BadRequest(c, err.Error()); return }
	utils.OK(c, "currency updated", item)
}

func (ctrl *CurrencyTypeController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error()); return
	}
	utils.OK(c, "currency deleted", nil)
}
