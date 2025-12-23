package encryption

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
)

// Analyzer analyzes encrypted traffic (TLS/SSL)
type Analyzer struct {
	logger          *logrus.Logger
	suspiciousJA3   map[string]string // JA3 hash -> malware name
	minTLSVersion   string
}

// NewAnalyzer creates a new encryption analyzer
func NewAnalyzer(logger *logrus.Logger) *Analyzer {
	return &Analyzer{
		logger: logger,
		suspiciousJA3: map[string]string{
			"6734f37431670b3ab4292b8f60f29984": "Metasploit",
			"51c64c77e60f3980eea90869b68c58a8": "Cobalt Strike",
			"a0e9f5d64349fb13191bc781f81f42e1": "Trickbot",
		},
		minTLSVersion: "TLS 1.2",
	}
}

// AnalyzeTLS analyzes TLS handshake
func (a *Analyzer) AnalyzeTLS(hs *models.TLSHandshake) []string {
	var anomalies []string

	// Check TLS version
	if hs.Version < a.minTLSVersion {
		anomalies = append(anomalies, "outdated_tls_version")
	}

	// Check JA3 fingerprint
	if malware, exists := a.suspiciousJA3[hs.JA3]; exists {
		anomalies = append(anomalies, fmt.Sprintf("malicious_ja3:%s", malware))
	}

	// Check missing SNI
	if hs.ServerName == "" {
		anomalies = append(anomalies, "missing_sni")
	}

	// Check non-standard port
	if hs.DstPort != 443 && hs.DstPort != 8443 {
		anomalies = append(anomalies, "non_standard_tls_port")
	}

	return anomalies
}

// CalculateJA3 calculates JA3 fingerprint
func (a *Analyzer) CalculateJA3(version, ciphers, extensions string) string {
	ja3String := fmt.Sprintf("%s,%s,%s", version, ciphers, extensions)
	hash := md5.Sum([]byte(ja3String))
	return fmt.Sprintf("%x", hash)
}

// DetectC2Beacon detects command and control beacon traffic
func (a *Analyzer) DetectC2Beacon(conns []*models.Connection) []*models.Alert {
	// Group connections by src-dst pair
	connMap := make(map[string][]*models.Connection)
	
	for _, conn := range conns {
		key := fmt.Sprintf("%s-%s:%d", conn.SrcIP, conn.DstIP, conn.DstPort)
		connMap[key] = append(connMap[key], conn)
	}

	var alerts []*models.Alert

	// Detect beaconing patterns
	for key, group := range connMap {
		if len(group) < 5 {
			continue
		}

		// Check for regular intervals
		if isRegularInterval(group) {
			parts := strings.Split(key, "-")
			alerts = append(alerts, &models.Alert{
				Type:        "c2_beacon",
				Severity:    "high",
				SrcIP:       parts[0],
				DstIP:       strings.Split(parts[1], ":")[0],
				Description: "C2 beacon traffic detected",
				Confidence:  0.85,
			})
		}
	}

	return alerts
}

func isRegularInterval(conns []*models.Connection) bool {
	if len(conns) < 5 {
		return false
	}

	// Calculate intervals
	var intervals []float64
	for i := 1; i < len(conns); i++ {
		interval := conns[i].Timestamp.Sub(conns[i-1].Timestamp).Seconds()
		intervals = append(intervals, interval)
	}

	// Check if intervals are similar (within 20% variance)
	avg := average(intervals)
	for _, interval := range intervals {
		variance := (interval - avg) / avg
		if variance > 0.2 || variance < -0.2 {
			return false
		}
	}

	return true
}

func average(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0.0
	for _, n := range nums {
		sum += n
	}
	return sum / float64(len(nums))
}
