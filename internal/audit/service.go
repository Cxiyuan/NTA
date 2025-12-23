package audit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides audit logging
type Service struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewService creates a new audit service
func NewService(db *gorm.DB, logger *logrus.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// Log creates an audit log entry
func (s *Service) Log(user, action, resource string, details map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(details)
	
	entry := &models.AuditLog{
		Timestamp: time.Now(),
		User:      user,
		Action:    action,
		Resource:  resource,
		Details:   string(detailsJSON),
		Result:    "success",
	}

	// Calculate checksum
	data := fmt.Sprintf("%s|%s|%s|%s", user, action, resource, string(detailsJSON))
	hash := sha256.Sum256([]byte(data))
	entry.Checksum = fmt.Sprintf("%x", hash)

	if err := s.db.Create(entry).Error; err != nil {
		s.logger.Errorf("Failed to create audit log: %v", err)
		return err
	}

	s.logger.Infof("Audit: %s %s %s", user, action, resource)
	return nil
}

// Query queries audit logs
func (s *Service) Query(filters map[string]interface{}, limit int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	
	query := s.db.Model(&models.AuditLog{})
	
	for key, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Limit(limit).Order("timestamp DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
