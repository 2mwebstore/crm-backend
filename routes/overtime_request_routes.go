package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterOvertimeRequestRoutes(rg *gin.RouterGroup, ctrl *controllers.OvertimeRequestController, userRepo repositories.UserRepository) {
	g := rg.Group("/overtime-requests")
	g.Use(middlewares.Auth())
	{
		// Submitting/managing your own Overtime requests requires
		// overtime_requests.request — a sibling to View/Approve below,
		// not a prerequisite for them.
		g.POST("", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestSubmit), ctrl.Create)
		g.PATCH("/:id", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestSubmit), ctrl.EditReason)
		g.POST("/:id/cancel", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestSubmit), ctrl.Cancel)
		g.GET("/mine", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestSubmit), ctrl.Mine)
		g.GET("", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestView), ctrl.List)
		g.GET("/:id", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestView), ctrl.GetByID)
		g.POST("/:id/approve", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestApprove), ctrl.Approve)
		g.POST("/:id/reject", middlewares.RequirePermission(userRepo, models.PermOvertimeRequestApprove), ctrl.Reject)
	}
}
