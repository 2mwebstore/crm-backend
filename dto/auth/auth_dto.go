package authdto

// LoginRequest is the payload for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest is the payload for POST /auth/register
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"omitempty,oneof=admin manager sales"`
}

// ChangePasswordRequest is the payload for POST /auth/change-password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest is the payload for PUT /auth/profile
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"omitempty,min=2,max=100"`
}
