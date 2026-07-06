package utils

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ── Entity type constants ──────────────────────────────────────────────────────

type EntityType = string

const (
	EntityClient      EntityType = "client"
	EntityIC          EntityType = "interesting_client"
	EntityDeposit     EntityType = "deposit"
	EntityWithdrawal  EntityType = "withdrawal"
	EntityTurnoverBet EntityType = "turnover_bet"
	EntityFollowUp    EntityType = "follow_up"
)

// entityPrefix returns the 3-letter prefix for each entity type.
func entityPrefix(entity EntityType) string {
	switch entity {
	case EntityClient:
		return "CLT"
	case EntityIC:
		return "INT"
	case EntityDeposit:
		return "DEP"
	case EntityWithdrawal:
		return "WDR"
	case EntityTurnoverBet:
		return "TB"
	case EntityFollowUp:
		return "FU"
	default:
		return "GEN"
	}
}

// ── Code generation ────────────────────────────────────────────────────────────
//
// If user has a branch assigned → per-branch counter
//   Format: {PREFIX}-{000001}  e.g. INT-000001
//   Each branch has its own counter per entity type stored in code_sequences.
//   branch_id = actual branch ID from user_branches
//
// If user has NO branch → global counter (branch_id = 0)
//   Same format: INT-000001 but shared across all no-branch users

// GenerateCode produces a sequential code for the given user + entity.
// Counter is per-branch (or global if no branch assigned).
func GenerateCode(db *gorm.DB, userID uint, entity EntityType) string {
	if db == nil {
		return fallbackCode(entity)
	}

	// Get the branch ID for this user (walk up parent chain)
	branchID := getUserBranchID(db, userID)
	// branch_id = 0 means global counter (no branch assigned)

	return nextSequentialCode(db, branchID, entity)
}

// getUserBranchID walks up the parent chain to find the first branch assigned.
// Returns 0 if no branch found (will use global counter).
func getUserBranchID(db *gorm.DB, userID uint) uint {
	type row struct {
		ID           uint
		ParentID     *uint
		IsSuperAdmin bool
	}

	current := userID
	for i := 0; i < 10; i++ {
		var u row
		if err := db.Table("users").
			Select("id, parent_id, is_super_admin").
			Where("id = ?", current).
			First(&u).Error; err != nil {
			break
		}

		// Super Admin has no branch — use global counter
		if u.IsSuperAdmin {
			return 0
		}

		// Check if this user has a branch assigned
		var branchID uint
		db.Raw(`SELECT branch_id FROM user_branches WHERE user_id = ? LIMIT 1`, u.ID).Scan(&branchID)
		if branchID != 0 {
			return branchID
		}

		if u.ParentID == nil {
			break
		}
		current = *u.ParentID
	}
	return 0 // no branch found → global counter
}

// nextSequentialCode atomically increments the counter for (branchID, entity)
// and returns the formatted code.
func nextSequentialCode(db *gorm.DB, branchID uint, entity EntityType) string {
	err := db.Exec(`
		INSERT INTO code_sequences (branch_id, entity_type, last_seq, updated_at)
		VALUES (?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE last_seq = last_seq + 1, updated_at = NOW()
	`, branchID, entity).Error
	if err != nil {
		return fallbackCode(entity)
	}

	var seq uint64
	if err := db.Raw(`
		SELECT last_seq FROM code_sequences
		WHERE branch_id = ? AND entity_type = ?
	`, branchID, entity).Scan(&seq).Error; err != nil || seq == 0 {
		return fallbackCode(entity)
	}

	return fmt.Sprintf("%s-%06d", entityPrefix(entity), seq)
}

// ── Fallback (when db unavailable) ────────────────────────────────────────────

func fallbackCode(entity EntityType) string {
	n := rand.Intn(999999) + 1
	return fmt.Sprintf("%s-%06d", entityPrefix(entity), n)
}

// Legacy helpers kept for backward compatibility
func GenerateClientCode() string      { return fallbackCode(EntityClient) }
func GenerateInterestingCode() string { return fallbackCode(EntityIC) }
func GenerateDepositCode() string     { return fallbackCode(EntityDeposit) }
func GenerateWithdrawalCode() string  { return fallbackCode(EntityWithdrawal) }

// PeekNextCode returns what the NEXT code will be for a branch+entity
// WITHOUT incrementing the counter. Used by the frontend to show a preview.
// Returns the zero-padded suffix only (e.g. "001"), not the full code.
func PeekNextSuffix(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return "001"
	}
	var seq uint64
	db.Raw(`
		SELECT COALESCE(last_seq, 0) FROM code_sequences
		WHERE branch_id = ? AND entity_type = ?
	`, branchID, entity).Scan(&seq)
	return fmt.Sprintf("%03d", seq+1)
}

// PeekNextCode returns what the next code WOULD be for a branch+entity,
// without actually incrementing the counter. Used for preview in the frontend.
func PeekNextCode(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return fmt.Sprintf("%s-000001", entityPrefix(entity))
	}
	var seq uint64
	db.Raw(`SELECT COALESCE(last_seq, 0) FROM code_sequences WHERE branch_id = ? AND entity_type = ?`, branchID, entity).Scan(&seq)
	return fmt.Sprintf("%s-%06d", entityPrefix(entity), seq+1)
}

