package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type CompanyBankController struct {
	svc     services.CompanyBankService
	userSvc services.UserService
}

func NewCompanyBankController(svc services.CompanyBankService, userSvc services.UserService) *CompanyBankController {
	return &CompanyBankController{svc, userSvc}
}

func (ctrl *CompanyBankController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}

func (ctrl *CompanyBankController) List(c *gin.Context) {
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

func (ctrl *CompanyBankController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	item, err := ctrl.svc.GetByID(uint(id), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "company bank")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *CompanyBankController) Create(c *gin.Context) {
	var x models.CompanyBank
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), &x)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "company bank created", item)
}

func (ctrl *CompanyBankController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var x models.CompanyBank
	if err := utils.BindJSON(c, &x); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), ctrl.scope(c), &x)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "company bank updated", item)
}

func (ctrl *CompanyBankController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	if err := ctrl.svc.Delete(uint(id), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "company bank deleted", nil)
}

// TopUpCash godoc — POST /company-banks/:id/topup-cash  { "amount": 100.00, "remark": "..." }
// amount must be positive. Applies atomically at the DB level and writes a
// BalanceTransaction ledger row in the same transaction.
func (ctrl *CompanyBankController) TopUpCash(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	// NOTE: no `binding:"required"` on Amount — Gin's required validator
	// treats a zero float64 as "empty", so the validation is done manually
	// in the service instead (and here amount must be > 0, not just != 0,
	// since direction is now expressed by which endpoint is called).
	var req struct {
		Amount float64 `json:"amount"`
		Remark string  `json:"remark"`
	}
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.TopUpCash(uint(id), ctrl.scope(c), req.Amount, req.Remark, middlewares.GetUserID(c))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "cash topped up", item)
}

// WithdrawCash godoc — POST /company-banks/:id/withdraw-cash  { "amount": 50.00, "remark": "..." }
// amount must be positive; the service rejects it if it would take cash negative.
func (ctrl *CompanyBankController) WithdrawCash(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid id")
		return
	}
	var req struct {
		Amount float64 `json:"amount"`
		Remark string  `json:"remark"`
	}
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.WithdrawCash(uint(id), ctrl.scope(c), req.Amount, req.Remark, middlewares.GetUserID(c))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "cash withdrawn", item)
}
