package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/models"
	"crm-backend/services"
	"crm-backend/utils"
)

type BalanceTransactionController struct {
	svc services.BalanceTransactionService
}

func NewBalanceTransactionController(svc services.BalanceTransactionService) *BalanceTransactionController {
	return &BalanceTransactionController{svc}
}

// List godoc — GET /balance-transactions?entity_type=company_bank&entity_id=1&page=1&page_size=20
// entity_type must be one of the models.BalanceEntityType constants (e.g.
// "company_bank", "product_type").
func (ctrl *BalanceTransactionController) List(c *gin.Context) {
	entityType := models.BalanceEntityType(c.Query("entity_type"))
	if entityType == "" {
		utils.BadRequest(c, "entity_type is required")
		return
	}
	entityID, err := strconv.ParseUint(c.Query("entity_id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "entity_id is required")
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	txType := models.BalanceTxType(c.Query("type")) // "" = no filter

	list, total, err := ctrl.svc.ListByEntity(entityType, uint(entityID), txType, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}
