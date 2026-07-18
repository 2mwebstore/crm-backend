package routes

import (
	"github.com/gin-gonic/gin"

	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
)

func RegisterLeaveRequestRoutes(rg *gin.RouterGroup, ctrl *controllers.LeaveRequestController, userRepo repositories.UserRepository) {
	g := rg.Group("/leave-requests")
	g.Use(middlewares.Auth())
	{
		// Submitting/managing your own Leave requests requires
		// leave_requests.request — a sibling to View/Approve below, not
		// a prerequisite for them.
		g.POST("", middlewares.RequirePermission(userRepo, models.PermLeaveRequestSubmit), ctrl.Create)
		g.PATCH("/:id", middlewares.RequirePermission(userRepo, models.PermLeaveRequestSubmit), ctrl.EditReason)
		g.POST("/:id/cancel", middlewares.RequirePermission(userRepo, models.PermLeaveRequestSubmit), ctrl.Cancel)
		g.GET("/mine", middlewares.RequirePermission(userRepo, models.PermLeaveRequestSubmit), ctrl.Mine)
		g.GET("", middlewares.RequirePermission(userRepo, models.PermLeaveRequestView), ctrl.List)
		g.GET("/:id", middlewares.RequirePermission(userRepo, models.PermLeaveRequestView), ctrl.GetByID)
		g.POST("/:id/approve", middlewares.RequirePermission(userRepo, models.PermLeaveRequestApprove), ctrl.Approve)
		g.POST("/:id/reject", middlewares.RequirePermission(userRepo, models.PermLeaveRequestApprove), ctrl.Reject)
	}
}
