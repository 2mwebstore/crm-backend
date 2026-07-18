package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type OvertimeRequestController struct {
	svc services.OvertimeRequestService
}

func NewOvertimeRequestController(svc services.OvertimeRequestService) *OvertimeRequestController {
	return &OvertimeRequestController{svc}
}

type overtimeRequestCreateBody struct {
	BranchID  *uint   `json:"branch_id"`
	Date      string  `json:"date" binding:"required"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Reason    string  `json:"reason"`
}

// Create godoc — POST /overtime-requests
func (ctrl *OvertimeRequestController) Create(c *gin.Context) {
	var body overtimeRequestCreateBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), services.OvertimeRequestInput{
		BranchID:  body.BranchID,
		Date:      body.Date,
		StartTime: body.StartTime,
		EndTime:   body.EndTime,
		Reason:    body.Reason,
	})
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "overtime request submitted", item)
}

// Mine godoc — GET /overtime-requests/mine?status=&page=&page_size=
func (ctrl *OvertimeRequestController) Mine(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.OvertimeRequestFilter{
		UserID: middlewares.GetUserID(c),
		Status: models.OvertimeRequestStatus(c.Query("status")),
	}
	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// List godoc — GET /overtime-requests?user_id=&branch_id=&status=&date_from=&date_to=&page=&page_size=
func (ctrl *OvertimeRequestController) List(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.OvertimeRequestFilter{
		UserID:   userID,
		BranchID: branchID,
		Status:   models.OvertimeRequestStatus(c.Query("status")),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}
	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

func (ctrl *OvertimeRequestController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "overtime request")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *OvertimeRequestController) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.Approve(uint(id), middlewares.GetUserID(c))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "overtime request approved", item)
}

func (ctrl *OvertimeRequestController) Reject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Reason string `json:"reason"`
	}
	_ = utils.BindJSON(c, &body)
	item, err := ctrl.svc.Reject(uint(id), middlewares.GetUserID(c), body.Reason)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "overtime request rejected", item)
}

// EditReason godoc — PATCH /overtime-requests/:id
// Self-service — the original requester editing their own still-pending
// request's reason. Not an admin action.
func (ctrl *OvertimeRequestController) EditReason(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.EditReason(middlewares.GetUserID(c), uint(id), body.Reason)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "overtime request updated", item)
}

// Cancel godoc — POST /overtime-requests/:id/cancel
// Self-service — the original requester withdrawing their own still-
// pending request. Not an admin action (see Reject for that).
func (ctrl *OvertimeRequestController) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.Cancel(middlewares.GetUserID(c), uint(id))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "overtime request cancelled", item)
}
