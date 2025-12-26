package service

import (
	"github.com/Cxiyuan/NTA/services/report-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type ReportService struct {
	repo   *repository.ReportRepository
	logger *logrus.Logger
}

func NewReportService(repo *repository.ReportRepository, logger *logrus.Logger) *ReportService {
	return &ReportService{
		repo:   repo,
		logger: logger,
	}
}
