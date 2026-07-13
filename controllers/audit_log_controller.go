package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type AuditLogController struct {
	svc services.AuditLogService
}

func NewAuditLogController(svc services.AuditLogService) *AuditLogController {
	return &AuditLogController{svc}
}

// List godoc — GET /audit-logs
//
//	?user_id=3&branch_id=1&method=DELETE&search=/deposits
//	&date_from=2026-07-01&date_to=2026-07-13&page=1&page_size=20
//
// Every filter is optional.
func (ctrl *AuditLogController) List(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	filter := repositories.AuditLogFilter{
		UserID:   userID,
		BranchID: branchID,
		Method:   c.Query("method"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Search:   c.Query("search"),
	}

	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// DeleteOld godoc — DELETE /audit-logs?period=week|month
// Permanently deletes every entry older than the given period. This is a
// bulk, irreversible operation — gated by its own audit_logs.delete
// permission, separate from just viewing the log.
func (ctrl *AuditLogController) DeleteOld(c *gin.Context) {
	period := c.Query("period")
	count, err := ctrl.svc.DeleteOlderThan(period)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "audit log entries deleted", gin.H{"deleted": count})
}
