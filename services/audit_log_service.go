package services

import (
	"errors"
	"time"

	"crm-backend/models"
	"crm-backend/repositories"
)

type AuditLogService interface {
	List(filter repositories.AuditLogFilter, page, pageSize int) ([]models.AuditLog, int64, error)
	// DeleteOlderThan removes every entry older than the given period —
	// "week" (7 days) or "month" (1 calendar month) — and returns how
	// many rows were deleted.
	DeleteOlderThan(period string) (int64, error)
}

type auditLogService struct {
	repo repositories.AuditLogRepository
}

func NewAuditLogService(repo repositories.AuditLogRepository) AuditLogService {
	return &auditLogService{repo}
}

func (s *auditLogService) List(filter repositories.AuditLogFilter, page, pageSize int) ([]models.AuditLog, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

func (s *auditLogService) DeleteOlderThan(period string) (int64, error) {
	now := time.Now()
	var cutoff time.Time
	switch period {
	case "week":
		cutoff = now.AddDate(0, 0, -7)
	case "month":
		cutoff = now.AddDate(0, -1, 0)
	default:
		return 0, errors.New(`period must be "week" or "month"`)
	}
	return s.repo.DeleteBefore(cutoff)
}
