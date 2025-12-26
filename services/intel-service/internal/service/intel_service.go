package service

import (
	"github.com/Cxiyuan/NTA/services/intel-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type IntelService struct {
	repo   *repository.IntelRepository
	logger *logrus.Logger
}

func NewIntelService(repo *repository.IntelRepository, logger *logrus.Logger) *IntelService {
	return &IntelService{
		repo:   repo,
		logger: logger,
	}
}
