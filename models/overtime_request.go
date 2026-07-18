package models

import "time"

type OvertimeRequestStatus string

const (
	OvertimeRequestPending   OvertimeRequestStatus = "pending"
	OvertimeRequestApproved  OvertimeRequestStatus = "approved"
	OvertimeRequestRejected  OvertimeRequestStatus = "rejected"
	OvertimeRequestCancelled OvertimeRequestStatus = "cancelled"
)

// OvertimeRequest is a staff Overtime request — always a single date
// (unlike Leave, which can span a range for a Full Day), reviewed by an
// admin/manager.
type OvertimeRequest struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	UserID   uint    `gorm:"not null;index" json:"user_id"`
	User     *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BranchID *uint   `gorm:"index" json:"branch_id,omitempty"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	Date      string  `gorm:"type:date;not null;index" json:"date"`
	StartTime *string `gorm:"type:varchar(5)" json:"start_time,omitempty"` // "HH:MM"
	EndTime   *string `gorm:"type:varchar(5)" json:"end_time,omitempty"`
	// Duration is hours between StartTime/EndTime, computed and stored at
	// submission time (see OvertimeRequestService.Create) — nil if either
	// time wasn't provided, since there's nothing to compute from.
	Duration *float64 `gorm:"type:decimal(4,2)" json:"duration,omitempty"`

	Reason string `gorm:"type:varchar(500)" json:"reason,omitempty"`

	Status       OvertimeRequestStatus `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	ApprovedByID *uint                 `json:"approved_by_id,omitempty"`
	ApprovedBy   *User                 `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	ApprovedAt   *time.Time            `json:"approved_at,omitempty"`
	RejectReason string                `gorm:"type:varchar(500)" json:"reject_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
