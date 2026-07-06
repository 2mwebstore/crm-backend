package repositories

import (
	"crm-backend/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error

	// Sub-user tree queries
	ListAll() ([]models.User, error)                      // all users (super admin only)
	ListChildren(parentID uint) ([]models.User, error)    // direct children only
	ListDescendants(parentID uint) ([]models.User, error) // all descendants (recursive)
	GetDescendantIDs(userID uint) ([]uint, error)         // all descendant IDs + self
	GetRootAncestorID(userID uint) (uint, error)          // walk up to root
	GetScopeIDs(userID uint) ([]uint, error)
	GetUserBranchIDs(userID uint) ([]uint, error)
	GetLookupScope(userID uint) ([]uint, error)
	GetUsersInScope(userID uint) ([]models.User, error)
	AssignBranches(userID uint, branchIDs []uint) error
	CountChildren(parentID uint) (int64, error)
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Role.Permissions").Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Role.Permissions").First(&u, id).Error
	return &u, err
}

func (r *userRepository) Update(u *models.User) error {
	// Clear preloaded associations before Save to avoid GORM trying to update them.
	// We only want to update the FK columns (role_id, etc.), not the nested structs.
	u.Role = nil
	u.Branches = nil
	return r.db.Save(u).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) ListAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Role").Preload("Branches").
		Order("parent_id IS NOT NULL, parent_id ASC, created_at ASC").
		Find(&users).Error
	return users, err
}

func (r *userRepository) ListDescendants(parentID uint) ([]models.User, error) {
	type row struct{ ID uint }
	var rows []row
	err := r.db.Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id FROM users WHERE parent_id = ?
			UNION ALL
			SELECT u.id FROM users u
			INNER JOIN subtree s ON u.parent_id = s.id
		)
		SELECT id FROM subtree
	`, parentID).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return []models.User{}, err
	}
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	var users []models.User
	err = r.db.Preload("Role").
		Where("id IN ?", ids).
		Order("parent_id ASC, created_at ASC").
		Find(&users).Error
	return users, err
}

func (r *userRepository) ListChildren(parentID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Role").
		Where("parent_id = ?", parentID).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// GetDescendantIDs returns the given userID plus all descendant IDs using
// a recursive CTE (MySQL 8+). This is the core of the ownership scope system.
func (r *userRepository) GetDescendantIDs(userID uint) ([]uint, error) {
	// Recursive CTE: walk the tree downward from userID
	type row struct{ ID uint }
	var rows []row
	err := r.db.Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id FROM users WHERE id = ?
			UNION ALL
			SELECT u.id FROM users u
			INNER JOIN subtree s ON u.parent_id = s.id
		)
		SELECT id FROM subtree
	`, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

// GetRootAncestorID walks up the parent chain to find the root user.
func (r *userRepository) GetRootAncestorID(userID uint) (uint, error) {
	type row struct{ ID uint }
	var result row
	err := r.db.Raw(`
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id FROM users WHERE id = ?
			UNION ALL
			SELECT u.id, u.parent_id FROM users u
			INNER JOIN ancestors a ON u.id = a.parent_id
		)
		SELECT id FROM ancestors WHERE parent_id IS NULL LIMIT 1
	`, userID).Scan(&result).Error
	if err != nil {
		return 0, err
	}
	if result.ID == 0 {
		return userID, nil // fallback: return self
	}
	return result.ID, nil
}

