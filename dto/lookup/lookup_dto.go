package lookupdto

// ── BankType ──────────────────────────────────────────────────────────────────

type CreateBankTypeRequest struct {
	BranchID    *uint  `json:"branch_id"`
	Name        string `json:"name" binding:"required,min=1,max=191"`
	Code        string `json:"code" binding:"required,min=1,max=50"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateBankTypeRequest struct {
	BranchID    *uint   `json:"branch_id"`
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Logo        *string `json:"logo"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

// ── ProductType ───────────────────────────────────────────────────────────────

type CreateProductTypeRequest struct {
	BranchID    *uint  `json:"branch_id"`
	Name        string `json:"name" binding:"required,min=1,max=191"`
	Code        string `json:"code" binding:"required,min=1,max=50"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateProductTypeRequest struct {
	BranchID    *uint   `json:"branch_id"`
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

// ── BonusOptionType ───────────────────────────────────────────────────────────

type CreateBonusOptionTypeRequest struct {
	BranchID    *uint   `json:"branch_id"`
	Name        string  `json:"name" binding:"required,min=1,max=191"`
	Code        string  `json:"code" binding:"required,min=1,max=50"`
	Description string  `json:"description"`
	CalcType    string  `json:"calc_type" binding:"required,oneof=fixed percentage"`
	BonusValue  float64 `json:"bonus_value" binding:"required,min=0"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateBonusOptionTypeRequest struct {
	BranchID    *uint    `json:"branch_id"`
	Name        *string  `json:"name"`
	Code        *string  `json:"code"`
	Description *string  `json:"description"`
	CalcType    *string  `json:"calc_type" binding:"omitempty,oneof=fixed percentage"`
	BonusValue  *float64 `json:"bonus_value" binding:"omitempty,min=0"`
	SortOrder   *int     `json:"sort_order"`
	IsActive    *bool    `json:"is_active"`
}

// ── CurrencyType ──────────────────────────────────────────────────────────────

type CreateCurrencyTypeRequest struct {
	Code      string `json:"code" binding:"required,min=1,max=10"`
	Name      string `json:"name" binding:"required,min=1,max=100"`
	Symbol    string `json:"symbol"`
	IsBase    bool   `json:"is_base"`
	SortOrder int    `json:"sort_order"`
}

type UpdateCurrencyTypeRequest struct {
	Name      *string `json:"name"`
	Symbol    *string `json:"symbol"`
	IsBase    *bool   `json:"is_base"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

// ConvertRequest is used to convert an amount between USD and KHR.
type ConvertRequest struct {
	Amount   float64 `json:"amount" binding:"required,min=0"`
	From     string  `json:"from" binding:"required,oneof=USD KHR"`
	To       string  `json:"to" binding:"required,oneof=USD KHR"`
	RateDate string  `json:"rate_date"` // optional — uses latest if empty
}

// ConvertResponse is returned by the conversion endpoint.
type ConvertResponse struct {
	From            string  `json:"from"`
	To              string  `json:"to"`
	OriginalAmount  float64 `json:"original_amount"`
	ConvertedAmount float64 `json:"converted_amount"`
	Rate            float64 `json:"rate"`
	RateDate        string  `json:"rate_date"`
}

// ── Role ──────────────────────────────────────────────────────────────────────

type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=191"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}
