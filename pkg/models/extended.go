package models

import "time"

type Report struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type"`
	TimeRange string    `json:"time_range"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	FilePath  string    `json:"file_path"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationConfig struct {
	ID               uint   `json:"id" gorm:"primaryKey"`
	EmailEnabled     bool   `json:"email_enabled"`
	EmailSMTP        string `json:"email_smtp"`
	EmailUsername    string `json:"email_username"`
	EmailPassword    string `json:"email_password"`
	EmailRecipients  string `json:"email_recipients"`
	WebhookEnabled   bool   `json:"webhook_enabled"`
	WebhookURL       string `json:"webhook_url"`
	DingTalkEnabled  bool   `json:"dingtalk_enabled"`
	DingTalkWebhook  string `json:"dingtalk_webhook"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PCAPSession struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SessionID    string    `json:"session_id" gorm:"uniqueIndex"`
	SrcIP        string    `json:"src_ip" gorm:"index"`
	DstIP        string    `json:"dst_ip" gorm:"index"`
	SrcPort      int       `json:"src_port"`
	DstPort      int       `json:"dst_port"`
	Protocol     string    `json:"protocol"`
	StartTime    time.Time `json:"start_time" gorm:"index"`
	EndTime      time.Time `json:"end_time"`
	PacketCount  int       `json:"packet_count"`
	BytesTotal   int64     `json:"bytes_total"`
	FilePath     string    `json:"file_path"`
	CreatedAt    time.Time `json:"created_at"`
}
