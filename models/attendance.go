package models

import "time"

// Attendance is one user's check-in/check-out record for one calendar day
// (Asia/Phnom_Penh) at one branch — at most one row per (user_id, date).
type Attendance struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	UserID   uint    `gorm:"not null;uniqueIndex:idx_attendance_user_date" json:"user_id"`
	User     *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BranchID uint    `gorm:"not null;index" json:"branch_id"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	Date time.Time `gorm:"type:date;not null;uniqueIndex:idx_attendance_user_date" json:"date"` // Asia/Phnom_Penh calendar day, time-of-day component unused

	CheckInAt       *time.Time `json:"check_in_at,omitempty"`
	CheckInLat      *float64   `gorm:"type:decimal(10,7)" json:"check_in_lat,omitempty"`
	CheckInLng      *float64   `gorm:"type:decimal(10,7)" json:"check_in_lng,omitempty"`
	CheckInDistance *float64   `json:"check_in_distance,omitempty"` // meters from the branch's configured location
	// CheckInViaActivity records whether this check-in was let through
	// WITHOUT a normal in-range distance check, because the user had an
	// approved Activity request covering this date.
	CheckInViaActivity bool `gorm:"column:check_in_via_outdoor;default:false" json:"check_in_via_activity"`
	// CheckInReason is only ever populated when an Activity request
	// auto-drove this check-in (see
	// ActivityRequestController.Create) — carries that request's
	// own Reason over. Left blank for a normal manual check-in.
	CheckInReason string `gorm:"type:varchar(500)" json:"check_in_reason,omitempty"`
	// CheckInStatus labels timeliness against the user's own
	// User.ShiftCheckInTime — "early" (before shift time), "good" (shift time to
	// shift time + 15min), or "late" (beyond that). Computed once at
	// check-in and stored (see AttendanceService.CheckIn), not
	// recalculated on every read. Empty if the user has no ShiftTime
	// configured — nothing to compare against.
	CheckInStatus string `gorm:"type:varchar(10)" json:"check_in_status,omitempty"`

	CheckOutAt       *time.Time `json:"check_out_at,omitempty"`
	CheckOutLat      *float64   `gorm:"type:decimal(10,7)" json:"check_out_lat,omitempty"`
	CheckOutLng      *float64   `gorm:"type:decimal(10,7)" json:"check_out_lng,omitempty"`
	CheckOutDistance *float64   `json:"check_out_distance,omitempty"`
	// CheckOutViaActivity mirrors CheckInViaActivity for the check-out
	// side — true when this check-out was auto-driven by an Activity
	// request (see ActivityRequestController.Create) rather than a
	// manual self-service check-out.
	CheckOutViaActivity bool `gorm:"column:check_out_via_outdoor;default:false" json:"check_out_via_activity"`
	// CheckOutReason mirrors CheckInReason, for the check-out side.
	CheckOutReason string `gorm:"type:varchar(500)" json:"check_out_reason,omitempty"`
	// CheckOutStatus mirrors CheckInStatus, computed against the user's
	// User.ShiftCheckOutTime instead of ShiftCheckInTime.
	CheckOutStatus string `gorm:"type:varchar(10)" json:"check_out_status,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
