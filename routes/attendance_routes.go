package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterAttendanceRoutes(rg *gin.RouterGroup, ctrl *controllers.AttendanceController, userRepo repositories.UserRepository) {
	g := rg.Group("/attendance")
	g.Use(middlewares.Auth())
	{
		// Checking yourself in/out, and seeing your own "today" status,
		// require attendance.request — a sibling to View/Edit below, not
		// a prerequisite for them. A user with none of the three
		// (View, Edit, Request) sees/can do nothing under Attendance.
		g.POST("/check-in", middlewares.RequirePermission(userRepo, models.PermAttendanceRequest), ctrl.CheckIn)
		g.POST("/check-out", middlewares.RequirePermission(userRepo, models.PermAttendanceRequest), ctrl.CheckOut)
		g.GET("/today", middlewares.RequirePermission(userRepo, models.PermAttendanceRequest), ctrl.Today)
		g.GET("/mine", middlewares.RequirePermission(userRepo, models.PermAttendanceRequest), ctrl.Mine)
		// Viewing the full attendance list (everyone, or someone else
		// specifically) requires attendance.view.
		g.GET("", middlewares.RequirePermission(userRepo, models.PermAttendanceView), ctrl.List)
		// The Attendance Detail report is a dedicated report-only
		// endpoint (not shared with the List page above), so it's gated
		// by the Report permission instead of attendance.view.
		g.GET("/summary", middlewares.RequirePermission(userRepo, models.PermAttendanceReportView), ctrl.Summary)
		// Correcting an existing record requires attendance.edit.
		g.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermAttendanceEdit), ctrl.Update)
	}
}
