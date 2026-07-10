package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type BalanceTransactionController struct {
	svc services.BalanceTransactionService
}

func NewBalanceTransactionController(svc services.BalanceTransactionService) *BalanceTransactionController {
	return &BalanceTransactionController{svc}
}

// List godoc — GET /balance-transactions
//
//	?entity_type=company_bank&entity_id=1&page=1&page_size=20
//	&source=configuration&type=topup
//	&date_from=2026-07-01&date_to=2026-07-10&created_by_id=3
//
// entity_type must be one of the models.BalanceEntityType constants (e.g.
// "company_bank", "product_type"). Every other param is optional.
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

	var createdByID uint
	if v, err := strconv.ParseUint(c.Query("created_by_id"), 10, 64); err == nil {
		createdByID = uint(v)
	}

	filter := repositories.BalanceTransactionFilter{
		Type:        models.BalanceTxType(c.Query("type")),
		Source:      models.BalanceTxSource(c.Query("source")),
		DateFrom:    c.Query("date_from"),
		DateTo:      c.Query("date_to"),
		CreatedByID: createdByID,
	}

	list, total, err := ctrl.svc.ListByEntity(entityType, uint(entityID), filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}
