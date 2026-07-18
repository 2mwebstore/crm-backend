package models

import "time"

// UserScheduleOverride temporarily replaces a user's normal
// ShiftCheckInTime/ShiftCheckOutTime for a specific date range — e.g. a
// week where someone is covering a different shift. AttendanceService
// checks for an active override covering "today" before falling back to
// the user's own User.ShiftCheckInTime/ShiftCheckOutTime.
type UserScheduleOverride struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	DateFrom string `gorm:"type:date;not null;index" json:"date_from"`
	DateTo   string `gorm:"type:date;not null" json:"date_to"`

	// Either left nil falls back to the user's own default for that side
	// — an override doesn't have to change both check-in and check-out.
	ShiftCheckInTime  *string `gorm:"type:varchar(5)" json:"shift_check_in_time,omitempty"`
	ShiftCheckOutTime *string `gorm:"type:varchar(5)" json:"shift_check_out_time,omitempty"`

	Reason string `gorm:"type:varchar(500)" json:"reason,omitempty"`

	CreatedByID uint      `gorm:"index" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
