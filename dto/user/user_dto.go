package userdto

// ── Admin DTOs (Super Admin only) ─────────────────────────────────────────────

type AdminCreateUserRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Email     string `json:"email" binding:"required"` // login identifier, not validated as a real email format (e.g. "name@branchcode")
	Password  string `json:"password" binding:"required,min=6"`
	RoleID    *uint  `json:"role_id"`
	ParentID  *uint  `json:"parent_id"`  // nil = root/simple user, set = sub-user
	BranchIDs []uint `json:"branch_ids"` // assigned by super admin
}

type AdminUpdateUserRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=2,max=100"`
	Password  *string `json:"password"`   // validated in service: skip if empty, min=6 if set
	RoleID    *uint   `json:"role_id"`    // 0 = remove role
	ParentID  *uint   `json:"parent_id"`  // 0 = make root user
	BranchIDs []uint  `json:"branch_ids"` // empty = remove all branches
	IsActive  *bool   `json:"is_active"`
}

// ── Simple/Sub user DTOs ──────────────────────────────────────────────────────

type CreateSubUserRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Email     string `json:"email" binding:"required"` // login identifier, not validated as a real email format (e.g. "name@branchcode")
	Password  string `json:"password" binding:"required,min=6"`
	RoleID    *uint  `json:"role_id"`
	BranchIDs []uint `json:"branch_ids"` // up to 1 branch for sub-users
}

type UpdateSubUserRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=2,max=100"`
	Password  *string `json:"password"` // validated in service: skip if empty, min=6 if set
	RoleID    *uint   `json:"role_id"`
	BranchIDs []uint  `json:"branch_ids"` // up to 1 branch; empty = remove
	IsActive  *bool   `json:"is_active"`
}
