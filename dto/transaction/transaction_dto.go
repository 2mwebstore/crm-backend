package transactiondto

import "crm-backend/utils"

type CreateRequest struct {
	TransactionNo   string         `json:"transaction_no"`
	Date            utils.FlexTime `json:"date" binding:"required"`
	ClientID        uint           `json:"client_id" binding:"required"`
	ClientProductID uint           `json:"client_product_id" binding:"required"`
	ClientBankID    uint           `json:"client_bank_id" binding:"required"`
	CompanyBankID   uint           `json:"company_bank_id" binding:"required"`
	Amount          float64        `json:"amount" binding:"required,gt=0"`
	Currency        string         `json:"currency" binding:"omitempty,oneof=USD KHR"`
	BranchID        *uint          `json:"branch_id"`
	BonusOptionID   *uint          `json:"bonus_option_id"`
	BonusAmount     float64        `json:"bonus_amount"`
	TO              float64        `json:"to"`
	OS              float64        `json:"os"`
	Bal             float64        `json:"bal"`
	Play            float64        `json:"play"`
	Remark          string         `json:"remark"`
}

type UpdateRequest struct {
	Date          *utils.FlexTime `json:"date"`
	ClientBankID  *uint           `json:"client_bank_id"`
	CompanyBankID *uint           `json:"company_bank_id"`
	BranchID      *uint           `json:"branch_id"`
	BonusOptionID *uint           `json:"bonus_option_id"`
	BonusAmount   *float64        `json:"bonus_amount"`
	Amount        *float64        `json:"amount"`
	TO            *float64        `json:"to"`
	OS            *float64        `json:"os"`
	Bal           *float64        `json:"bal"`
	Play          *float64        `json:"play"`
	Remark        *string         `json:"remark"`
}

type FilterQuery struct {
	Search            string `form:"search"`
	ClientID          *uint  `form:"client_id"`
	ClientProductID   *uint  `form:"client_product_id"`
	CompanyBankID     *uint  `form:"company_bank_id"`
	BranchID          *uint  `form:"branch_id"`
	CreatedByID       *uint  `form:"created_by_id"`
	ApprovedByID      *uint  `form:"approved_by_id"`
	CompanyBankTypeID *uint  `form:"bank_type_id"`
	ProductTypeID     *uint  `form:"product_type_id"`
	Status            string `form:"status"`
	DateFrom          string `form:"date_from"`
	DateTo            string `form:"date_to"`
	Currency          string `form:"currency"`
	SortBy            string `form:"sort_by"`
	SortDir           string `form:"sort_dir"`
}

type BalanceResponse struct {
	ClientID         uint    `json:"client_id"`
	ClientProductID  uint    `json:"client_product_id"`
	Currency         string  `json:"currency"`
	TotalDeposits    float64 `json:"total_deposits"`
	TotalWithdrawals float64 `json:"total_withdrawals"`
	CurrentBalance   float64 `json:"current_balance"`
}

// ApproveRequest — PUT /deposits/:id/approve or /withdrawals/:id/approve
type ApproveRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}
