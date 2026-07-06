package turnoverbetdto

import "crm-backend/utils"

type CreateRequest struct {
	BranchID      *uint          `json:"branch_id"`
	Date          utils.FlexTime `json:"date" binding:"required"`
	ProductTypeID uint           `json:"product_type_id" binding:"required"`
	Amount        float64        `json:"amount" binding:"required,gt=0"`
	Currency      string         `json:"currency" binding:"omitempty,oneof=USD KHR"`
	Remark        string         `json:"remark"`
}

type UpdateRequest struct {
	Date          *utils.FlexTime `json:"date"`
	ProductTypeID *uint           `json:"product_type_id"`
	Amount        *float64        `json:"amount" binding:"omitempty,gt=0"`
	Currency      *string         `json:"currency" binding:"omitempty,oneof=USD KHR"`
	Remark        *string         `json:"remark"`
}

type ApproveRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

type FilterQuery struct {
	ProductTypeID *uint  `form:"product_type_id"`
	BranchID      *uint  `form:"branch_id"`
	CreatedByID   *uint  `form:"created_by_id"`
	ApprovedByID  *uint  `form:"approved_by_id"`
	DateFrom      string `form:"date_from"`
	DateTo        string `form:"date_to"`
	Status        string `form:"status"`
	Currency      string `form:"currency"`
	SortBy        string `form:"sort_by"`
	SortDir       string `form:"sort_dir"`
}
