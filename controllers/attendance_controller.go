package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type AttendanceController struct {
	svc services.AttendanceService
}

func NewAttendanceController(svc services.AttendanceService) *AttendanceController {
	return &AttendanceController{svc}
}

type attendanceGeoBody struct {
	BranchID  uint    `json:"branch_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	// Reason is optional on a normal check-in/out, but the service layer
	// requires it when the computed timeliness status is late (check-in)
	// or early (check-out) — see AttendanceService.CheckIn/CheckOut.
	Reason string `json:"reason"`
}

// CheckIn godoc — POST /attendance/check-in
func (ctrl *AttendanceController) CheckIn(c *gin.Context) {
	var body attendanceGeoBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.CheckIn(middlewares.GetUserID(c), body.BranchID, body.Latitude, body.Longitude, body.Reason)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "checked in", item)
}

// CheckOut godoc — POST /attendance/check-out
func (ctrl *AttendanceController) CheckOut(c *gin.Context) {
	var body attendanceGeoBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.CheckOut(middlewares.GetUserID(c), body.BranchID, body.Latitude, body.Longitude, body.Reason)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "checked out", item)
}

// Today godoc — GET /attendance/today?branch_id=1
func (ctrl *AttendanceController) Today(c *gin.Context) {
	branchID, _ := strconv.ParseUint(c.Query("branch_id"), 10, 64)
	item, err := ctrl.svc.Today(middlewares.GetUserID(c), uint(branchID))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", item)
}

// Mine godoc — GET /attendance/mine?date_from=&date_to=&page=&page_size=
// Self-service — always scoped to the caller's own user ID, no special
// permission required (same as Leave/Overtime/Activity's own /mine).
func (ctrl *AttendanceController) Mine(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.AttendanceFilter{
		UserID:   middlewares.GetUserID(c),
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

type attendanceUpdateBody struct {
	CheckInAt  string `json:"check_in_at"`  // "2006-01-02T15:04", "" = leave untouched
	CheckOutAt string `json:"check_out_at"` // "2006-01-02T15:04", "" = leave untouched
}

// Update godoc — PUT /attendance/:id
// Admin correction of an existing record's check-in/check-out timestamps
// — recomputes the timeliness status for whichever side is changed.
func (ctrl *AttendanceController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body attendanceUpdateBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.AdminUpdate(uint(id), body.CheckInAt, body.CheckOutAt)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "attendance record updated", item)
}

// List godoc — GET /attendance?user_id=&branch_id=&date_from=&date_to=&page=&page_size=
func (ctrl *AttendanceController) List(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	filter := repositories.AttendanceFilter{
		UserID:       userID,
		BranchID:     branchID,
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
		ActivityOnly: c.Query("activity_only") == "true",
	}
	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// Summary godoc — GET /attendance/summary?date_from=&date_to=&user_id=&branch_id=
// Per-user, day-by-day ATTEND/ABSENT/LEAVE/DAY_OFF breakdown for the
// given date range, computed server-side.
func (ctrl *AttendanceController) Summary(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	result, err := ctrl.svc.Summary(middlewares.GetUserID(c), c.Query("date_from"), c.Query("date_to"), userID, branchID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "success", result)
}
