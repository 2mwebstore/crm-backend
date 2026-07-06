package routes

import (
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterTurnoverBetRoutes(rg *gin.RouterGroup, ctrl *controllers.TurnoverBetController, userRepo repositories.UserRepository) {
	g := rg.Group("/turnover-bets")
	g.Use(middlewares.Auth())
	{
		g.GET("", middlewares.RequirePermission(userRepo, models.PermTurnoverView), ctrl.List)
		g.POST("", middlewares.RequirePermission(userRepo, models.PermTurnoverCreate), ctrl.Create)
		g.GET("/:id", middlewares.RequirePermission(userRepo, models.PermTurnoverView), ctrl.GetByID)
		g.PUT("/:id", middlewares.RequirePermission(userRepo, models.PermTurnoverEdit), ctrl.Update)
		g.DELETE("/:id", middlewares.RequirePermission(userRepo, models.PermTurnoverDelete), ctrl.Delete)
		g.PUT("/:id/approve", middlewares.RequirePermission(userRepo, models.PermTurnoverApprove), ctrl.Approve)
	}
}
