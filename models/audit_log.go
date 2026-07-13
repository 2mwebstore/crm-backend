package models

import "time"

// AuditLog is an append-only record of every meaningful state-changing
// action taken in the app — who did it, from which branch (best-effort,
// not every action has an obvious branch), what it was, and what happened.
// Captured generically at the HTTP middleware level (see
// middlewares.AuditLog) rather than by hand in every single service
// method, so new endpoints are covered automatically without needing to
// remember to add logging calls to them.
type AuditLog struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// BranchID is best-effort — pulled from the request's own branch_id
	// field (body or query) when present. Many actions genuinely have no
	// branch context (e.g. managing roles), so this is nil for those.
	BranchID *uint   `gorm:"index" json:"branch_id,omitempty"`
	Branch   *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`

	Method     string `gorm:"type:varchar(10);not null" json:"method"`      // GET, POST, PUT, DELETE, PATCH
	Path       string `gorm:"type:varchar(255);not null;index" json:"path"` // e.g. /api/v1/deposits/5
	StatusCode int    `gorm:"not null" json:"status_code"`

	// RequestBody is the JSON body actually sent, with sensitive fields
	// (password, token, bot_token, secret, ...) redacted — see
	// middlewares.redactSensitiveFields. Truncated to a reasonable length
	// so one huge payload can't bloat this table unexpectedly.
	RequestBody string `gorm:"type:text" json:"request_body,omitempty"`

	IPAddress string `gorm:"type:varchar(64)" json:"ip_address,omitempty"`
	UserAgent string `gorm:"type:varchar(255)" json:"user_agent,omitempty"`

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
