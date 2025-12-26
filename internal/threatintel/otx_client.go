package threatintel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
)

type OTXClient struct {
	apiKey string
	logger *logrus.Logger
}

type OTXResponse struct {
	PulseInfo struct {
		Count  int `json:"count"`
		Pulses []struct {
			Name            string   `json:"name"`
			Description     string   `json:"description"`
			Adversary       string   `json:"adversary"`
			Tags            []string `json:"tags"`
			MalwareFamilies []struct {
				DisplayName string `json:"display_name"`
			} `json:"malware_families"`
		} `json:"pulses"`
	} `json:"pulse_info"`
	Validation []struct {
		Source string `json:"source"`
		Name   string `json:"name"`
	} `json:"validation"`
	Malware struct {
		Count int `json:"count"`
		Data  []struct {
			Hash       string `json:"hash"`
			Detections int    `json:"detections"`
		} `json:"data"`
	} `json:"malware"`
}

func NewOTXClient(apiKey string, logger *logrus.Logger) *OTXClient {
	return &OTXClient{
		apiKey: apiKey,
		logger: logger,
	}
}

func (c *OTXClient) CheckIP(ctx context.Context, ip string) (*models.ThreatIntel, error) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/IPv4/%s/general", ip)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-OTX-API-KEY", c.apiKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OTX API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var otxResp OTXResponse
	if err := json.Unmarshal(body, &otxResp); err != nil {
		return nil, err
	}
	
	if otxResp.PulseInfo.Count > 0 || len(otxResp.Validation) > 0 {
		severity := "medium"
		if otxResp.PulseInfo.Count > 5 {
			severity = "high"
		}
		
		description := fmt.Sprintf("Found in %d OTX pulses", otxResp.PulseInfo.Count)
		if len(otxResp.Validation) > 0 {
			description += fmt.Sprintf(", validated by %s", otxResp.Validation[0].Name)
		}
		
		tags := ""
		adversary := ""
		malwareFamily := ""
		
		if len(otxResp.PulseInfo.Pulses) > 0 {
			pulse := otxResp.PulseInfo.Pulses[0]
			if pulse.Adversary != "" {
				adversary = pulse.Adversary
				description += fmt.Sprintf(", APT: %s", adversary)
			}
			if len(pulse.MalwareFamilies) > 0 {
				malwareFamily = pulse.MalwareFamilies[0].DisplayName
				description += fmt.Sprintf(", Malware: %s", malwareFamily)
			}
			if len(pulse.Tags) > 0 {
				tagsJSON, _ := json.Marshal(pulse.Tags)
				tags = string(tagsJSON)
			}
		}
		
		intel := &models.ThreatIntel{
			Type:        "ip",
			Value:       ip,
			Severity:    severity,
			Source:      "alienvault_otx",
			Description: description,
			Tags:        tags,
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
		}
		
		intel.ThreatLabel = GetThreatLabel(intel)
		
		return intel, nil
	}
	
	return nil, nil
}

func (c *OTXClient) CheckDomain(ctx context.Context, domain string) (*models.ThreatIntel, error) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/general", domain)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-OTX-API-KEY", c.apiKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OTX API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var otxResp OTXResponse
	if err := json.Unmarshal(body, &otxResp); err != nil {
		return nil, err
	}
	
	if otxResp.PulseInfo.Count > 0 {
		severity := "medium"
		if otxResp.PulseInfo.Count > 5 {
			severity = "high"
		}
		
		description := fmt.Sprintf("Found in %d OTX pulses", otxResp.PulseInfo.Count)
		tags := ""
		
		if len(otxResp.PulseInfo.Pulses) > 0 {
			pulse := otxResp.PulseInfo.Pulses[0]
			if pulse.Adversary != "" {
				description += fmt.Sprintf(", APT: %s", pulse.Adversary)
			}
			if len(pulse.MalwareFamilies) > 0 {
				description += fmt.Sprintf(", Malware: %s", pulse.MalwareFamilies[0].DisplayName)
			}
			if len(pulse.Tags) > 0 {
				tagsJSON, _ := json.Marshal(pulse.Tags)
				tags = string(tagsJSON)
			}
		}
		
		intel := &models.ThreatIntel{
			Type:        "domain",
			Value:       domain,
			Severity:    severity,
			Source:      "alienvault_otx",
			Description: description,
			Tags:        tags,
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
		}
		
		intel.ThreatLabel = GetThreatLabel(intel)
		
		return intel, nil
	}
	
	return nil, nil
}

func (c *OTXClient) CheckHash(ctx context.Context, hash string) (*models.ThreatIntel, error) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/file/%s/analysis", hash)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-OTX-API-KEY", c.apiKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OTX API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var otxResp OTXResponse
	if err := json.Unmarshal(body, &otxResp); err != nil {
		return nil, err
	}
	
	if otxResp.Malware.Count > 0 && len(otxResp.Malware.Data) > 0 {
		detections := otxResp.Malware.Data[0].Detections
		
		severity := "low"
		if detections > 10 {
			severity = "high"
		} else if detections > 3 {
			severity = "medium"
		}
		
		intel := &models.ThreatIntel{
			Type:        "hash",
			Value:       hash,
			Severity:    severity,
			Source:      "alienvault_otx",
			Description: fmt.Sprintf("Detected by %d malware engines", detections),
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
		}
		
		intel.ThreatLabel = GetThreatLabel(intel)
		
		return intel, nil
	}
	
	return nil, nil
}

func (c *OTXClient) CheckURL(ctx context.Context, url string) (*models.ThreatIntel, error) {
	apiURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/url/%s/general", url)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("X-OTX-API-KEY", c.apiKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OTX API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var otxResp OTXResponse
	if err := json.Unmarshal(body, &otxResp); err != nil {
		return nil, err
	}
	
	if otxResp.PulseInfo.Count > 0 {
		description := fmt.Sprintf("Malicious URL in %d OTX pulses", otxResp.PulseInfo.Count)
		tags := ""
		
		if len(otxResp.PulseInfo.Pulses) > 0 {
			pulse := otxResp.PulseInfo.Pulses[0]
			if pulse.Adversary != "" {
				description += fmt.Sprintf(", APT: %s", pulse.Adversary)
			}
			if len(pulse.MalwareFamilies) > 0 {
				description += fmt.Sprintf(", Malware: %s", pulse.MalwareFamilies[0].DisplayName)
			}
			if len(pulse.Tags) > 0 {
				tagsJSON, _ := json.Marshal(pulse.Tags)
				tags = string(tagsJSON)
			}
		}
		
		intel := &models.ThreatIntel{
			Type:        "url",
			Value:       url,
			Severity:    "high",
			Source:      "alienvault_otx",
			Description: description,
			Tags:        tags,
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
		}
		
		intel.ThreatLabel = GetThreatLabel(intel)
		
		return intel, nil
	}
	
	return nil, nil
}