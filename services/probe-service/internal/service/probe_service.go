package service

import (
	"github.com/Cxiyuan/NTA/services/probe-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type ProbeService struct {
	repo   *repository.ProbeRepository
	logger *logrus.Logger
}

func NewProbeService(repo *repository.ProbeRepository, logger *logrus.Logger) *ProbeService {
	return &ProbeService{
		repo:   repo,
		logger: logger,
	}
}
