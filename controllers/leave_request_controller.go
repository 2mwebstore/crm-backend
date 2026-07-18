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

type LeaveRequestController struct {
	svc services.LeaveRequestService
}

func NewLeaveRequestController(svc services.LeaveRequestService) *LeaveRequestController {
	return &LeaveRequestController{svc}
}

type leaveRequestCreateBody struct {
	BranchID    *uint                      `json:"branch_id"`
	LeaveTypeID uint                       `json:"leave_type_id" binding:"required"`
	DayType     models.LeaveRequestDayType `json:"day_type"`
	DateFrom    string                     `json:"date_from" binding:"required"`
	DateTo      string                     `json:"date_to" binding:"required"`
	Reason      string                     `json:"reason"`
}

// Create godoc — POST /leave-requests
func (ctrl *LeaveRequestController) Create(c *gin.Context) {
	var body leaveRequestCreateBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), services.LeaveRequestInput{
		BranchID:    body.BranchID,
		LeaveTypeID: body.LeaveTypeID,
		DayType:     body.DayType,
		DateFrom:    body.DateFrom,
		DateTo:      body.DateTo,
		Reason:      body.Reason,
	})
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "leave request submitted", item)
}

// Mine godoc — GET /leave-requests/mine?status=&page=&page_size=
// Self-service — always scoped to the caller's own user ID.
func (ctrl *LeaveRequestController) Mine(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.LeaveRequestFilter{
		UserID: middlewares.GetUserID(c),
		Status: models.LeaveRequestStatus(c.Query("status")),
	}
	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// List godoc — GET /leave-requests?user_id=&branch_id=&status=&date_from=&date_to=&page=&page_size=
func (ctrl *LeaveRequestController) List(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.LeaveRequestFilter{
		UserID:   userID,
		BranchID: branchID,
		Status:   models.LeaveRequestStatus(c.Query("status")),
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

func (ctrl *LeaveRequestController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "leave request")
		return
	}
	utils.OK(c, "success", item)
}

func (ctrl *LeaveRequestController) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.Approve(uint(id), middlewares.GetUserID(c))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "leave request approved", item)
}

func (ctrl *LeaveRequestController) Reject(c *gin.Context) {
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
	utils.OK(c, "leave request rejected", item)
}

// EditReason godoc — PATCH /leave-requests/:id
// Self-service — the original requester editing their own still-pending
// request's reason. Not an admin action.
func (ctrl *LeaveRequestController) EditReason(c *gin.Context) {
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
	utils.OK(c, "leave request updated", item)
}

// Cancel godoc — POST /leave-requests/:id/cancel
// Self-service — the original requester withdrawing their own still-
// pending request. Not an admin action (see Reject for that).
func (ctrl *LeaveRequestController) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.Cancel(middlewares.GetUserID(c), uint(id))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "leave request cancelled", item)
}
