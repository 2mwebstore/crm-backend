package controllers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"crm-backend/middlewares"
	"crm-backend/services"
	"crm-backend/utils"
)

type PayrollExportController struct {
	svc services.PayrollExportService
}

func NewPayrollExportController(svc services.PayrollExportService) *PayrollExportController {
	return &PayrollExportController{svc}
}

// Export godoc — GET /attendance/payroll-export?date_from=&date_to=&user_id=&branch_id=
// Streams the generated .xlsx workbook directly as a file download.
func (ctrl *PayrollExportController) Export(c *gin.Context) {
	var userID, branchID uint
	if v, err := strconv.ParseUint(c.Query("user_id"), 10, 64); err == nil {
		userID = uint(v)
	}
	if v, err := strconv.ParseUint(c.Query("branch_id"), 10, 64); err == nil {
		branchID = uint(v)
	}

	f, err := ctrl.svc.Generate(middlewares.GetUserID(c), c.Query("date_from"), c.Query("date_to"), userID, branchID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	filename := fmt.Sprintf("payroll-%s-to-%s.xlsx", c.Query("date_from"), c.Query("date_to"))
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := f.Write(c.Writer); err != nil {
		utils.InternalError(c, err)
	}
}
