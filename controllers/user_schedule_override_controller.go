package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type UserScheduleOverrideController struct {
	svc services.UserScheduleOverrideService
}

func NewUserScheduleOverrideController(svc services.UserScheduleOverrideService) *UserScheduleOverrideController {
	return &UserScheduleOverrideController{svc}
}

type scheduleOverrideBody struct {
	UserID            uint    `json:"user_id" binding:"required"`
	DateFrom          string  `json:"date_from" binding:"required"`
	DateTo            string  `json:"date_to" binding:"required"`
	ShiftCheckInTime  *string `json:"shift_check_in_time"`
	ShiftCheckOutTime *string `json:"shift_check_out_time"`
	Reason            string  `json:"reason"`
}

func (b scheduleOverrideBody) toInput() services.UserScheduleOverrideInput {
	return services.UserScheduleOverrideInput{
		UserID:            b.UserID,
		DateFrom:          b.DateFrom,
		DateTo:            b.DateTo,
		ShiftCheckInTime:  b.ShiftCheckInTime,
		ShiftCheckOutTime: b.ShiftCheckOutTime,
		Reason:            b.Reason,
	}
}

// Create godoc — POST /schedule-overrides
func (ctrl *UserScheduleOverrideController) Create(c *gin.Context) {
	var body scheduleOverrideBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Create(middlewares.GetUserID(c), body.toInput())
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "schedule override created", item)
}

// Update godoc — PUT /schedule-overrides/:id
func (ctrl *UserScheduleOverrideController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body scheduleOverrideBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	item, err := ctrl.svc.Update(uint(id), body.toInput())
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "schedule override updated", item)
}

// List godoc — GET /schedule-overrides?user_id=&branch_id=&search=&page=&page_size=
// Every filter is optional — with none set, returns overrides across ALL
// users.
func (ctrl *UserScheduleOverrideController) List(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	filter := repositories.UserScheduleOverrideFilter{
		UserID:   userID,
		BranchID: branchID,
		Search:   c.Query("search"),
	}
	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// ListForUser godoc — GET /schedule-overrides/for-user?user_id=5
// Unpaginated convenience endpoint — full history for exactly one user.
func (ctrl *UserScheduleOverrideController) ListForUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil || userID == 0 {
		utils.BadRequest(c, "user_id is required")
		return
	}
	list, err := ctrl.svc.ListForUser(uint(userID))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", list)
}

// Delete godoc — DELETE /schedule-overrides/:id
func (ctrl *UserScheduleOverrideController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := ctrl.svc.Delete(uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "schedule override deleted", nil)
}
