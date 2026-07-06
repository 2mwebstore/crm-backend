package models

import "time"

// EntityType identifies which entity the code sequence belongs to.
type EntityType string

const (
	EntityClient      EntityType = "client"
	EntityIC          EntityType = "interesting_client"
	EntityDeposit     EntityType = "deposit"
	EntityWithdrawal  EntityType = "withdrawal"
)

// CodeSequence stores the running counter for code generation per branch + entity type.
// e.g. branch "PHNM" + entity "client" → last_seq = 3 → next code = "PHNM0000004"
type CodeSequence struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	BranchID   uint       `gorm:"not null;uniqueIndex:idx_branch_entity" json:"branch_id"`
	EntityType EntityType `gorm:"type:varchar(50);not null;uniqueIndex:idx_branch_entity" json:"entity_type"`
	LastSeq    uint64     `gorm:"not null;default:0" json:"last_seq"`
	Branch     *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
