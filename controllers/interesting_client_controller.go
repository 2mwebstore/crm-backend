package controllers

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	interestingdto "crm-backend/dto/interesting_client"
	"crm-backend/middlewares"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/services"
	"crm-backend/utils"
)

type InterestingClientController struct {
	svc        services.InterestingClientService
	userSvc    services.UserService
	clientRepo repositories.ClientRepository
	userRepo   repositories.UserRepository
	db         *gorm.DB
}

func NewInterestingClientController(svc services.InterestingClientService, userSvc services.UserService, clientRepo repositories.ClientRepository, userRepo repositories.UserRepository, db *gorm.DB) *InterestingClientController {
	return &InterestingClientController{svc, userSvc, clientRepo, userRepo, db}
}

func (ctrl *InterestingClientController) canViewPhone(c *gin.Context) bool {
	if middlewares.IsSuperAdmin(c) {
		return true
	}
	user, err := ctrl.userRepo.FindByID(middlewares.GetUserID(c))
	if err != nil {
		return false
	}
	return user.HasPermission(models.PermPhoneView)
}

func (ctrl *InterestingClientController) scope(c *gin.Context) []uint {
	ids, _ := ctrl.userSvc.GetUserBranchIDs(middlewares.GetUserID(c))
	return ids
}
func (ctrl *InterestingClientController) callerID(c *gin.Context) uint {
	return middlewares.GetUserID(c)
}
func (ctrl *InterestingClientController) icID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}
func (ctrl *InterestingClientController) subID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("sub_id"), 10, 64)
	return uint(id)
}

func (ctrl *InterestingClientController) List(c *gin.Context) {
	var filter interestingdto.FilterQuery
	if err := utils.BindQuery(c, &filter); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	p := utils.ParsePagination(c)
	list, total, err := ctrl.svc.List(filter, p, middlewares.GetUserID(c))
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	if !ctrl.canViewPhone(c) {
		for i := range list {
			for j := range list[i].Phones {
				list[i].Phones[j].Phone = utils.MaskPhone(list[i].Phones[j].Phone)
			}
		}
	}
	utils.OKPaginated(c, list, utils.BuildMeta(p, total))
}

func (ctrl *InterestingClientController) Create(c *gin.Context) {
	var req interestingdto.CreateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ic, err := ctrl.svc.Create(ctrl.callerID(c), req)
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Created(c, "interesting client created", ic)
}

func (ctrl *InterestingClientController) GetByID(c *gin.Context) {
	ic, err := ctrl.svc.GetByID(ctrl.icID(c), ctrl.scope(c))
	if err != nil {
		utils.NotFound(c, "interesting client")
		return
	}
	if !ctrl.canViewPhone(c) {
		for j := range ic.Phones {
			ic.Phones[j].Phone = utils.MaskPhone(ic.Phones[j].Phone)
		}
	}
	utils.OK(c, "success", ic)
}

func (ctrl *InterestingClientController) Update(c *gin.Context) {
	var req interestingdto.UpdateRequest
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ic, err := ctrl.svc.Update(ctrl.icID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "updated", ic)
}

func (ctrl *InterestingClientController) Delete(c *gin.Context) {
	if err := ctrl.svc.Delete(ctrl.icID(c), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "deleted", nil)
}

func (ctrl *InterestingClientController) Convert(c *gin.Context) {
	var req interestingdto.ConvertRequest
	_ = utils.BindJSON(c, &req)
	client, err := ctrl.svc.Convert(ctrl.icID(c), ctrl.scope(c), req, ctrl.callerID(c), ctrl.clientRepo)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "converted to client successfully", client)
}

// ── Phone management ──────────────────────────────────────────────────────────

func (ctrl *InterestingClientController) UpdatePhone(c *gin.Context) {
	var req interestingdto.PhoneInput
	if err := utils.BindJSON(c, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	p, err := ctrl.svc.UpdatePhone(ctrl.icID(c), ctrl.subID(c), ctrl.scope(c), req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "phone updated", p)
}

func (ctrl *InterestingClientController) DeletePhone(c *gin.Context) {
	if err := ctrl.svc.DeletePhone(ctrl.icID(c), ctrl.subID(c), ctrl.scope(c)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.OK(c, "phone deleted", nil)
}

// CheckCode godoc — GET /interesting-clients/check-code?code=INT001&exclude_id=1
func (ctrl *InterestingClientController) CheckCode(c *gin.Context) {
	code := c.Query("code")
	excludeID := c.Query("exclude_id")
	if code == "" {
		utils.BadRequest(c, "code is required")
		return
	}
	utils.OK(c, "success", gin.H{"available": ctrl.svc.CheckCodeAvailable(code, excludeID)})
}

// NextCode godoc — GET /interesting-clients/next-code?branch_id=1
func (ctrl *InterestingClientController) NextCode(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	branchID, _ := strconv.ParseUint(branchIDStr, 10, 64)
	if branchID == 0 {
		utils.BadRequest(c, "branch_id is required")
		return
	}
	suffix := ctrl.svc.PeekNextSuffix(uint(branchID))
	utils.OK(c, "success", gin.H{"suffix": suffix})
}

// PreviewCode godoc — GET /interesting-clients/preview-code?branch_id=1
// Returns the next code that would be generated for this branch without incrementing.
func (ctrl *InterestingClientController) PreviewCode(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	if branchIDStr == "" {
		utils.BadRequest(c, "branch_id is required")
		return
	}
	var branchID uint
	fmt.Sscanf(branchIDStr, "%d", &branchID)

	preview := utils.PeekICNextCode(ctrl.db, branchID, utils.EntityIC)
	utils.OK(c, "success", gin.H{"code": preview})
}
