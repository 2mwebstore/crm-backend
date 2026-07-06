package interestingdto

import "crm-backend/utils"

type PhoneInput struct {
	ID        *uint  `json:"id"`
	Phone     string `json:"phone" binding:"required"`
	Label     string `json:"label"`
	IsPrimary bool   `json:"is_primary"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status"`
}

type CreateRequest struct {
	Code            string          `json:"code"` // branch prefix + suffix e.g. CRNS001 (blank = auto-generate)
	FullName        string          `json:"full_name" binding:"required,min=1,max=191"`
	DateJoined      *utils.FlexTime `json:"date_joined"`
	Remark          string          `json:"remark"`
	IsActive        *bool           `json:"is_active"`
	BranchID        *uint           `json:"branch_id"`
	ContactSourceID *uint           `json:"contact_source_id"`
	Phones          []PhoneInput    `json:"phones"`
}

type UpdateRequest struct {
	Code            *string         `json:"code"`
	FullName        *string         `json:"full_name"`
	DateJoined      *utils.FlexTime `json:"date_joined"`
	Remark          *string         `json:"remark"`
	IsActive        *bool           `json:"is_active"`
	BranchID        *uint           `json:"branch_id"`
	ContactSourceID *uint           `json:"contact_source_id"`
	Phones          []PhoneInput    `json:"phones"`
}

type FilterQuery struct {
	Search          string `form:"search"`
	IsActive        *bool  `form:"is_active"`
	IsConverted     *bool  `form:"is_converted"`
	CreatedByID     *uint  `form:"created_by_id"`
	BranchID        *uint  `form:"branch_id"`
	ContactSourceID *uint  `form:"contact_source_id"`
	DateFrom        string `form:"date_from"`
	DateTo          string `form:"date_to"`
	SortBy          string `form:"sort_by"`
	SortDir         string `form:"sort_dir"`
}

type ConvertRequest struct {
	ExistingClientID *uint  `json:"existing_client_id"`
	Code             string `json:"code"`
	BranchID         *uint  `json:"branch_id"`
}
