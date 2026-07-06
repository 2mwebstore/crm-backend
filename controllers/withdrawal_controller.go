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

type WithdrawalController struct {
	db      *gorm.DB
	svc     services.WithdrawalService
	userSvc services.UserService
}

func NewWithdrawalController(db *gorm.DB, svc services.WithdrawalService, userSvc services.UserService) *WithdrawalController {
	return &WithdrawalController{db: db, svc: svc, userSvc: userSvc}
}

func (ctrl *WithdrawalController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

// List godoc
// @Summary      List withdrawals with filtering and pagination
// @Tags         Withdrawals
// @Security     BearerAuth
// @Produce      json
// @Param        search    query string false "Search by transaction no or client name"
// @Param        client_id query int    false "Filter by client"
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to   query string false "End date (YYYY-MM-DD)"
// @Param        page      query int    false "Page"
// @Param        page_size query int    false "Page size"
// @Success      200 {object} utils.Response
// @Router       /withdrawals [get]
func (ctrl *WithdrawalController) List(c *gin.Context) {
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
// @Summary      Create a withdrawal
// @Tags         Withdrawals
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body transactiondto.CreateRequest true "Payload"
// @Success      201 {object} utils.Response
// @Router       /withdrawals [post]
func (ctrl *WithdrawalController) Create(c *gin.Context) {
	var req transactiondto.CreateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	withdrawal, err := ctrl.svc.Create(middlewares.GetUserID(c), req)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Created(c, "withdrawal created", withdrawal)
}

// GetByID godoc
// @Summary      Get withdrawal by ID
// @Tags         Withdrawals
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Withdrawal ID"
// @Success      200 {object} utils.Response
// @Router       /withdrawals/{id} [get]
func (ctrl *WithdrawalController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	withdrawal, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "withdrawal")
		return
	}
	utils.OK(c, "success", withdrawal)
}

// Update godoc
// @Summary      Update a withdrawal (manual fields only)
// @Tags         Withdrawals
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Withdrawal ID"
// @Param        body body transactiondto.UpdateRequest true "Payload"
// @Success      200 {object} utils.Response
// @Router       /withdrawals/{id} [put]
func (ctrl *WithdrawalController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req transactiondto.UpdateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	withdrawal, err := ctrl.svc.Update(uint(id), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "withdrawal updated", withdrawal)
}

// Delete godoc
// @Summary      Delete a withdrawal
// @Tags         Withdrawals
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Withdrawal ID"
// @Success      200 {object} utils.Response
// @Router       /withdrawals/{id} [delete]
func (ctrl *WithdrawalController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "withdrawal deleted", nil)
}

// GetBalance godoc
// @Summary      Get current balance for a client+product combination
// @Tags         Withdrawals
// @Security     BearerAuth
// @Produce      json
// @Param        client_id         query int true "Client ID"
// @Param        client_product_id query int true "Client Product ID"
// @Success      200 {object} utils.Response
// @Router       /withdrawals/balance [get]
func (ctrl *WithdrawalController) GetBalance(c *gin.Context) {
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
// @Summary Approve or reject a withdrawal
func (ctrl *WithdrawalController) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req transactiondto.ApproveRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	withdrawal, err := ctrl.svc.Approve(uint(id), middlewares.GetUserID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "withdrawal "+req.Status, withdrawal)
}
