package followupdto

import "crm-backend/utils"

type CreateRequest struct {
	ClientID      uint           `json:"client_id" binding:"required"`
	BranchID      *uint          `json:"branch_id"`
	FollowUpAt    utils.FlexTime `json:"follow_up_at" binding:"required"`
	BonusOptionID *uint          `json:"bonus_option_id"`
	Interest      bool           `json:"interest"`
	GivenAccount  bool           `json:"given_account"`
	BankAccount   bool           `json:"bank_account"`
	Remark        string         `json:"remark" binding:"required"`
}

type FilterQuery struct {
	ClientID      *uint  `form:"client_id"`
	BranchID      *uint  `form:"branch_id"`
	CreatedByID   *uint  `form:"created_by_id"`
	BonusOptionID *uint  `form:"bonus_option_id"`
	DateFrom      string `form:"date_from"`
	DateTo        string `form:"date_to"`
	SortBy        string `form:"sort_by"`
	SortDir       string `form:"sort_dir"`
}
