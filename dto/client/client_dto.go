package clientdto

import "crm-backend/utils"

type PhoneInput struct {
	ID        *uint  `json:"id"`
	Phone     string `json:"phone" binding:"required"`
	Label     string `json:"label"`
	IsPrimary bool   `json:"is_primary"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

type BankInput struct {
	ID          *uint  `json:"id"`
	BankTypeID  uint   `json:"bank_type_id" binding:"required"`
	AccountNo   string `json:"account_no" binding:"required"`
	AccountName string `json:"account_name" binding:"required"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type ProductInput struct {
	ID            *uint  `json:"id"`
	ProductTypeID uint   `json:"product_type_id" binding:"required"`
	AccountID     string `json:"account_id" binding:"required"`
	IsActive      bool   `json:"is_active"`
	SortOrder     int    `json:"sort_order"`
}

type CreateClientRequest struct {
	Code            string          `json:"code"` // branch prefix + suffix e.g. CRNS-C000001 (blank = auto-generate)
	Name            string          `json:"name" binding:"required,min=1,max=191"`
	DateJoined      *utils.FlexTime `json:"date_joined"`
	Remark          string          `json:"remark"`
	IsActive        *bool           `json:"is_active"`
	BranchID        *uint           `json:"branch_id"`
	LevelID         *uint           `json:"level_id"`
	ContactSourceID *uint           `json:"contact_source_id"`
	Phones          []PhoneInput    `json:"phones"`
	Banks           []BankInput     `json:"banks"`
	Products        []ProductInput  `json:"products"`
}

type UpdateClientRequest struct {
	Code            *string         `json:"code"`
	Name            *string         `json:"name"`
	DateJoined      *utils.FlexTime `json:"date_joined"`
	Remark          *string         `json:"remark"`
	IsActive        *bool           `json:"is_active"`
	BranchID        *uint           `json:"branch_id"`
	LevelID         *uint           `json:"level_id"`
	ContactSourceID *uint           `json:"contact_source_id"`
	Phones          []PhoneInput    `json:"phones"`
	Banks           []BankInput     `json:"banks"`
	Products        []ProductInput  `json:"products"`
}

type ClientFilterQuery struct {
	Search          string `form:"search"`
	IsActive        *bool  `form:"is_active"`
	LevelID         *uint  `form:"level_id"`
	ContactSourceID *uint  `form:"contact_source_id"`
	BranchID        *uint  `form:"branch_id"`
	CreatedByID     *uint  `form:"created_by_id"`
	DateFrom        string `form:"date_from"`
	DateTo          string `form:"date_to"`
	SortBy          string `form:"sort_by"`
	SortDir         string `form:"sort_dir"`
}

type FollowUpInput struct {
	Interest     bool           `json:"interest"`
	GivenAccount bool           `json:"given_account"`
	BankAccount  bool           `json:"bank_account"`
	Remark       string         `json:"remark" binding:"required"`
	FollowUpAt   utils.FlexTime `json:"follow_up_at" binding:"required"`
}
