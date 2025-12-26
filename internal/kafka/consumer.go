package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Cxiyuan/NTA/internal/detector"
	"github.com/Cxiyuan/NTA/internal/threatintel"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Consumer struct {
	reader      *kafka.Reader
	db          *gorm.DB
	logger      *logrus.Logger
	detector    *detector.AdvancedDetector
	threatIntel *threatintel.Service
}

func NewConsumer(brokers []string, topic string, groupID string, db *gorm.DB, logger *logrus.Logger, threatIntel *threatintel.Service) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1024,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{
		reader:      reader,
		db:          db,
		logger:      logger,
		detector:    detector.NewAdvancedDetector(logger),
		threatIntel: threatIntel,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Infof("Starting Kafka consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			return c.reader.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				c.logger.Errorf("Failed to fetch message: %v", err)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Errorf("Failed to process message: %v", err)
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Errorf("Failed to commit message: %v", err)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	topic := c.reader.Config().Topic

	switch topic {
	case "zeek-conn":
		return c.processConnLog(msg.Value)
	case "zeek-dns":
		return c.processDNSLog(msg.Value)
	case "zeek-http":
		return c.processHTTPLog(msg.Value)
	case "zeek-ssl":
		return c.processSSLLog(msg.Value)
	case "zeek-notice":
		return c.processNoticeLog(msg.Value)
	default:
		c.logger.Warnf("Unknown topic: %s", topic)
	}

	return nil
}

func (c *Consumer) processConnLog(data []byte) error {
	var conn models.Connection
	if err := json.Unmarshal(data, &conn); err != nil {
		return err
	}

	ctx := context.Background()

	if c.threatIntel != nil {
		if srcIntel, err := c.threatIntel.CheckIP(ctx, conn.SrcIP); err == nil && srcIntel != nil && srcIntel.Severity != "none" {
			alert := &models.Alert{
				Type:         "threat_intel_match",
				Severity:     srcIntel.Severity,
				SrcIP:        conn.SrcIP,
				DstIP:        conn.DstIP,
				Description:  fmt.Sprintf("威胁情报匹配 [%s]: %s - %s", srcIntel.ThreatLabel, conn.SrcIP, srcIntel.Description),
				ThreatLabel:  srcIntel.ThreatLabel,
				ThreatSource: srcIntel.Source,
				Confidence:   0.95,
				Timestamp:    time.Now(),
				Status:       "new",
			}
			c.db.Create(alert)
			c.logger.Warnf("Threat intel match: %s (%s)", conn.SrcIP, srcIntel.ThreatLabel)
		}

		if dstIntel, err := c.threatIntel.CheckIP(ctx, conn.DstIP); err == nil && dstIntel != nil && dstIntel.Severity != "none" {
			alert := &models.Alert{
				Type:         "threat_intel_match",
				Severity:     dstIntel.Severity,
				SrcIP:        conn.SrcIP,
				DstIP:        conn.DstIP,
				Description:  fmt.Sprintf("威胁情报匹配 [%s]: %s - %s", dstIntel.ThreatLabel, conn.DstIP, dstIntel.Description),
				ThreatLabel:  dstIntel.ThreatLabel,
				ThreatSource: dstIntel.Source,
				Confidence:   0.95,
				Timestamp:    time.Now(),
				Status:       "new",
			}
			c.db.Create(alert)
			c.logger.Warnf("Threat intel match: %s (%s)", conn.DstIP, dstIntel.ThreatLabel)
		}
	}

	if isC2, score, c2Type := c.detector.DetectC2Communication(&conn); isC2 {
		alert := &models.Alert{
			Type:        "c2_communication",
			Severity:    "high",
			SrcIP:       conn.SrcIP,
			DstIP:       conn.DstIP,
			Description: "检测到C2通信: " + c2Type,
			Confidence:  score,
			Timestamp:   time.Now(),
			Status:      "new",
		}
		c.db.Create(alert)
		c.logger.Warnf("C2 detected: %s -> %s (score: %.2f)", conn.SrcIP, conn.DstIP, score)
	}

	baseline := int64(1024 * 1024)
	if isExfil, score := c.detector.DetectDataExfiltration(&conn, baseline); isExfil {
		alert := &models.Alert{
			Type:        "data_exfiltration",
			Severity:    "critical",
			SrcIP:       conn.SrcIP,
			DstIP:       conn.DstIP,
			Description: "检测到数据渗出行为",
			Confidence:  score,
			Timestamp:   time.Now(),
			Status:      "new",
		}
		c.db.Create(alert)
	}

	return c.db.Create(&conn).Error
}

func (c *Consumer) processDNSLog(data []byte) error {
	var dnsQuery map[string]interface{}
	if err := json.Unmarshal(data, &dnsQuery); err != nil {
		return err
	}

	ctx := context.Background()

	if query, ok := dnsQuery["query"].(string); ok {
		if c.threatIntel != nil {
			if domainIntel, err := c.threatIntel.CheckDomain(ctx, query); err == nil && domainIntel != nil && domainIntel.Severity != "none" {
				alert := &models.Alert{
					Type:         "threat_intel_match",
					Severity:     domainIntel.Severity,
					Description:  fmt.Sprintf("威胁情报匹配 [%s]: %s - %s", domainIntel.ThreatLabel, query, domainIntel.Description),
					ThreatLabel:  domainIntel.ThreatLabel,
					ThreatSource: domainIntel.Source,
					Confidence:   0.95,
					Timestamp:    time.Now(),
					Status:       "new",
				}
				if srcIP, ok := dnsQuery["id.orig_h"].(string); ok {
					alert.SrcIP = srcIP
				}
				c.db.Create(alert)
				c.logger.Warnf("Threat intel match (domain): %s (%s)", query, domainIntel.ThreatLabel)
			}
		}

		if isDGA, score := c.detector.DetectDGA(query); isDGA {
			alert := &models.Alert{
				Type:        "dga_domain",
				Severity:    "medium",
				Description: "检测到DGA生成域名: " + query,
				Confidence:  score,
				Timestamp:   time.Now(),
				Status:      "new",
			}
			if srcIP, ok := dnsQuery["id.orig_h"].(string); ok {
				alert.SrcIP = srcIP
			}
			c.db.Create(alert)
		}
	}

	return nil
}

func (c *Consumer) processHTTPLog(data []byte) error {
	var httpLog map[string]interface{}
	if err := json.Unmarshal(data, &httpLog); err != nil {
		return err
	}

	// 检测WebShell
	var httpLogs []string
	if uri, ok := httpLog["uri"].(string); ok {
		httpLogs = append(httpLogs, uri)
	}
	if userAgent, ok := httpLog["user_agent"].(string); ok {
		httpLogs = append(httpLogs, userAgent)
	}

	if isWebShell, score := c.detector.DetectWebShell(httpLogs); isWebShell {
		alert := &models.Alert{
			Type:        "webshell",
			Severity:    "critical",
			Description: "检测到WebShell特征",
			Confidence:  score,
			Timestamp:   time.Now(),
		}
		c.db.Create(alert)
	}

	return nil
}

func (c *Consumer) processSSLLog(data []byte) error {
	var sslLog models.TLSHandshake
	if err := json.Unmarshal(data, &sslLog); err != nil {
		return err
	}

	return c.db.Create(&sslLog).Error
}

func (c *Consumer) processNoticeLog(data []byte) error {
	var notice map[string]interface{}
	if err := json.Unmarshal(data, &notice); err != nil {
		return err
	}

	alert := &models.Alert{
		Type:        notice["note"].(string),
		Severity:    "high",
		SrcIP:       notice["src"].(string),
		Description: notice["msg"].(string),
		Timestamp:   time.Now(),
	}

	return c.db.Create(alert).Error
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}