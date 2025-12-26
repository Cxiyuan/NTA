package service

import (
	"context"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/services/alert-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type AlertService struct {
	repo   *repository.AlertRepository
	logger *logrus.Logger
}

func NewAlertService(repo *repository.AlertRepository, logger *logrus.Logger) *AlertService {
	return &AlertService{repo: repo, logger: logger}
}

func (s *AlertService) GetAlertByID(ctx context.Context, id uint) (*models.Alert, error) {
	return s.repo.GetAlertByID(ctx, id)
}

func (s *AlertService) ListAlerts(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]models.Alert, int64, error) {
	return s.repo.ListAlerts(ctx, limit, offset, filters)
}

func (s *AlertService) CreateAlert(ctx context.Context, alert *models.Alert) error {
	return s.repo.CreateAlert(ctx, alert)
}

func (s *AlertService) UpdateAlert(ctx context.Context, id uint, updates map[string]interface{}) error {
	return s.repo.UpdateAlert(ctx, id, updates)
}

func (s *AlertService) DeleteAlert(ctx context.Context, id uint) error {
	return s.repo.DeleteAlert(ctx, id)
}
