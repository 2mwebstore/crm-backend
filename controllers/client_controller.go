package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/config"
	clientdto "crm-backend/dto/client"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type ClientController struct {
	svc      services.ClientService
	userSvc  services.UserService
	userRepo repositories.UserRepository
}

func NewClientController(svc services.ClientService, userSvc services.UserService, userRepo repositories.UserRepository) *ClientController {
	return &ClientController{svc, userSvc, userRepo}
}

// canViewPhone returns true if the caller has the phone.view permission (or is super admin).
func (ctrl *ClientController) canViewPhone(c *gin.Context) bool {
	if middlewares.IsSuperAdmin(c) {
		return true
	}
	userID := middlewares.GetUserID(c)
	user, err := ctrl.userRepo.FindByID(userID)
	if err != nil {
		return false
	}
	return user.HasPermission(models.PermPhoneView)
}

func maskClientPhones(clients []models.Client) {
	for i := range clients {
		for j := range clients[i].Phones {
			clients[i].Phones[j].Phone = utils.MaskPhone(clients[i].Phones[j].Phone)
		}
	}
}

func maskClientPhone(client *models.Client) {
	if client == nil {
		return
	}
	for i := range client.Phones {
		client.Phones[i].Phone = utils.MaskPhone(client.Phones[i].Phone)
	}
}

func (ctrl *ClientController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}
func (ctrl *ClientController) callerID(c *gin.Context) uint { return middlewares.GetUserID(c) }
func (ctrl *ClientController) clientID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}
func (ctrl *ClientController) subID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("sub_id"), 10, 64)
	return uint(id)
}

// ── Core CRUD ─────────────────────────────────────────────────────────────────

func (ctrl *ClientController) List(c *gin.Context) {
	var filter clientdto.ClientFilterQuery
	if err := utils.BindQuery(c, &filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	p := utils.ParsePagination(c)
	clients, total, err := ctrl.svc.List(filter, p, middlewares.GetUserID(c))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	if !ctrl.canViewPhone(c) {
		maskClientPhones(clients)
	}
	utils.OKPaginated(c, clients, utils.BuildMeta(p, total))
}

func (ctrl *ClientController) Create(c *gin.Context) {
	var req clientdto.CreateClientRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	client, err := ctrl.svc.Create(ctrl.callerID(c), req)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Created(c, "client created", client)
}

func (ctrl *ClientController) GetByID(c *gin.Context) {
	client, err := ctrl.svc.GetByID(ctrl.clientID(c), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "client")
		return
	}
	if !ctrl.canViewPhone(c) {
		maskClientPhone(client)
	}
	utils.OK(c, "success", client)
}

func (ctrl *ClientController) Update(c *gin.Context) {
	var req clientdto.UpdateClientRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	client, err := ctrl.svc.Update(ctrl.clientID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "client updated", client)
}

func (ctrl *ClientController) Delete(c *gin.Context) {
	if err := ctrl.svc.Delete(ctrl.clientID(c), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "client deleted", nil)
}

func (ctrl *ClientController) UploadPicture(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "file is required")
		return
	}
	if !utils.AllowedImageMime(fh.Header.Get("Content-Type")) {
		utils.BadRequest(c, "only jpeg, png, webp, gif images are allowed")
		return
	}
	cfg := config.Get()
	result, err := utils.SaveFile(fh, cfg.App.UploadDir, "clients/pictures", cfg.App.BaseURL)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	client, err := ctrl.svc.UploadPicture(ctrl.clientID(c), ctrl.scope(c), result.FileURL)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "picture uploaded", client)
}

// ── Bank Section ──────────────────────────────────────────────────────────────

func (ctrl *ClientController) AddBank(c *gin.Context) {
	var req clientdto.BankInput
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	bank, err := ctrl.svc.AddBank(ctrl.clientID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "bank added", bank)
}

func (ctrl *ClientController) UpdateBank(c *gin.Context) {
	var req clientdto.BankInput
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	bank, err := ctrl.svc.UpdateBank(ctrl.clientID(c), ctrl.subID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "bank updated", bank)
}

func (ctrl *ClientController) DeleteBank(c *gin.Context) {
	if err := ctrl.svc.DeleteBank(ctrl.clientID(c), ctrl.subID(c), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "bank deleted", nil)
}

// ── Product (Player) Section ──────────────────────────────────────────────────

func (ctrl *ClientController) AddProduct(c *gin.Context) {
	var req clientdto.ProductInput
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	product, err := ctrl.svc.AddProduct(ctrl.clientID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "product added", product)
}

func (ctrl *ClientController) UpdateProduct(c *gin.Context) {
	var req clientdto.ProductInput
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	product, err := ctrl.svc.UpdateProduct(ctrl.clientID(c), ctrl.subID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "product updated", product)
}

func (ctrl *ClientController) DeleteProduct(c *gin.Context) {
	if err := ctrl.svc.DeleteProduct(ctrl.clientID(c), ctrl.subID(c), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "product deleted", nil)
}

// ── Follow Up Section ─────────────────────────────────────────────────────────

func (ctrl *ClientController) AddFollowUp(c *gin.Context) {
	var req clientdto.FollowUpInput
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	fu, err := ctrl.svc.AddFollowUp(ctrl.clientID(c), ctrl.scope(c), ctrl.callerID(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Created(c, "follow-up logged", fu)
}

func (ctrl *ClientController) ListFollowUps(c *gin.Context) {
	p := utils.ParsePagination(c)
	list, total, err := ctrl.svc.ListFollowUps(ctrl.clientID(c), ctrl.scope(c), p.Page, p.PageSize)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OKPaginated(c, list, utils.BuildMeta(p, total))
}

func (ctrl *ClientController) DeleteFollowUp(c *gin.Context) {
	if err := ctrl.svc.DeleteFollowUp(ctrl.clientID(c), ctrl.subID(c), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "follow-up deleted", nil)
}

// CheckCode godoc — GET /clients/check-code?code=ABC&exclude_id=1
// Returns {"available": true/false}
func (ctrl *ClientController) CheckCode(c *gin.Context) {
	code := c.Query("code")
	excludeID := c.Query("exclude_id")
	if code == "" {
		utils.BadRequest(c, "code is required")
		return
	}
	available := ctrl.svc.CheckCodeAvailable(code, excludeID)
	utils.OK(c, "success", gin.H{"available": available})
}

// NextCode godoc — GET /clients/next-code?branch_id=1
func (ctrl *ClientController) NextCode(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	branchID, _ := strconv.ParseUint(branchIDStr, 10, 64)
	if branchID == 0 {
		utils.BadRequest(c, "branch_id is required")
		return
	}
	suffix := ctrl.svc.PeekNextSuffix(uint(branchID))
	utils.OK(c, "success", gin.H{"suffix": suffix})
}