// GetScopeIDs returns the data scope for a user based on branch assignment.
//
// New branch-based scope logic:
//  1. Walk UP to the root ancestor (Simple User / direct user of Super Admin).
//  2. Get branches assigned to that root user via user_branches table.
//  3. If branches exist:
//     a. Find ALL root-level users who share any of those same branches.
//     b. For each such user, collect them + all their descendants.
//     c. Return the combined set — this is the full branch scope.
//  4. Fallback (no branches assigned): return old subtree scope (root + descendants).
//
// Sub-user behavior:
//
//	Sub User A1 has no branches → walks up to Simple User A → inherits A's branches.
//	Sub User A1 sees exactly the same data as Simple User A.
//
// Branch sharing:
//
//	If Simple User A and Simple User B are both assigned to branch PHNM,
//	they (and all their sub-users) see each other's data within that branch.
//
// GetScopeIDs returns the data scope for a user based on their type and branch assignment.
//
// Rules:
//  1. Super Admin           → nil (no filter, sees ALL data)
//  2. Sub-user of SA        → nil (inherits SA scope, sees ALL data)
//  3. Simple User (branch)  → all users sharing the same branch(es)
//  4. Sub-user (branch)     → inherits parent's branch scope
//  5. No branch assigned    → own parent-group only (root + descendants)
func (r *userRepository) GetScopeIDs(userID uint) ([]uint, error) {
	// Step 1: walk up to root ancestor
	rootID, err := r.GetRootAncestorID(userID)
	if err != nil {
		return nil, err
	}

	// Step 2: check if the root ancestor is Super Admin
	// If so — this user (or their sub-user) inherits Super Admin's all-access scope
	type saRow struct{ IsSuperAdmin bool }
	var sa saRow
	r.db.Raw(`SELECT is_super_admin FROM users WHERE id = ?`, rootID).Scan(&sa)
	if sa.IsSuperAdmin {
		// Return nil = no filter = sees all data (same as Super Admin)
		return nil, nil
	}

	// Step 3: get branches assigned to root user
	var branchIDs []uint
	r.db.Raw(`SELECT branch_id FROM user_branches WHERE user_id = ?`, rootID).Scan(&branchIDs)

	if len(branchIDs) == 0 {
		// No branch assigned — no data access
		return []uint{}, nil
	}

	// Step 4: find all root users (parent_id IS NULL) sharing any of these branches
	var rootUserIDs []uint
	r.db.Raw(`
		SELECT DISTINCT u.id
		FROM users u
		INNER JOIN user_branches ub ON ub.user_id = u.id
		WHERE ub.branch_id IN ? AND u.parent_id IS NULL
	`, branchIDs).Scan(&rootUserIDs)

	if len(rootUserIDs) == 0 {
		return r.GetDescendantIDs(rootID)
	}

	// Step 5: collect all descendants of each branch-sharing root user (deduped)
	seen := make(map[uint]bool)
	var allIDs []uint
	for _, rid := range rootUserIDs {
		desc, _ := r.GetDescendantIDs(rid)
		for _, id := range desc {
			if !seen[id] {
				seen[id] = true
				allIDs = append(allIDs, id)
			}
		}
	}

	if len(allIDs) == 0 {
		return r.GetDescendantIDs(rootID)
	}
	return allIDs, nil
}

