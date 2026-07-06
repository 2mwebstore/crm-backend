package services

import (
	"errors"

	"gorm.io/gorm"

	"crm-backend/config"
	authdto "crm-backend/dto/auth"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/utils"
)

type AuthService interface {
	Register(req authdto.RegisterRequest) (*models.User, error)
	Login(req authdto.LoginRequest) (string, *models.User, error)
	ChangePassword(userID uint, req authdto.ChangePasswordRequest) error
	UpdateProfile(userID uint, req authdto.UpdateProfileRequest) (*models.User, error)
	GetProfile(userID uint) (*models.User, error)
}

type authService struct {
	userRepo repositories.UserRepository
	roleRepo repositories.RoleRepository
}

func NewAuthService(userRepo repositories.UserRepository, roleRepo repositories.RoleRepository) AuthService {
	return &authService{userRepo, roleRepo}
}

func (s *authService) Register(req authdto.RegisterRequest) (*models.User, error) {
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// New root users get the system "Owner" role
	var ownerRoleID *uint
	ownerRole, err := s.roleRepo.FindByName("Owner", nil)
	if err == nil {
		ownerRoleID = &ownerRole.ID
	}

	user := &models.User{
		Name: req.Name, Email: req.Email, PasswordHash: hash, RoleID: ownerRoleID,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(user.ID)
}

func (s *authService) Login(req authdto.LoginRequest) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid email or password")
		}
		return "", nil, err
	}
	if !user.IsActive {
		return "", nil, errors.New("account is disabled")
	}
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return "", nil, errors.New("invalid email or password")
	}

	cfg := config.Get()
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}
	token, err := utils.GenerateToken(user.ID, user.Email, roleName, cfg.JWT.Secret, cfg.JWT.ExpireHours, user.IsSuperAdmin)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *authService) ChangePassword(userID uint, req authdto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}
	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.userRepo.Update(user)
}

func (s *authService) UpdateProfile(userID uint, req authdto.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) GetProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}
