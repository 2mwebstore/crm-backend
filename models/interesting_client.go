package models

import "time"

// InterestingClient — lightweight prospect/lead record.
// Keeps only core identity + contact info + phones.
type InterestingClient struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	// ── Core fields ────────────────────────────────────────────────────────
	FullName   string     `gorm:"type:varchar(191);not null" json:"full_name"`
	DateJoined *time.Time `json:"date_joined,omitempty"`
	Remark     string     `gorm:"type:text" json:"remark,omitempty"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`

	// ── Foreign Keys ───────────────────────────────────────────────────────
	BranchID        *uint `gorm:"index" json:"branch_id,omitempty"`
	ContactSourceID *uint `gorm:"index" json:"contact_source_id,omitempty"`
	CreatedByID     uint  `gorm:"not null;index" json:"created_by_id"`

	// ── Conversion tracking ────────────────────────────────────────────────
	IsConverted       bool       `gorm:"default:false" json:"is_converted"`
	ConvertedAt       *time.Time `json:"converted_at,omitempty"`
	ConvertedClientID *uint      `gorm:"index" json:"converted_client_id,omitempty"`

	// ── Relations ──────────────────────────────────────────────────────────
	Branch        *Branch                  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ContactSource *ContactSource           `gorm:"foreignKey:ContactSourceID" json:"contact_source,omitempty"`
	CreatedBy     *User                    `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Phones        []InterestingClientPhone `gorm:"foreignKey:InterestingClientID" json:"phones,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
