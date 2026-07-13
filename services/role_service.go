package services

import (
	"errors"

	"gorm.io/gorm"

	lookupdto "crm-backend/dto/lookup"
	"crm-backend/models"
	"crm-backend/repositories"
)

type RoleService interface {
	Create(createdByID uint, req lookupdto.CreateRoleRequest) (*models.Role, error)
	ListAccessible(callerID uint, nameFilter string, createdByID *uint) ([]models.Role, error)
	ListForAssignment(callerID uint) ([]models.Role, error)
	GetByID(id uint, callerSubtree []uint) (*models.Role, error)
	Update(id uint, callerID uint, callerSubtree []uint, req lookupdto.UpdateRoleRequest) (*models.Role, error)
	Delete(id uint, callerSubtree []uint) error
	AssignPermissions(roleID uint, permIDs []uint, callerSubtree []uint) (*models.Role, error)
}

type roleService struct {
	roleRepo repositories.RoleRepository
	permRepo repositories.PermissionRepository
	userRepo repositories.UserRepository
}

func NewRoleService(rr repositories.RoleRepository, pr repositories.PermissionRepository, ur repositories.UserRepository) RoleService {
	return &roleService{rr, pr, ur}
}

func (s *roleService) Create(createdByID uint, req lookupdto.CreateRoleRequest) (*models.Role, error) {
	// Prevent duplicate name within creator's scope
	if _, err := s.roleRepo.FindByName(req.Name, &createdByID); err == nil {
		return nil, errors.New("role name already exists in your scope")
	}

	var perms []models.Permission
	if len(req.PermissionIDs) > 0 {
		var err error
		perms, err = s.permRepo.FindByIDs(req.PermissionIDs)
		if err != nil {
			return nil, err
		}
		// Validate caller has all requested permissions (no privilege escalation)
		caller, err := s.userRepo.FindByID(createdByID)
		if err != nil {
			return nil, err
		}
		if !caller.IsSuperAdmin {
			for _, p := range perms {
				if !caller.HasPermission(p.Name) {
					return nil, errors.New("cannot assign permission '" + p.Name + "' — you do not have it yourself")
				}
			}
		}
	}

	// Super Admin created roles are system roles (is_system = true)
	caller2, _ := s.userRepo.FindByID(createdByID)
	isSystem := caller2 != nil && caller2.IsSuperAdmin
	var createdByRef *uint
	if !isSystem {
		createdByRef = &createdByID
	}
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    isSystem,
		CreatedByID: createdByRef,
		Permissions: perms,
	}
	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}
	return s.roleRepo.FindByID(role.ID)
}

// ListAccessible returns roles visible to the caller, optionally narrowed by
// role name (partial match) and by the user who created the role.
//
// nameFilter and createdByID are applied on top of the existing scope rules:
//   - SA/SA sub-user (scope == nil): only system roles, optionally filtered
//     by name and/or a specific creator.
//   - Simple/Sub User (scope != nil): only their own created roles. Passing a
//     createdByID other than callerID here will simply return no results,
//     since the repo already scopes to created_by_id = callerID.
func (s *roleService) ListAccessible(callerID uint, nameFilter string, createdByID *uint) ([]models.Role, error) {
	scope, err := s.userRepo.GetScopeIDs(callerID)
	if err != nil {
		return nil, err
	}
	if scope == nil {
		// SA or SA sub-user: only system roles (is_system = 1)
		return s.roleRepo.ListAccessible(0, nameFilter, createdByID)
	}
	// Simple User / Sub User: ONLY roles they created (created_by_id = callerID)
	// No system roles shown in the list — they assign system roles to sub-users separately
	return s.roleRepo.ListAccessible(callerID, nameFilter, createdByID)
}

func (s *roleService) GetByID(id uint, callerSubtree []uint) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	// nil subtree = SA/SA sub-user → can see all roles
	if !role.IsSystem && callerSubtree != nil && !inSlice(role.CreatedByID, callerSubtree) {
		return nil, errors.New("role not found")
	}
	return role, nil
}

func (s *roleService) Update(id uint, callerID uint, callerSubtree []uint, req lookupdto.UpdateRoleRequest) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}
	// callerSubtree == nil means SA/SA-sub: can edit any role (including is_system=1)
	// Simple User: can only edit their own created roles (not system roles)
	if callerSubtree != nil {
		if role.IsSystem {
			return nil, errors.New("you cannot modify system roles")
		}
		if !inSlice(role.CreatedByID, callerSubtree) {
			return nil, errors.New("role not found")
		}
	}
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}
	return s.roleRepo.FindByID(id)
}

func (s *roleService) Delete(id uint, callerSubtree []uint) error {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return errors.New("role not found")
	}
	if callerSubtree != nil {
		if role.IsSystem {
			return errors.New("you cannot delete system roles")
		}
		if !inSlice(role.CreatedByID, callerSubtree) {
			return errors.New("role not found")
		}
	}
	return s.roleRepo.Delete(id)
}

func (s *roleService) AssignPermissions(roleID uint, permIDs []uint, callerSubtree []uint) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return nil, errors.New("role not found")
	}
	if callerSubtree != nil {
		if role.IsSystem {
			return nil, errors.New("you cannot modify system role permissions")
		}
		if !inSlice(role.CreatedByID, callerSubtree) {
			return nil, errors.New("role not found")
		}
	}
	perms, err := s.permRepo.FindByIDs(permIDs)
	if err != nil {
		return nil, err
	}
	if err := s.roleRepo.SetPermissions(role, perms); err != nil {
		return nil, err
	}
	return s.roleRepo.FindByID(roleID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func inSlice(idPtr *uint, ids []uint) bool {
	if idPtr == nil {
		return false
	}
	for _, id := range ids {
		if id == *idPtr {
			return true
		}
	}
	return false
}

// ListForAssignment returns roles available for assigning to users:
// - SA/SA sub-users: all system roles
// - Simple/Sub users: system roles + own created roles
func (s *roleService) ListForAssignment(callerID uint) ([]models.Role, error) {
	scope, err := s.userRepo.GetScopeIDs(callerID)
	if err != nil {
		return nil, err
	}
	if scope == nil {
		// SA scope: only system roles for assignment
		return s.roleRepo.ListForAssignment(0)
	}
	// Regular user: system roles + their own roles
	return s.roleRepo.ListForAssignment(callerID)
}
