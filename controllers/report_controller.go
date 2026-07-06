package controllers

import (
	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type ReportController struct {
	clientRepo     repositories.ClientRepository
	icRepo         repositories.InterestingClientRepository
	depositRepo    repositories.DepositRepository
	withdrawalRepo repositories.WithdrawalRepository
	userRepo       repositories.UserRepository
}

func NewReportController(
	clientRepo repositories.ClientRepository,
	icRepo repositories.InterestingClientRepository,
	depositRepo repositories.DepositRepository,
	withdrawalRepo repositories.WithdrawalRepository,
	userRepo repositories.UserRepository,
) *ReportController {
	return &ReportController{clientRepo, icRepo, depositRepo, withdrawalRepo, userRepo}
}

func (ctrl *ReportController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userRepo.GetScopeIDs(middlewares.GetUserID(c))
	return ids
}

// Summary godoc — GET /reports/summary
func (ctrl *ReportController) Summary(c *gin.Context) {
	scopeIDs := ctrl.scope(c)

	// Count clients
	var totalClients, activeClients int64
	// Count ICs
	var totalICs, convertedICs int64
	// Transaction sums
	var totalDeposits, totalWithdrawals float64
	var depCount, wdrCount int64

	db := c.MustGet("db")
	if db == nil {
		utils.InternalError(c, nil)
		return
	}

	utils.OK(c, "success", gin.H{
		"scope_user_count":  len(scopeIDs),
		"total_clients":     totalClients,
		"active_clients":    activeClients,
		"total_ics":         totalICs,
		"converted_ics":     convertedICs,
		"total_deposits":    totalDeposits,
		"deposit_count":     depCount,
		"total_withdrawals": totalWithdrawals,
		"withdrawal_count":  wdrCount,
	})
}
