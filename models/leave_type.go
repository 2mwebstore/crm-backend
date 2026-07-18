package models

import "time"

// LeaveType is an admin-configured leave category (e.g. "Sick Leave",
// "Annual Leave", "Day Off") — same lookup-table pattern as BankType/
// ProductType/etc. elsewhere in this app.
type LeaveType struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"type:varchar(100);not null" json:"name"`
	Code      string `gorm:"type:varchar(20);uniqueIndex" json:"code,omitempty"` // e.g. "SICK", "AL", "DAYOFF"
	IsActive  bool   `gorm:"default:true" json:"is_active"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`

	// BranchID assigns this leave type to a specific branch — nil means
	// global (visible to every branch), matching Company Bank's own
	// optional branch scoping. Not required.
	BranchID *uint   `gorm:"index" json:"branch_id,omitempty"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	// AnnualLimit/MonthlyLimit cap how many DAYS of this leave type a
	// single user can use within a calendar year / calendar month. Either
	// left nil means unlimited for that period — enforced in
	// LeaveRequestService.Create, not here (this is just the config).
	AnnualLimit  *int `json:"annual_limit,omitempty"`
	MonthlyLimit *int `json:"monthly_limit,omitempty"`

	// MonthlyUsed/AnnualUsed are NOT persisted (gorm:"-") — computed on
	// the fly by LeaveTypeService.List for whichever user is asking, so
	// the "Submit a Request" form can show "2 of 3 used this month"
	// alongside each leave type instead of the limit alone.
	MonthlyUsed *float64 `gorm:"-" json:"monthly_used,omitempty"`
	AnnualUsed  *float64 `gorm:"-" json:"annual_used,omitempty"`

	CreatedByID uint      `gorm:"index" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
