package services

import (
	"errors"

	"gorm.io/gorm"

	userdto "crm-backend/dto/user"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

// UserService manages three tiers of users:
//
//	Super Admin  (is_super_admin=true)
//	  └─ can list/create/edit/delete ALL users across the system
//	  └─ cannot be deleted or modified by anyone
//
//	Simple User  (parent_id=nil, not super admin)  ← registered via /auth/register
//	  └─ owner of their own data scope
//	  └─ can create Sub Users under themselves
//	  └─ sees own data + all descendants' data
//
//	Sub User     (parent_id=X)
//	  └─ created by a Simple User or another Sub User
//	  └─ sees own data + all their own descendants' data
//	  └─ cannot see sibling sub-users' data
type UserService interface {
	// Called by Super Admin to manage all users
	ListAllUsers() ([]models.User, error)
	AdminCreateUser(req userdto.AdminCreateUserRequest) (*models.User, error)
	AdminUpdateUser(targetID uint, req userdto.AdminUpdateUserRequest) (*models.User, error)
	AdminDeleteUser(targetID uint) error

	// Called by Simple/Sub users to manage their own sub-users
	CreateSubUser(callerID uint, req userdto.CreateSubUserRequest) (*models.User, error)
	ListSubUsers(callerID uint) ([]models.User, error)
	GetSubUser(callerID uint, targetID uint) (*models.User, error)
	UpdateSubUser(callerID uint, targetID uint, req userdto.UpdateSubUserRequest) (*models.User, error)
	DeleteSubUser(callerID uint, targetID uint) error

	// Helpers used by other services
	GetDescendantIDs(userID uint) ([]uint, error)
	GetScopeIDs(userID uint) ([]uint, error)
	GetUserBranchIDs(userID uint) ([]uint, error)
	GetLookupScope(userID uint) ([]uint, error)
	GetUsersInScope(userID uint) ([]models.User, error)
	GetUserRepo() repositories.UserRepository
	GetByID(id uint) (*models.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
	roleRepo repositories.RoleRepository
}

func NewUserService(ur repositories.UserRepository, rr repositories.RoleRepository) UserService {
	return &userService{ur, rr}
}

// ── Super Admin operations ────────────────────────────────────────────────────

func (s *userService) ListAllUsers() ([]models.User, error) {
	return s.userRepo.ListAll()
}

func (s *userService) AdminCreateUser(req userdto.AdminCreateUserRequest) (*models.User, error) {
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var roleID *uint
	if req.RoleID != nil {
		if _, err := s.roleRepo.FindByID(*req.RoleID); err != nil {
			return nil, errors.New("role not found")
		}
		roleID = req.RoleID
	}

	user := &models.User{
		Name: req.Name, Email: req.Email, PasswordHash: hash,
		RoleID: roleID, IsActive: true,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	_ = s.userRepo.AssignBranches(user.ID, req.BranchIDs) // empty = clear all branches
	return s.userRepo.FindByID(user.ID)
}

func (s *userService) AdminUpdateUser(targetID uint, req userdto.AdminUpdateUserRequest) (*models.User, error) {
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if target.IsSuperAdmin {
		return nil, errors.New("the super admin account cannot be modified")
	}
	// Block role assignment on users who are Super Admin type
	if req.RoleID != nil && target.IsSuperAdmin {
		return nil, errors.New("cannot change role of a super admin user")
	}

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}
	if req.RoleID != nil {
		if *req.RoleID == 0 {
			target.RoleID = nil
		} else {
			if _, err := s.roleRepo.FindByID(*req.RoleID); err != nil {
				return nil, errors.New("role not found")
			}
			target.RoleID = req.RoleID
		}
	}

	if req.Password != nil && *req.Password != "" {
		if len(*req.Password) < 6 {
			return nil, errors.New("password must be at least 6 characters")
		}
		hash, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		target.PasswordHash = hash
	}

	if err := s.userRepo.Update(target); err != nil {
		return nil, err
	}

	// Update branch assignments if provided
	if req.BranchIDs != nil {
		_ = s.userRepo.AssignBranches(targetID, req.BranchIDs)
	}

	return s.userRepo.FindByID(targetID)
}

func (s *userService) AdminDeleteUser(targetID uint) error {
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return errors.New("user not found")
	}
	if target.IsSuperAdmin {
		return errors.New("the super admin account cannot be deleted")
	}
	return s.userRepo.Delete(targetID)
}

// ── Simple / Sub user operations ──────────────────────────────────────────────

func (s *userService) CreateSubUser(callerID uint, req userdto.CreateSubUserRequest) (*models.User, error) {
	caller, err := s.userRepo.FindByID(callerID)
	if err != nil {
		return nil, errors.New("caller not found")
	}
	if !caller.HasPermission(models.PermUserCreate) {
		return nil, errors.New("you do not have permission to create sub-users")
	}
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	var roleID *uint
	if req.RoleID != nil {
		role, err := s.roleRepo.FindByID(*req.RoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("role not found")
			}
			return nil, err
		}
		// Anti-escalation: all role permissions must be subset of caller's
		if !caller.IsSuperAdmin {
			for _, p := range role.Permissions {
				if !caller.HasPermission(p.Name) {
					return nil, errors.New("cannot assign permission '" + p.DisplayName + "' — you do not have it yourself")
				}
			}
		}
		roleID = req.RoleID
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name: req.Name, Email: req.Email, PasswordHash: hash,
		RoleID: roleID, IsActive: true,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	// Assign branch if provided
	if len(req.BranchIDs) > 0 {
		_ = s.userRepo.AssignBranches(user.ID, req.BranchIDs)
	}
	return s.userRepo.FindByID(user.ID)
}

func (s *userService) ListSubUsers(callerID uint) ([]models.User, error) {
	caller, err := s.userRepo.FindByID(callerID)
	if err != nil {
		return nil, err
	}
	if !caller.HasPermission(models.PermUserView) {
		return nil, errors.New("you do not have permission to view sub-users")
	}
	return s.userRepo.ListDescendants(callerID)
}

func (s *userService) GetSubUser(callerID uint, targetID uint) (*models.User, error) {
	subtree, err := s.userRepo.GetDescendantIDs(callerID)
	if err != nil {
		return nil, err
	}
	if !containsUint(subtree, targetID) || targetID == callerID {
		return nil, errors.New("user not found")
	}
	return s.userRepo.FindByID(targetID)
}

func (s *userService) UpdateSubUser(callerID uint, targetID uint, req userdto.UpdateSubUserRequest) (*models.User, error) {
	caller, err := s.userRepo.FindByID(callerID)
	if err != nil {
		return nil, err
	}
	if !caller.HasPermission(models.PermUserEdit) {
		return nil, errors.New("you do not have permission to edit sub-users")
	}

	subtree, err := s.userRepo.GetDescendantIDs(callerID)
	if err != nil {
		return nil, err
	}
	if !containsUint(subtree, targetID) || targetID == callerID {
		return nil, errors.New("user not found")
	}

	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if target.IsSuperAdmin {
		return nil, errors.New("the super admin account cannot be modified")
	}
	// Block role assignment on users who are Super Admin type
	if req.RoleID != nil && target.IsSuperAdmin {
		return nil, errors.New("cannot change role of a super admin user")
	}

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	if req.RoleID != nil {
		role, err := s.roleRepo.FindByID(*req.RoleID)
		if err != nil {
			return nil, errors.New("role not found")
		}
		if !caller.IsSuperAdmin {
			for _, p := range role.Permissions {
				if !caller.HasPermission(p.Name) {
					return nil, errors.New("cannot assign permission '" + p.DisplayName + "' — you do not have it yourself")
				}
			}
		}
		target.RoleID = req.RoleID
	}
	if req.Password != nil {
		hash, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		target.PasswordHash = hash
	}

	if err := s.userRepo.Update(target); err != nil {
		return nil, err
	}
	// Update branch if provided
	if req.BranchIDs != nil {
		_ = s.userRepo.AssignBranches(targetID, req.BranchIDs)
	}
	return s.userRepo.FindByID(targetID)
}
func (s *userService) DeleteSubUser(callerID uint, targetID uint) error {
	caller, err := s.userRepo.FindByID(callerID)
	if err != nil {
		return err
	}
	if !caller.HasPermission(models.PermUserDelete) {
		return errors.New("you do not have permission to delete sub-users")
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return errors.New("user not found")
	}
	if target.IsSuperAdmin {
		return errors.New("the super admin account cannot be deleted")
	}

	subtree, err := s.userRepo.GetDescendantIDs(callerID)
	if err != nil {
		return err
	}
	if !containsUint(subtree, targetID) || targetID == callerID {
		return errors.New("user not found")
	}
	return s.userRepo.Delete(targetID)
}

// GetDescendantIDs kept for backward compat — delegates to branch-aware GetScopeIDs.
func (s *userService) GetDescendantIDs(userID uint) ([]uint, error) {
	return s.userRepo.GetScopeIDs(userID)
}

// GetScopeIDs returns the branch-aware data scope for a user.
func (s *userService) GetScopeIDs(userID uint) ([]uint, error) {
	return s.userRepo.GetScopeIDs(userID)
}

func (s *userService) GetByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func containsUint(slice []uint, val uint) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func (s *userService) GetUserRepo() repositories.UserRepository {
	return s.userRepo
}

func (s *userService) GetUserBranchIDs(userID uint) ([]uint, error) {
	return s.userRepo.GetUserBranchIDs(userID)
}

func (s *userService) GetLookupScope(userID uint) ([]uint, error) {
	return s.userRepo.GetLookupScope(userID)
}

func (s *userService) GetUsersInScope(userID uint) ([]models.User, error) {
	return s.userRepo.GetUsersInScope(userID)
}
