package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	logger *logrus.Logger
	config *models.NotificationConfig
}

func NewService(db *gorm.DB, logger *logrus.Logger) *Service {
	s := &Service{
		db:     db,
		logger: logger,
	}
	s.loadConfig()
	return s
}

func (s *Service) loadConfig() {
	var config models.NotificationConfig
	if err := s.db.First(&config).Error; err != nil {
		config = models.NotificationConfig{
			EmailEnabled:   false,
			WebhookEnabled: false,
		}
		s.db.Create(&config)
	}
	s.config = &config
}

func (s *Service) UpdateConfig(config *models.NotificationConfig) error {
	if err := s.db.Model(&models.NotificationConfig{}).Where("id = ?", s.config.ID).Updates(config).Error; err != nil {
		return err
	}
	s.loadConfig()
	return nil
}

func (s *Service) GetConfig() *models.NotificationConfig {
	return s.config
}

func (s *Service) SendAlertNotification(alert *models.Alert) error {
	var errors []string

	if s.config.EmailEnabled {
		if err := s.sendEmail(alert); err != nil {
			s.logger.Errorf("Failed to send email: %v", err)
			errors = append(errors, err.Error())
		}
	}

	if s.config.WebhookEnabled {
		if err := s.sendWebhook(alert); err != nil {
			s.logger.Errorf("Failed to send webhook: %v", err)
			errors = append(errors, err.Error())
		}
	}

	if s.config.DingTalkEnabled {
		if err := s.sendDingTalk(alert); err != nil {
			s.logger.Errorf("Failed to send DingTalk: %v", err)
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (s *Service) sendEmail(alert *models.Alert) error {
	if s.config.EmailSMTP == "" || s.config.EmailUsername == "" {
		return fmt.Errorf("email not configured")
	}

	recipients := strings.Split(s.config.EmailRecipients, ",")
	if len(recipients) == 0 {
		return fmt.Errorf("no email recipients configured")
	}

	subject := fmt.Sprintf("[NTA Alert] %s - %s", alert.Severity, alert.Type)
	body := fmt.Sprintf(`
Alert Details:
--------------
Severity: %s
Type: %s
Source: %s:%d
Destination: %s:%d
Description: %s
Confidence: %.2f
Timestamp: %s

Please login to NTA system for more details.
`, alert.Severity, alert.Type, alert.SrcIP, alert.SrcPort, alert.DstIP, alert.DstPort,
		alert.Description, alert.Confidence, alert.Timestamp.Format("2006-01-02 15:04:05"))

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.config.EmailUsername, strings.Join(recipients, ","), subject, body)

	auth := smtp.PlainAuth("", s.config.EmailUsername, s.config.EmailPassword, strings.Split(s.config.EmailSMTP, ":")[0])

	err := smtp.SendMail(s.config.EmailSMTP, auth, s.config.EmailUsername, recipients, []byte(message))
	if err != nil {
		return err
	}

	s.logger.Infof("Email notification sent for alert %d", alert.ID)
	return nil
}

func (s *Service) sendWebhook(alert *models.Alert) error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	payload := map[string]interface{}{
		"alert_id":    alert.ID,
		"severity":    alert.Severity,
		"type":        alert.Type,
		"src_ip":      alert.SrcIP,
		"dst_ip":      alert.DstIP,
		"description": alert.Description,
		"timestamp":   alert.Timestamp,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	s.logger.Infof("Webhook notification sent for alert %d", alert.ID)
	return nil
}

func (s *Service) sendDingTalk(alert *models.Alert) error {
	if s.config.DingTalkWebhook == "" {
		return fmt.Errorf("DingTalk webhook not configured")
	}

	message := fmt.Sprintf("### NTA 安全告警\n\n"+
		"- **等级**: %s\n"+
		"- **类型**: %s\n"+
		"- **源IP**: %s:%d\n"+
		"- **目标IP**: %s:%d\n"+
		"- **描述**: %s\n"+
		"- **时间**: %s\n",
		alert.Severity, alert.Type, alert.SrcIP, alert.SrcPort,
		alert.DstIP, alert.DstPort, alert.Description,
		alert.Timestamp.Format("2006-01-02 15:04:05"))

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": fmt.Sprintf("[%s] %s", alert.Severity, alert.Type),
			"text":  message,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.config.DingTalkWebhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("DingTalk returned status %d", resp.StatusCode)
	}

	s.logger.Infof("DingTalk notification sent for alert %d", alert.ID)
	return nil
}

func (s *Service) TestNotification(channel string) error {
	testAlert := &models.Alert{
		ID:          0,
		Severity:    "high",
		Type:        "test_alert",
		SrcIP:       "192.168.1.100",
		SrcPort:     12345,
		DstIP:       "192.168.1.200",
		DstPort:     80,
		Description: "This is a test notification from NTA system",
		Confidence:  1.0,
	}

	switch channel {
	case "email":
		return s.sendEmail(testAlert)
	case "webhook":
		return s.sendWebhook(testAlert)
	case "dingtalk":
		return s.sendDingTalk(testAlert)
	default:
		return fmt.Errorf("unknown channel: %s", channel)
	}
}
