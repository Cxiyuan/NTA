package service

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

type DetectionService struct {
	logger *logrus.Logger
}

func NewDetectionService(logger *logrus.Logger) *DetectionService {
	return &DetectionService{logger: logger}
}

type DetectionResult struct {
	IsMalicious bool    `json:"is_malicious"`
	Confidence  float64 `json:"confidence"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
}

func (s *DetectionService) DetectDGA(ctx context.Context, domain string) *DetectionResult {
	score := 0.0
	domainLen := len(domain)
	
	if domainLen > 20 {
		score += 0.3
	}
	
	digitCount := 0
	for _, c := range domain {
		if c >= '0' && c <= '9' {
			digitCount++
		}
	}
	if digitCount > 5 {
		score += 0.4
	}
	
	if !strings.HasSuffix(domain, ".com") && !strings.HasSuffix(domain, ".cn") {
		score += 0.3
	}

	isMalicious := score > 0.5

	return &DetectionResult{
		IsMalicious: isMalicious,
		Confidence:  score,
		Type:        "dga_domain",
		Description: "DGA域名检测",
	}
}

func (s *DetectionService) DetectC2(ctx context.Context, srcIP, dstIP string, packetCount int, avgInterval float64) *DetectionResult {
	score := 0.0

	if avgInterval > 20 && avgInterval < 120 {
		score += 0.5
	}

	if packetCount > 10 {
		score += 0.3
	}

	isMalicious := score > 0.6

	return &DetectionResult{
		IsMalicious: isMalicious,
		Confidence:  score,
		Type:        "c2_beacon",
		Description: "C2信标检测",
	}
}

func (s *DetectionService) DetectDNSTunnel(ctx context.Context, query string, queryLen int) *DetectionResult {
	score := 0.0

	if queryLen > 50 {
		score += 0.5
	}

	subdomains := strings.Count(query, ".")
	if subdomains > 5 {
		score += 0.3
	}

	isMalicious := score > 0.5

	return &DetectionResult{
		IsMalicious: isMalicious,
		Confidence:  score,
		Type:        "dns_tunnel",
		Description: "DNS隧道检测",
	}
}

func (s *DetectionService) DetectWebShell(ctx context.Context, uri, method string) *DetectionResult {
	score := 0.0

	suspiciousPatterns := []string{"eval", "assert", "base64_decode", "shell_exec", "exec", "system"}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(uri), pattern) {
			score += 0.4
			break
		}
	}

	if method == "POST" && strings.Contains(uri, ".php") {
		score += 0.2
	}

	isMalicious := score > 0.4

	return &DetectionResult{
		IsMalicious: isMalicious,
		Confidence:  score,
		Type:        "webshell",
		Description: "WebShell检测",
	}
}
