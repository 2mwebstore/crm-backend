package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type DailyStartBalanceController struct {
	svc services.DailyStartBalanceService
}

func NewDailyStartBalanceController(svc services.DailyStartBalanceService) *DailyStartBalanceController {
	return &DailyStartBalanceController{svc}
}

type dailyBalanceBranchBody struct {
	BranchID uint `json:"branch_id" binding:"required"`
}

func branchIDFromQuery(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Query("branch_id"), 10, 64)
	return uint(id), err == nil && id > 0
}

// StartToday godoc — POST /daily-balances/start
// Body: {"branch_id": 1}
func (ctrl *DailyStartBalanceController) StartToday(c *gin.Context) {
	var body dailyBalanceBranchBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	snap, err := ctrl.svc.StartToday(middlewares.GetUserID(c), body.BranchID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "today's balance snapshot started", snap)
}

// CloseToday godoc — POST /daily-balances/close
// Body: {"branch_id": 1}
// Captures the end-of-day Close Cash / Close Credit totals for staff
// finishing their shift — fails if today hasn't been started yet, or was
// already closed.
func (ctrl *DailyStartBalanceController) CloseToday(c *gin.Context) {
	var body dailyBalanceBranchBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	snap, err := ctrl.svc.CloseToday(middlewares.GetUserID(c), body.BranchID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "today's balance closed", snap)
}

// Today godoc — GET /daily-balances/today?branch_id=1
func (ctrl *DailyStartBalanceController) Today(c *gin.Context) {
	branchID, ok := branchIDFromQuery(c)
	if !ok {
		utils.BadRequest(c, "branch_id is required")
		return
	}
	resp, err := ctrl.svc.GetToday(middlewares.GetUserID(c), branchID)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", resp)
}

// History godoc — GET /daily-balances/history?branch_id=1&page=1&page_size=20
func (ctrl *DailyStartBalanceController) History(c *gin.Context) {
	branchID, ok := branchIDFromQuery(c)
	if !ok {
		utils.BadRequest(c, "branch_id is required")
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	list, total, err := ctrl.svc.History(middlewares.GetUserID(c), branchID, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// ShiftBalanceTransactions godoc — GET /daily-balances/:id/balance-transactions
// On-demand lookup of the ledger entries (top-ups/withdrawals) recorded
// during a specific shift — open or already closed. Used from the History
// table's "View Transactions" action, rather than eagerly including this
// on every row of the paginated History list.
func (ctrl *DailyStartBalanceController) ShiftBalanceTransactions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid shift id")
		return
	}
	list, err := ctrl.svc.GetShiftBalanceTransactions(middlewares.GetUserID(c), uint(id))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "success", list)
}
