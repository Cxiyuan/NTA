package threatintel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FeedSyncer struct {
	db              *gorm.DB
	logger          *logrus.Logger
	otxClient       *OTXClient
	threatFoxClient *ThreatFoxClient
	updateInterval  time.Duration
	updateHour      int
}

func NewFeedSyncer(db *gorm.DB, logger *logrus.Logger, otxClient *OTXClient, threatFoxClient *ThreatFoxClient, updateIntervalHours int, updateHour int) *FeedSyncer {
	return &FeedSyncer{
		db:              db,
		logger:          logger,
		otxClient:       otxClient,
		threatFoxClient: threatFoxClient,
		updateInterval:  time.Duration(updateIntervalHours) * time.Hour,
		updateHour:      updateHour,
	}
}

func (fs *FeedSyncer) Start(ctx context.Context) {
	fs.logger.Info("Starting threat intelligence feed syncer")

	fs.syncNow(ctx)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fs.logger.Info("Feed syncer stopped")
			return
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == fs.updateHour {
				fs.syncNow(ctx)
			}
		}
	}
}

func (fs *FeedSyncer) syncNow(ctx context.Context) {
	fs.logger.Info("Starting threat intelligence feed synchronization")
	
	var totalAdded int
	var totalUpdated int

	if fs.threatFoxClient != nil {
		added, updated, err := fs.syncThreatFox(ctx)
		if err != nil {
			fs.logger.Errorf("ThreatFox sync failed: %v", err)
		} else {
			totalAdded += added
			totalUpdated += updated
			fs.logger.Infof("ThreatFox sync completed: %d added, %d updated", added, updated)
		}
	}

	if fs.otxClient != nil {
		added, updated, err := fs.syncAlienVaultOTX(ctx)
		if err != nil {
			fs.logger.Errorf("AlienVault OTX sync failed: %v", err)
		} else {
			totalAdded += added
			totalUpdated += updated
			fs.logger.Infof("AlienVault OTX sync completed: %d added, %d updated", added, updated)
		}
	}

	fs.logger.Infof("Feed sync completed: total %d added, %d updated", totalAdded, totalUpdated)
}

func (fs *FeedSyncer) syncThreatFox(ctx context.Context) (int, int, error) {
	reqBody := map[string]string{
		"query": "get_iocs",
		"days":  "3",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://threatfox-api.abuse.ch/api/v1/", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var tfResp ThreatFoxResponse
	if err := json.Unmarshal(body, &tfResp); err != nil {
		return 0, 0, err
	}

	if tfResp.QueryStatus != "ok" {
		return 0, 0, fmt.Errorf("ThreatFox API returned status: %s", tfResp.QueryStatus)
	}

	added := 0
	updated := 0

	for _, data := range tfResp.Data {
		iocType := "ip"
		if strings.Contains(data.IOCType, "domain") {
			iocType = "domain"
		} else if strings.Contains(data.IOCType, "url") {
			iocType = "url"
		} else if strings.Contains(data.IOCType, "hash") || strings.Contains(data.IOCType, "md5") || strings.Contains(data.IOCType, "sha") {
			iocType = "hash"
		}

		severity := "medium"
		if data.ConfidenceLevel >= 90 {
			severity = "high"
		} else if data.ConfidenceLevel >= 75 {
			severity = "medium"
		} else {
			severity = "low"
		}

		description := fmt.Sprintf("%s (%s)", data.ThreatTypeDesc, data.MalwarePrintable)
		if data.MalwareAlias != "" {
			description += fmt.Sprintf(", Alias: %s", data.MalwareAlias)
		}

		tags := ""
		if len(data.Tags) > 0 {
			tagsJSON, _ := json.Marshal(data.Tags)
			tags = string(tagsJSON)
		}

		firstSeen := time.Now()
		if data.FirstSeen != "" {
			if t, err := time.Parse("2006-01-02 15:04:05 MST", data.FirstSeen); err == nil {
				firstSeen = t
			}
		}

		validUntil := firstSeen.Add(90 * 24 * time.Hour)

		iocValue := strings.Split(data.IOC, ":")[0]

		intel := models.ThreatIntel{
			Type:        iocType,
			Value:       iocValue,
			Severity:    severity,
			Source:      "threatfox",
			Description: description,
			Tags:        tags,
			FirstSeen:   firstSeen,
			LastSeen:    time.Now(),
			ValidUntil:  validUntil,
		}
		intel.ThreatLabel = GetThreatLabel(&intel)

		var existing models.ThreatIntel
		err := fs.db.Where("type = ? AND value = ? AND source = ?", iocType, iocValue, "threatfox").First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := fs.db.Create(&intel).Error; err != nil {
				fs.logger.Warnf("Failed to insert ThreatFox IOC %s: %v", iocValue, err)
			} else {
				added++
			}
		} else if err == nil {
			intel.ID = existing.ID
			if err := fs.db.Model(&existing).Updates(&intel).Error; err != nil {
				fs.logger.Warnf("Failed to update ThreatFox IOC %s: %v", iocValue, err)
			} else {
				updated++
			}
		}
	}

	return added, updated, nil
}

