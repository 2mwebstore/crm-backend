package models

import "time"

// Branch represents an organizational branch.
// Super Admin assigns branches to users (many-to-many via user_branches).
// The Code field is used as a prefix in document code generation:
//
//	Format: {ENTITY_PREFIX}-{YYYYMMDD}-{BRANCH_CODE}
//	e.g.   INT-20260701-CRNS
type Branch struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Code        string `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"` // short code, e.g. "CRNS"
	Description string `gorm:"type:varchar(500)" json:"description,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Telegram notification target for this branch — when set, every
	// Deposit/Withdrawal created against this branch posts a message to
	// this bot/group. Notifications are simply skipped if
	// TelegramBotToken or TelegramChatID is empty.
	TelegramBotToken string `gorm:"type:varchar(255)" json:"telegram_bot_token,omitempty"`
	// TelegramChatID is the target group's chat ID (e.g. "-1001234567890"
	// for a supergroup) — Telegram's own numeric ID format, not this app's.
	// Shared by both Deposit and Withdrawal notifications — it's the same
	// group either way, just routed to a different topic within it.
	TelegramChatID string `gorm:"type:varchar(50)" json:"telegram_chat_id,omitempty"`
	// TelegramDepositTopicID/TelegramWithdrawalTopicID are the forum
	// topic/thread IDs within that group for each transaction type — a
	// single group commonly has separate topics per transaction type
	// (e.g. one topic for Withdrawals, another for Deposits). Either can
	// be left nil to post that type to the group's General topic (or a
	// non-forum group) instead of a specific thread.
	TelegramDepositTopicID    *int `json:"telegram_deposit_topic_id,omitempty"`
	TelegramWithdrawalTopicID *int `json:"telegram_withdrawal_topic_id,omitempty"`

	// Attendance geofence — where check-in/check-out distance is measured
	// from. Both nil means this branch has no location configured yet,
	// which the attendance service treats as "can't validate distance" —
	// see AttendanceService for how that's handled (fails closed, not
	// silently allowed).
	Latitude  *float64 `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
	Longitude *float64 `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`
	// CheckInRadiusMeters is how far from (Latitude, Longitude) a normal
	// check-in/check-out is still allowed. 0 falls back to a sane default
	// (200m) at the service layer rather than allowing an unbounded
	// radius by accident.
	CheckInRadiusMeters int `gorm:"default:200" json:"check_in_radius_meters"`

	CreatedByID uint      `gorm:"index;default:0" json:"created_by_id"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
