package models

import "time"

type LeaveRequestStatus string

const (
	LeaveRequestPending   LeaveRequestStatus = "pending"
	LeaveRequestApproved  LeaveRequestStatus = "approved"
	LeaveRequestRejected  LeaveRequestStatus = "rejected"
	LeaveRequestCancelled LeaveRequestStatus = "cancelled"
)

// LeaveRequestDayType — Full Day can span DateFrom→DateTo; either Half Day
// is always a single date (DateFrom must equal DateTo) and counts as 0.5
// days for LeaveType.AnnualLimit/MonthlyLimit purposes.
type LeaveRequestDayType string

const (
	LeaveDayFull          LeaveRequestDayType = "full"
	LeaveDayHalfMorning   LeaveRequestDayType = "half_morning"
	LeaveDayHalfAfternoon LeaveRequestDayType = "half_afternoon"
)

// LeaveRequest is a staff Leave request — its own table (not merged with
// Overtime/Activity), reviewed by an admin/manager.
type LeaveRequest struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	UserID   uint    `gorm:"not null;index" json:"user_id"`
	User     *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BranchID *uint   `gorm:"index" json:"branch_id,omitempty"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	LeaveTypeID uint       `gorm:"not null;index" json:"leave_type_id"`
	LeaveType   *LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`

	DayType  LeaveRequestDayType `gorm:"type:varchar(20);not null;default:full" json:"day_type"`
	DateFrom string              `gorm:"type:date;not null;index" json:"date_from"`
	DateTo   string              `gorm:"type:date;not null" json:"date_to"`

	// Duration is computed and stored at submission time — 0.5 for a Half
	// Day, otherwise the actual day count between DateFrom/DateTo
	// inclusive. Persisted (not recomputed on every read) so a later
	// change to the day-counting logic can never silently alter what an
	// already-submitted request's duration was reported as.
	Duration float64 `gorm:"type:decimal(4,1);not null" json:"duration"`

	Reason string `gorm:"type:varchar(500)" json:"reason,omitempty"`

	Status       LeaveRequestStatus `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	ApprovedByID *uint              `json:"approved_by_id,omitempty"`
	ApprovedBy   *User              `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	ApprovedAt   *time.Time         `json:"approved_at,omitempty"`
	RejectReason string             `gorm:"type:varchar(500)" json:"reject_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
