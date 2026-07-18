package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type ActivityRequestController struct {
	svc           services.ActivityRequestService
	attendanceSvc services.AttendanceService
}

func NewActivityRequestController(svc services.ActivityRequestService, attendanceSvc services.AttendanceService) *ActivityRequestController {
	return &ActivityRequestController{svc, attendanceSvc}
}

type activityRequestCreateBody struct {
	BranchID uint   `json:"branch_id" binding:"required"`
	Date     string `json:"date" binding:"required"`
	Reason   string `json:"reason"`
	// Latitude/Longitude are only used when Date is today — see the auto
	// check-in/out logic in Create below.
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// todayDateString mirrors services.todayDateString() (unexported there,
// so re-implemented here) — Asia/Phnom_Penh, "2006-01-02".
func todayDateString() string {
	loc, err := time.LoadLocation("Asia/Phnom_Penh")
	if err != nil {
		loc = time.FixedZone("+07", 7*60*60)
	}
	return time.Now().In(loc).Format("2006-01-02")
}

// Create godoc — POST /activity-requests
//
// For a request covering TODAY with a latitude/longitude included, this
// also drives attendance automatically — self-service Activity is meant
// to replace the separate manual Check In/Check Out step entirely:
//   - not checked in yet today → checks the user in (via the activity
//     exemption, which this newly-created request itself now satisfies)
//   - already checked in, not checked out → checks the user out
//   - already checked out today → the request itself is already blocked
//     by the service layer before reaching here
//
// Any failure in that follow-up attendance step is intentionally
// swallowed — the request itself already succeeded, and that's the
// primary outcome; a person can still check in/out manually from My
// Attendance if the automatic step didn't apply for some reason.
func (ctrl *ActivityRequestController) Create(c *gin.Context) {
	var body activityRequestCreateBody
	if err := utils.BindJSON(c, &body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	userID := middlewares.GetUserID(c)
	item, err := ctrl.svc.Create(userID, services.ActivityRequestInput{
		BranchID: body.BranchID,
		Date:     body.Date,
		Reason:   body.Reason,
	})
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if body.Date == todayDateString() && body.Latitude != nil && body.Longitude != nil {
		att, _ := ctrl.attendanceSvc.Today(userID, body.BranchID)
		if att == nil || att.CheckInAt == nil {
			_, _ = ctrl.attendanceSvc.CheckIn(userID, body.BranchID, *body.Latitude, *body.Longitude, body.Reason)
		} else if att.CheckOutAt == nil {
			_, _ = ctrl.attendanceSvc.CheckOut(userID, body.BranchID, *body.Latitude, *body.Longitude, body.Reason)
		}
	}

	utils.Created(c, "activity request submitted", item)
}

// Mine godoc — GET /activity-requests/mine?page=&page_size=
func (ctrl *ActivityRequestController) Mine(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.ActivityRequestFilter{UserID: middlewares.GetUserID(c)}
	list, total, err := ctrl.svc.List(filter, page, pageSize)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.OK(c, "success", gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

// List godoc — GET /activity-requests?user_id=&branch_id=&date_from=&date_to=&page=&page_size=
func (ctrl *ActivityRequestController) List(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	filter := repositories.ActivityRequestFilter{
		UserID:   userID,
		BranchID: branchID,
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

func (ctrl *ActivityRequestController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := ctrl.svc.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "activity request")
		return
	}
	utils.OK(c, "success", item)
}
