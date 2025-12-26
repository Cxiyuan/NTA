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
)

type ThreatFoxClient struct {
	apiKey string
	logger *logrus.Logger
}

type ThreatFoxResponse struct {
	QueryStatus string `json:"query_status"`
	Data        []struct {
		ID              string   `json:"id"`
		IOC             string   `json:"ioc"`
		ThreatType      string   `json:"threat_type"`
		ThreatTypeDesc  string   `json:"threat_type_desc"`
		IOCType         string   `json:"ioc_type"`
		IOCTypeDesc     string   `json:"ioc_type_desc"`
		Malware         string   `json:"malware"`
		MalwarePrintable string  `json:"malware_printable"`
		MalwareAlias    string   `json:"malware_alias"`
		MalwareMalpedia string   `json:"malware_malpedia"`
		ConfidenceLevel int      `json:"confidence_level"`
		FirstSeen       string   `json:"first_seen"`
		LastSeen        string   `json:"last_seen"`
		Reference       string   `json:"reference"`
		Reporter        string   `json:"reporter"`
		Tags            []string `json:"tags"`
	} `json:"data"`
}

func NewThreatFoxClient(apiKey string, logger *logrus.Logger) *ThreatFoxClient {
	return &ThreatFoxClient{
		apiKey: apiKey,
		logger: logger,
	}
}

func (c *ThreatFoxClient) search(ctx context.Context, iocValue string) (*models.ThreatIntel, error) {
	reqBody := map[string]string{
		"query":       "search_ioc",
		"search_term": iocValue,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://threatfox-api.abuse.ch/api/v1/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Auth-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var tfResp ThreatFoxResponse
	if err := json.Unmarshal(body, &tfResp); err != nil {
		return nil, err
	}
	
	if tfResp.QueryStatus == "no_result" || len(tfResp.Data) == 0 {
		return nil, nil
	}
	
	data := tfResp.Data[0]
	
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
	
	iocType := "ip"
	if strings.Contains(data.IOCType, "domain") {
		iocType = "domain"
	} else if strings.Contains(data.IOCType, "url") {
		iocType = "url"
	} else if strings.Contains(data.IOCType, "hash") || strings.Contains(data.IOCType, "md5") || strings.Contains(data.IOCType, "sha") {
		iocType = "hash"
	}
	
	firstSeen := time.Now()
	lastSeen := time.Now()
	if data.FirstSeen != "" {
		if t, err := time.Parse("2006-01-02 15:04:05 MST", data.FirstSeen); err == nil {
			firstSeen = t
		}
	}
	if data.LastSeen != "" && data.LastSeen != "null" {
		if t, err := time.Parse("2006-01-02 15:04:05 MST", data.LastSeen); err == nil {
			lastSeen = t
		}
	}
	
	intel := &models.ThreatIntel{
		Type:        iocType,
		Value:       strings.Split(data.IOC, ":")[0],
		Severity:    severity,
		Source:      "threatfox",
		Description: description,
		Tags:        tags,
		FirstSeen:   firstSeen,
		LastSeen:    lastSeen,
	}
	
	intel.ThreatLabel = GetThreatLabel(intel)
	
	return intel, nil
}

func (c *ThreatFoxClient) CheckIP(ctx context.Context, ip string) (*models.ThreatIntel, error) {
	return c.search(ctx, ip)
}

func (c *ThreatFoxClient) CheckDomain(ctx context.Context, domain string) (*models.ThreatIntel, error) {
	return c.search(ctx, domain)
}

func (c *ThreatFoxClient) CheckHash(ctx context.Context, hash string) (*models.ThreatIntel, error) {
	return c.search(ctx, hash)
}

func (c *ThreatFoxClient) CheckURL(ctx context.Context, url string) (*models.ThreatIntel, error) {
	return c.search(ctx, url)
}
