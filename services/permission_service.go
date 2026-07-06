package services

import (
	"crm-backend/models"
	"crm-backend/repositories"
)

type PermissionService interface {
	ListAll() ([]models.Permission, error)
	ListGrouped() (map[string][]models.Permission, error)
}

type permissionService struct {
	repo repositories.PermissionRepository
}

func NewPermissionService(repo repositories.PermissionRepository) PermissionService {
	return &permissionService{repo}
}

func (s *permissionService) ListAll() ([]models.Permission, error) {
	return s.repo.FindAll()
}

func (s *permissionService) ListGrouped() (map[string][]models.Permission, error) {
	return s.repo.FindGrouped()
}
