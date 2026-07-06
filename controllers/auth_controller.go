package controllers

import (
	"github.com/gin-gonic/gin"

	authdto "crm-backend/dto/auth"
	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type AuthController struct {
	svc services.AuthService
}

func NewAuthController(svc services.AuthService) *AuthController {
	return &AuthController{svc}
}

// Register godoc
// @Summary      Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body authdto.RegisterRequest true "Register payload"
// @Success      201 {object} utils.Response
// @Router       /auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req authdto.RegisterRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	user, err := ctrl.svc.Register(req)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}
	utils.Created(c, "user registered successfully", user)
}

// Login godoc
// @Summary      Login and receive JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body authdto.LoginRequest true "Login payload"
// @Success      200 {object} utils.Response
// @Router       /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req authdto.LoginRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	token, user, err := ctrl.svc.Login(req)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}
	utils.OK(c, "login successful", gin.H{"token": token, "user": user})
}

// GetProfile godoc
// @Summary      Get current user profile
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /auth/me [get]
func (ctrl *AuthController) GetProfile(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	user, err := ctrl.svc.GetProfile(userID)
	if err != nil {
		utils.NotFound(c, "user")
		return
	}
	utils.OK(c, "success", user)
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body authdto.UpdateProfileRequest true "Payload"
// @Success      200 {object} utils.Response
// @Router       /auth/profile [put]
func (ctrl *AuthController) UpdateProfile(c *gin.Context) {
	var req authdto.UpdateProfileRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	userID := middlewares.GetUserID(c)
	user, err := ctrl.svc.UpdateProfile(userID, req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "profile updated", user)
}

// ChangePassword godoc
// @Summary      Change password
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body authdto.ChangePasswordRequest true "Payload"
// @Success      200 {object} utils.Response
// @Router       /auth/change-password [post]
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	var req authdto.ChangePasswordRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := ctrl.svc.ChangePassword(middlewares.GetUserID(c), req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "password changed successfully", nil)
}