// GenerateCodeForBranch generates code using a specific branch (not user's assigned branch).
// Used when user explicitly selects a branch during IC creation.
func GenerateCodeForBranch(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return fallbackCode(entity)
	}
	return nextSequentialCode(db, branchID, entity)
}

// GenerateICCodeForBranch generates IC code using branch code as prefix.
// Format: {BRANCHCODE}-{000001}  e.g. CRNS-000001
func GenerateICCodeForBranch(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return fallbackCode(entity)
	}
	// Get branch code
	var branchCode string
	if err := db.Raw("SELECT code FROM branches WHERE id = ?", branchID).Scan(&branchCode).Error; err != nil || branchCode == "" {
		return nextSequentialCode(db, branchID, entity)
	}
	// Atomic increment
	if err := db.Exec(`
		INSERT INTO code_sequences (branch_id, entity_type, last_seq, updated_at)
		VALUES (?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE last_seq = last_seq + 1, updated_at = NOW()
	`, branchID, entity).Error; err != nil {
		return fallbackCode(entity)
	}
	var seq uint64
	db.Raw("SELECT last_seq FROM code_sequences WHERE branch_id = ? AND entity_type = ?", branchID, entity).Scan(&seq)
	// Format: CRNS-000001
	return fmt.Sprintf("%s-%03d", branchCode, seq)
}

// PeekICNextCode previews the next IC code for a branch without incrementing.
func PeekICNextCode(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return "—"
	}
	var branchCode string
	db.Raw("SELECT code FROM branches WHERE id = ?", branchID).Scan(&branchCode)
	if branchCode == "" {
		return "—"
	}
	var seq uint64
	db.Raw("SELECT COALESCE(last_seq, 0) FROM code_sequences WHERE branch_id = ? AND entity_type = ?", branchID, entity).Scan(&seq)
	return fmt.Sprintf("%s-%03d", branchCode, seq+1)
}

// GenerateTxCodeForBranch generates transaction code using branch prefix.
// Format: {BRANCHCODE}-{PREFIX}-{001}  e.g. CRNS-DEP-001
// Falls back to global sequential if no branch.
func GenerateTxCodeForBranch(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return GenerateCode(db, 0, entity)
	}
	var branchCode string
	if err := db.Raw("SELECT code FROM branches WHERE id = ?", branchID).Scan(&branchCode).Error; err != nil || branchCode == "" {
		return GenerateCode(db, 0, entity)
	}
	if err := db.Exec(`
		INSERT INTO code_sequences (branch_id, entity_type, last_seq, updated_at)
		VALUES (?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE last_seq = last_seq + 1, updated_at = NOW()
	`, branchID, entity).Error; err != nil {
		return fallbackCode(entity)
	}
	var seq uint64
	db.Raw("SELECT last_seq FROM code_sequences WHERE branch_id = ? AND entity_type = ?", branchID, entity).Scan(&seq)
	return fmt.Sprintf("%s-%03d", branchCode, seq)
}

// PeekTxCode previews next transaction code for a branch without incrementing.
func PeekTxCode(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil || branchID == 0 {
		return fmt.Sprintf("%s-000001", entityPrefix(entity))
	}
	var branchCode string
	db.Raw("SELECT code FROM branches WHERE id = ?", branchID).Scan(&branchCode)
	if branchCode == "" {
		return fmt.Sprintf("%s-000001", entityPrefix(entity))
	}
	var seq uint64
	db.Raw("SELECT COALESCE(last_seq, 0) FROM code_sequences WHERE branch_id = ? AND entity_type = ?", branchID, entity).Scan(&seq)
	return fmt.Sprintf("%s-%03d", branchCode, seq+1)
}

// GeneratePrefixCodeForBranch generates a code using entity prefix + sequence.
// Format: {PREFIX}-{001}  e.g. DEP-001, WDR-001
// Uses branch_id for the counter (each branch has its own counter per entity).
func GeneratePrefixCodeForBranch(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil {
		return fallbackCode(entity)
	}
	branchKey := branchID
	if branchKey == 0 {
		branchKey = 1
	} // fallback key when no branch
	if err := db.Exec(`
		INSERT INTO code_sequences (branch_id, entity_type, last_seq, updated_at)
		VALUES (?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE last_seq = last_seq + 1, updated_at = NOW()
	`, branchKey, entity).Error; err != nil {
		return fallbackCode(entity)
	}
	var seq uint64
	db.Raw("SELECT last_seq FROM code_sequences WHERE branch_id = ? AND entity_type = ?", branchKey, entity).Scan(&seq)
	return fmt.Sprintf("%s-%03d", entityPrefix(entity), seq)
}

// PeekPrefixCode previews the next PREFIX-001 style code for a branch without incrementing.
func PeekPrefixCode(db *gorm.DB, branchID uint, entity EntityType) string {
	if db == nil {
		return fmt.Sprintf("%s-001", entityPrefix(entity))
	}
	branchKey := branchID
	if branchKey == 0 {
		branchKey = 1
	}
	var seq uint64
	db.Raw("SELECT COALESCE(last_seq, 0) FROM code_sequences WHERE branch_id = ? AND entity_type = ?", branchKey, entity).Scan(&seq)
	return fmt.Sprintf("%s-%03d", entityPrefix(entity), seq+1)
}
