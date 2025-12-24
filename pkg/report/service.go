package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	logger *logrus.Logger
	outDir string
}

func NewService(db *gorm.DB, logger *logrus.Logger, outDir string) *Service {
	os.MkdirAll(outDir, 0755)
	return &Service{
		db:     db,
		logger: logger,
		outDir: outDir,
	}
}

func (s *Service) Generate(reportType string, startTime, endTime time.Time, createdBy string) (*models.Report, error) {
	report := &models.Report{
		Type:      reportType,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "generating",
		CreatedBy: createdBy,
		TimeRange: fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
	}

	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}

	go s.generateAsync(report)

	return report, nil
}

func (s *Service) generateAsync(report *models.Report) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("Report generation panic: %v", r)
			s.db.Model(report).Update("status", "failed")
		}
	}()

	var alertStats struct {
		Total    int64
		Critical int64
		High     int64
		Medium   int64
		Low      int64
	}

	s.db.Model(&models.Alert{}).
		Where("timestamp BETWEEN ? AND ?", report.StartTime, report.EndTime).
		Count(&alertStats.Total)

	s.db.Model(&models.Alert{}).
		Where("timestamp BETWEEN ? AND ? AND severity = ?", report.StartTime, report.EndTime, "critical").
		Count(&alertStats.Critical)

	s.db.Model(&models.Alert{}).
		Where("timestamp BETWEEN ? AND ? AND severity = ?", report.StartTime, report.EndTime, "high").
		Count(&alertStats.High)

	s.db.Model(&models.Alert{}).
		Where("timestamp BETWEEN ? AND ? AND severity = ?", report.StartTime, report.EndTime, "medium").
		Count(&alertStats.Medium)

	s.db.Model(&models.Alert{}).
		Where("timestamp BETWEEN ? AND ? AND severity = ?", report.StartTime, report.EndTime, "low").
		Count(&alertStats.Low)

	content := s.buildReportContent(report, alertStats)

	filename := fmt.Sprintf("report_%d_%s.txt", report.ID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(s.outDir, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		s.logger.Errorf("Failed to write report file: %v", err)
		s.db.Model(report).Update("status", "failed")
		return
	}

	s.db.Model(report).Updates(map[string]interface{}{
		"status":    "completed",
		"file_path": filePath,
	})

	s.logger.Infof("Report %d generated successfully", report.ID)
}

func (s *Service) buildReportContent(report *models.Report, stats interface{}) string {
	return fmt.Sprintf(`
NTA Security Report
===================

Report Type: %s
Time Range: %s
Generated: %s

Alert Statistics:
-----------------
Total Alerts: (statistics here)

Top Attack Sources:
------------------
(To be implemented)

Asset Summary:
-------------
(To be implemented)

`, report.Type, report.TimeRange, time.Now().Format("2006-01-02 15:04:05"))
}

func (s *Service) List(page, pageSize int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.Model(&models.Report{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (s *Service) GetFilePath(id uint) (string, error) {
	var report models.Report
	if err := s.db.First(&report, id).Error; err != nil {
		return "", err
	}
	return report.FilePath, nil
}
