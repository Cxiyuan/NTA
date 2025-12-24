package detector

import (
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
)

type AdvancedDetector struct {
	logger *logrus.Logger
}

func NewAdvancedDetector(logger *logrus.Logger) *AdvancedDetector {
	return &AdvancedDetector{
		logger: logger,
	}
}

func (d *AdvancedDetector) DetectDGA(domain string) (bool, float64) {
	if len(domain) < 5 {
		return false, 0
	}

	score := 0.0
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false, 0
	}

	subdomain := parts[0]

	vowelCount := 0
	consonantCount := 0
	digitCount := 0
	for _, c := range subdomain {
		if c >= '0' && c <= '9' {
			digitCount++
		} else if strings.ContainsRune("aeiouAEIOU", c) {
			vowelCount++
		} else if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			consonantCount++
		}
	}

	totalChars := float64(len(subdomain))
	if totalChars == 0 {
		return false, 0
	}

	vowelRatio := float64(vowelCount) / totalChars
	digitRatio := float64(digitCount) / totalChars

	if vowelRatio < 0.2 || vowelRatio > 0.6 {
		score += 0.3
	}

	if digitRatio > 0.3 {
		score += 0.2
	}

	if len(subdomain) > 15 {
		score += 0.2
	}

	entropy := d.calculateEntropy(subdomain)
	if entropy > 3.5 {
		score += 0.3
	}

	if matched, _ := regexp.MatchString(`^[a-z0-9]{10,}$`, subdomain); matched {
		score += 0.2
	}

	isDGA := score > 0.6
	return isDGA, math.Min(score, 1.0)
}

func (d *AdvancedDetector) DetectDNSTunnel(dnsQueries []models.Connection) (bool, float64) {
	if len(dnsQueries) < 10 {
		return false, 0
	}

	var totalLength int
	var queryCount int
	domainLengthMap := make(map[string]int)

	for _, query := range dnsQueries {
		if query.DstPort == 53 {
			queryCount++
			domainLengthMap[query.Service]++
		}
	}

	if queryCount == 0 {
		return false, 0
	}

	score := 0.0

	avgLength := float64(totalLength) / float64(queryCount)
	if avgLength > 50 {
		score += 0.4
	}

	requestRate := float64(queryCount) / 60.0
	if requestRate > 10 {
		score += 0.3
	}

	uniqueDomains := len(domainLengthMap)
	if uniqueDomains > 20 {
		score += 0.3
	}

	isTunnel := score > 0.6
	return isTunnel, math.Min(score, 1.0)
}

func (d *AdvancedDetector) DetectC2Communication(conn *models.Connection) (bool, float64, string) {
	score := 0.0
	c2Type := ""

	if conn.Duration > 300 && conn.OrigBytes < 1000 && conn.RespBytes < 1000 {
		score += 0.3
		c2Type = "beacon"
	}

	if conn.DstPort == 443 || conn.DstPort == 8443 {
		if conn.OrigBytes > 0 && conn.RespBytes > 0 {
			ratio := float64(conn.OrigBytes) / float64(conn.RespBytes)
			if ratio > 0.8 && ratio < 1.2 {
				score += 0.2
			}
		}
	}

	uncommonPorts := []int{4444, 5555, 6666, 7777, 8888, 9999, 1337, 31337}
	for _, port := range uncommonPorts {
		if conn.DstPort == port {
			score += 0.3
			c2Type = "uncommon_port"
			break
		}
	}

	if conn.ConnState == "S0" || conn.ConnState == "REJ" {
		score = score * 0.5
	}

	isC2 := score > 0.5
	if isC2 && c2Type == "" {
		c2Type = "suspicious"
	}

	return isC2, math.Min(score, 1.0), c2Type
}

func (d *AdvancedDetector) DetectWebShell(httpLogs []string) (bool, float64) {
	if len(httpLogs) == 0 {
		return false, 0
	}

	score := 0.0
	suspiciousPatterns := []string{
		"eval\\(",
		"base64_decode",
		"system\\(",
		"exec\\(",
		"passthru",
		"shell_exec",
		"phpinfo\\(",
		"assert\\(",
		"<?php",
		"cmd=",
		"<?=",
	}

	matchCount := 0
	for _, log := range httpLogs {
		for _, pattern := range suspiciousPatterns {
			if matched, _ := regexp.MatchString(pattern, log); matched {
				matchCount++
				break
			}
		}
	}

	if matchCount > 0 {
		score = math.Min(float64(matchCount)*0.3, 1.0)
	}

	isWebShell := score > 0.5
	return isWebShell, score
}

func (d *AdvancedDetector) DetectDataExfiltration(conn *models.Connection, baseline int64) (bool, float64) {
	score := 0.0

	if conn.OrigBytes > baseline*5 {
		score += 0.4
	}

	if conn.Duration < 60 && conn.OrigBytes > 10*1024*1024 {
		score += 0.3
	}

	uncommonPorts := conn.DstPort > 1024 && conn.DstPort != 8080 && conn.DstPort != 8443
	if uncommonPorts {
		score += 0.2
	}

	outsideBusinessHours := time.Now().Hour() < 7 || time.Now().Hour() > 19
	if outsideBusinessHours {
		score += 0.1
	}

	isExfiltration := score > 0.6
	return isExfiltration, math.Min(score, 1.0)
}

func (d *AdvancedDetector) calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}