func (fs *FeedSyncer) syncAlienVaultOTX(ctx context.Context) (int, int, error) {
	url := "https://otx.alienvault.com/api/v1/pulses/subscribed"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("X-OTX-API-KEY", fs.otxClient.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("OTX API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var otxResp struct {
		Results []struct {
			Indicators []struct {
				Type        string `json:"type"`
				Indicator   string `json:"indicator"`
				Description string `json:"description"`
			} `json:"indicators"`
			Tags            []string `json:"tags"`
			Adversary       string   `json:"adversary"`
			MalwareFamilies []struct {
				DisplayName string `json:"display_name"`
			} `json:"malware_families"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &otxResp); err != nil {
		return 0, 0, err
	}

	added := 0
	updated := 0

	for _, pulse := range otxResp.Results {
		for _, indicator := range pulse.Indicators {
			iocType := "ip"
			switch indicator.Type {
			case "IPv4", "IPv6":
				iocType = "ip"
			case "domain", "hostname":
				iocType = "domain"
			case "URL":
				iocType = "url"
			case "FileHash-MD5", "FileHash-SHA1", "FileHash-SHA256":
				iocType = "hash"
			default:
				continue
			}

			description := indicator.Description
			if pulse.Adversary != "" {
				description += fmt.Sprintf(", APT: %s", pulse.Adversary)
			}
			if len(pulse.MalwareFamilies) > 0 {
				description += fmt.Sprintf(", Malware: %s", pulse.MalwareFamilies[0].DisplayName)
			}

			tags := ""
			if len(pulse.Tags) > 0 {
				tagsJSON, _ := json.Marshal(pulse.Tags)
				tags = string(tagsJSON)
			}

			intel := models.ThreatIntel{
				Type:        iocType,
				Value:       indicator.Indicator,
				Severity:    "medium",
				Source:      "alienvault_otx",
				Description: description,
				Tags:        tags,
				FirstSeen:   time.Now(),
				LastSeen:    time.Now(),
				ValidUntil:  time.Now().Add(90 * 24 * time.Hour),
			}
			intel.ThreatLabel = GetThreatLabel(&intel)

			var existing models.ThreatIntel
			err := fs.db.Where("type = ? AND value = ? AND source = ?", iocType, indicator.Indicator, "alienvault_otx").First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				if err := fs.db.Create(&intel).Error; err != nil {
					fs.logger.Warnf("Failed to insert OTX IOC %s: %v", indicator.Indicator, err)
				} else {
					added++
				}
			} else if err == nil {
				intel.ID = existing.ID
				if err := fs.db.Model(&existing).Updates(&intel).Error; err != nil {
					fs.logger.Warnf("Failed to update OTX IOC %s: %v", indicator.Indicator, err)
				} else {
					updated++
				}
			}

			if added+updated >= 10000 {
				return added, updated, nil
			}
		}
	}

	return added, updated, nil
}