func (r *userRepository) CountChildren(parentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// AssignBranches replaces all branches for a user.
func (r *userRepository) AssignBranches(userID uint, branchIDs []uint) error {
	user := &models.User{}
	user.ID = userID
	branches := make([]models.Branch, len(branchIDs))
	for i, id := range branchIDs {
		branches[i].ID = id
	}
	return r.db.Model(user).Association("Branches").Replace(branches)
}

// GetUserBranchIDs returns the branch IDs accessible to a user.
//   - Super Admin / SA sub-user → nil (no filter = all data)
//   - Simple User with branches → their branch IDs
//   - No branches assigned     → empty slice (fall back to created_by scope)
func (r *userRepository) GetUserBranchIDs(userID uint) ([]uint, error) {
	rootID, err := r.GetRootAncestorID(userID)
	if err != nil {
		return nil, err
	}

	var sa struct{ IsSuperAdmin bool }
	r.db.Raw(`SELECT is_super_admin FROM users WHERE id = ?`, rootID).Scan(&sa)
	if sa.IsSuperAdmin {
		return nil, nil
	}

	var branchIDs []uint
	r.db.Raw(`SELECT branch_id FROM user_branches WHERE user_id = ?`, rootID).Scan(&branchIDs)
	return branchIDs, nil
}

// GetScopeIDsByBranch returns user IDs whose records should be visible to the caller,
// using branch-assignment logic (mirrors branch ListForUser):
//   - SA / SA sub-user          → nil  (no filter = all)
//   - Simple User with branches → IDs of all users assigned to the same branches
//   - Simple User no branches   → just the root user themselves (own records only)
//   - Sub-user of Simple User   → same as root's assigned branches
func (r *userRepository) GetScopeIDsByBranch(userID uint) ([]uint, error) {
	// Walk to root ancestor
	rootID := userID
	for {
		var row struct {
			ParentID     *uint
			IsSuperAdmin bool
		}
		if row.IsSuperAdmin {
			return nil, nil // SA → no filter
		}
		if row.ParentID == nil {
			break
		}
		rootID = *row.ParentID
	}

	// Get branches assigned to this root user
	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", rootID).Scan(&branchIDs)

	if len(branchIDs) == 0 {
		// No branches → only own records
		return []uint{rootID}, nil
	}

	// Find all users (root-level) assigned to these same branches
	var userIDs []uint
	r.db.Raw(`
		SELECT DISTINCT ub.user_id
		FROM user_branches ub
		WHERE ub.branch_id IN ?
	`, branchIDs).Scan(&userIDs)

	// Expand each root user to include all their descendants
	allIDs := make([]uint, 0)
	seen := make(map[uint]bool)
	for _, uid := range userIDs {
		desc, err := r.GetDescendantIDs(uid)
		if err != nil {
			continue
		}
		for _, d := range desc {
			if !seen[d] {
				seen[d] = true
				allIDs = append(allIDs, d)
			}
		}
	}
	if len(allIDs) == 0 {
		return []uint{rootID}, nil
	}
	return allIDs, nil
}

// GetLookupScope returns scope for lookup items (levels, bank types, etc):
//   - SA / SA sub-user          → nil  (no filter = all)
//   - Simple User with branches → nil  (no filter = all — lookups are shared)
//   - Simple User no branches   → []uint{} (empty = no access)
//   - Sub-user                  → same as root
func (r *userRepository) GetLookupScope(userID uint) ([]uint, error) {
	// Walk to root ancestor
	rootID := userID
	for {
		var row struct {
			ParentID     *uint
			IsSuperAdmin bool
		}
		if err := r.db.Raw("SELECT parent_id, is_super_admin FROM users WHERE id = ?", rootID).Scan(&row).Error; err != nil {
			break
		}
		if row.IsSuperAdmin {
			return nil, nil // SA → no filter
		}
		if row.ParentID == nil {
			break
		}
		rootID = *row.ParentID
	}

	// Check if root has branches assigned
	var count int64
	r.db.Raw("SELECT COUNT(*) FROM user_branches WHERE user_id = ?", rootID).Scan(&count)
	if count == 0 {
		return []uint{0}, nil // no branches → empty result (0 never matches created_by_id)
	}

	// Has branches → full access to all lookups
	return nil, nil
}

// GetUsersInScope returns all users visible to the caller for filter dropdowns.
//   - SA → all users
//   - User with branches → all users assigned to same branches
//   - User with no branches → just themselves
func (r *userRepository) GetUsersInScope(userID uint) ([]models.User, error) {
	var isSA bool
	r.db.Raw("SELECT is_super_admin FROM users WHERE id = ?", userID).Scan(&isSA)
	if isSA {
		var users []models.User
		err := r.db.Select("id, name, email").Find(&users).Error
		return users, err
	}
	// Get caller's branches
	var branchIDs []uint
	r.db.Raw("SELECT branch_id FROM user_branches WHERE user_id = ?", userID).Scan(&branchIDs)
	if len(branchIDs) == 0 {
		var users []models.User
		err := r.db.Select("id, name, email").Where("id = ?", userID).Find(&users).Error
		return users, err
	}
	// Get all users in those branches
	var userIDs []uint
	r.db.Raw("SELECT DISTINCT user_id FROM user_branches WHERE branch_id IN ?", branchIDs).Scan(&userIDs)
	var users []models.User
	err := r.db.Select("id, name, email").Where("id IN ?", userIDs).Find(&users).Error
	return users, err
}
