package service

import (
	"github.com/Cxiyuan/NTA/services/notification-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	repo   *repository.NotificationRepository
	logger *logrus.Logger
}

func NewNotificationService(repo *repository.NotificationRepository, logger *logrus.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}
