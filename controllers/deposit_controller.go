package controllers

import (
	"strconv"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	transactiondto "crm-backend/dto/transaction"
	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type DepositController struct {
	db      *gorm.DB
	svc     services.DepositService
	userSvc services.UserService
}

func NewDepositController(db *gorm.DB, svc services.DepositService, userSvc services.UserService) *DepositController {
	return &DepositController{db: db, svc: svc, userSvc: userSvc}
}

func (ctrl *DepositController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

// List godoc
// @Summary      List deposits with filtering and pagination
// @Tags         Deposits
// @Security     BearerAuth
// @Produce      json
// @Param        search        query string false "Search by transaction no or client name"
// @Param        client_id     query int    false "Filter by client"
// @Param        date_from     query string false "Start date (YYYY-MM-DD)"
// @Param        date_to       query string false "End date (YYYY-MM-DD)"
// @Param        page          query int    false "Page"
// @Param        page_size     query int    false "Page size"
// @Success      200 {object} utils.Response
// @Router       /deposits [get]
func (ctrl *DepositController) List(c *gin.Context) {
	var filter transactiondto.FilterQuery
	if err := utils.BindQuery(c, &filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	p := utils.ParsePagination(c)
	list, total, err := ctrl.svc.List(filter, p, middlewares.GetUserID(c))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OKPaginated(c, list, utils.BuildMeta(p, total))
}

// Create godoc
// @Summary      Create a deposit
// @Tags         Deposits
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body transactiondto.CreateRequest true "Payload"
// @Success      201 {object} utils.Response
// @Router       /deposits [post]
func (ctrl *DepositController) Create(c *gin.Context) {
	var req transactiondto.CreateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	deposit, err := ctrl.svc.Create(middlewares.GetUserID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "deposit created", deposit)
}

// GetByID godoc
// @Summary      Get deposit by ID
// @Tags         Deposits
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Deposit ID"
// @Success      200 {object} utils.Response
// @Router       /deposits/{id} [get]
func (ctrl *DepositController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	deposit, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "deposit")
		return
	}
	utils.OK(c, "success", deposit)
}

// Update godoc
// @Summary      Update a deposit (manual fields only)
// @Tags         Deposits
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Deposit ID"
// @Param        body body transactiondto.UpdateRequest true "Payload"
// @Success      200 {object} utils.Response
// @Router       /deposits/{id} [put]
func (ctrl *DepositController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req transactiondto.UpdateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	deposit, err := ctrl.svc.Update(uint(id), ctrl.scope(c), req, middlewares.GetUserID(c))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "deposit updated", deposit)
}

// Delete godoc
// @Summary      Delete a deposit
// @Tags         Deposits
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Deposit ID"
// @Success      200 {object} utils.Response
// @Router       /deposits/{id} [delete]
func (ctrl *DepositController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c), middlewares.GetUserID(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "deposit deleted", nil)
}

// GetBalance godoc
// @Summary      Get current balance for a client+product combination
// @Tags         Deposits
// @Security     BearerAuth
// @Produce      json
// @Param        client_id         query int true "Client ID"
// @Param        client_product_id query int true "Client Product ID"
// @Success      200 {object} utils.Response
// @Router       /deposits/balance [get]
func (ctrl *DepositController) GetBalance(c *gin.Context) {
	clientID, _ := strconv.ParseUint(c.Query("client_id"), 10, 64)
	productID, _ := strconv.ParseUint(c.Query("client_product_id"), 10, 64)
	if clientID == 0 || productID == 0 {
		utils.BadRequest(c, "client_id and client_product_id are required")
		return
	}
	bal, err := ctrl.svc.GetBalance(uint(clientID), uint(productID))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", bal)
}

// Approve godoc
// @Summary Approve or reject a deposit
func (ctrl *DepositController) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req transactiondto.ApproveRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	deposit, err := ctrl.svc.Approve(uint(id), middlewares.GetUserID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "deposit "+req.Status, deposit)
}
