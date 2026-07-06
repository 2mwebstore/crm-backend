package models

import "time"

// ClientFollowUp records a follow-up interaction on a Client.
type ClientFollowUp struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientID      uint      `gorm:"not null;index" json:"client_id"`
	BranchID      *uint     `gorm:"index" json:"branch_id,omitempty"`
	Interest      bool      `gorm:"default:false" json:"interest"`
	GivenAccount  bool      `gorm:"default:false" json:"given_account"`
	BankAccount   bool      `gorm:"default:false" json:"bank_account"`
	Remark        string    `gorm:"type:text;not null" json:"remark"`
	FollowUpAt    time.Time `gorm:"not null" json:"follow_up_at"`
	BonusOptionID *uint     `gorm:"index" json:"bonus_option_id,omitempty"`
	CreatedByID   uint      `gorm:"not null;index" json:"created_by_id"`

	Client      *Client          `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Branch      *Branch          `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	BonusOption *BonusOptionType `gorm:"foreignKey:BonusOptionID" json:"bonus_option,omitempty"`
	CreatedBy   *User            `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
