package analyzer

import (
	"sync"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
)

// LateralMovementDetector detects lateral movement attacks
type LateralMovementDetector struct {
	scanTracker   map[string]*ScanTracker
	authTracker   map[string]*AuthTracker
	execTracker   map[string]*ExecTracker
	mu            sync.RWMutex
	logger        *logrus.Logger
	scanThreshold int
	timeWindow    int
}

type ScanTracker struct {
	SourceIP     string
	TargetIPs    map[string]bool
	Timestamps   []time.Time
	FailureCount int
}

type AuthTracker struct {
	SourceIP      string
	FailedAttempts int
	HashSeen      map[string][]string // hash -> list of target IPs
	LastSeen      time.Time
}

type ExecTracker struct {
	SourceIP  string
	TargetIP  string
	Events    []string
	Timestamp time.Time
}

// NewLateralMovementDetector creates a new detector
func NewLateralMovementDetector(logger *logrus.Logger, scanThreshold, timeWindow int) *LateralMovementDetector {
	return &LateralMovementDetector{
		scanTracker:   make(map[string]*ScanTracker),
		authTracker:   make(map[string]*AuthTracker),
		execTracker:   make(map[string]*ExecTracker),
		logger:        logger,
		scanThreshold: scanThreshold,
		timeWindow:    timeWindow,
	}
}

// DetectScan detects port scanning and host discovery
func (d *LateralMovementDetector) DetectScan(conn *models.Connection) *models.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	tracker, exists := d.scanTracker[conn.SrcIP]
	if !exists {
		tracker = &ScanTracker{
			SourceIP:  conn.SrcIP,
			TargetIPs: make(map[string]bool),
			Timestamps: []time.Time{},
		}
		d.scanTracker[conn.SrcIP] = tracker
	}

	tracker.TargetIPs[conn.DstIP] = true
	tracker.Timestamps = append(tracker.Timestamps, conn.Timestamp)

	// Clean old timestamps
	cutoff := time.Now().Add(-time.Duration(d.timeWindow) * time.Second)
	validTimestamps := []time.Time{}
	for _, ts := range tracker.Timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	tracker.Timestamps = validTimestamps

	// Check if threshold exceeded
	if len(tracker.TargetIPs) >= d.scanThreshold {
		return &models.Alert{
			Timestamp:   time.Now(),
			Severity:    "high",
			Type:        "lateral_scan",
			SrcIP:       conn.SrcIP,
			Description: "Lateral movement scan detected",
			Confidence:  0.9,
		}
	}

	return nil
}

// DetectPTH detects Pass-the-Hash attacks
func (d *LateralMovementDetector) DetectPTH(srcIP, hash, dstIP string) *models.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := srcIP + ":" + hash
	tracker, exists := d.authTracker[key]
	if !exists {
		tracker = &AuthTracker{
			SourceIP: srcIP,
			HashSeen: make(map[string][]string),
			LastSeen: time.Now(),
		}
		d.authTracker[key] = tracker
	}

	tracker.HashSeen[hash] = append(tracker.HashSeen[hash], dstIP)
	tracker.LastSeen = time.Now()

	// If same hash used on 3+ different hosts
	if len(tracker.HashSeen[hash]) >= 3 {
		return &models.Alert{
			Timestamp:   time.Now(),
			Severity:    "critical",
			Type:        "pass_the_hash",
			SrcIP:       srcIP,
			Description: "Pass-the-Hash attack detected",
			Confidence:  0.95,
		}
	}

	return nil
}

// DetectRemoteExec detects PSExec/WMI remote execution
func (d *LateralMovementDetector) DetectRemoteExec(srcIP, dstIP, method string) *models.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := srcIP + ":" + dstIP
	tracker, exists := d.execTracker[key]
	if !exists {
		tracker = &ExecTracker{
			SourceIP:  srcIP,
			TargetIP:  dstIP,
			Events:    []string{},
			Timestamp: time.Now(),
		}
		d.execTracker[key] = tracker
	}

	tracker.Events = append(tracker.Events, method)
	tracker.Timestamp = time.Now()

	// Detect PSExec pattern: ADMIN$ access + svcctl
	if contains(tracker.Events, "admin_share") && contains(tracker.Events, "svcctl") {
		return &models.Alert{
			Timestamp:   time.Now(),
			Severity:    "critical",
			Type:        "psexec",
			SrcIP:       srcIP,
			DstIP:       dstIP,
			Description: "PSExec remote execution detected",
			Confidence:  0.92,
		}
	}

	// Detect WMI execution
	if contains(tracker.Events, "wmi_exec") {
		return &models.Alert{
			Timestamp:   time.Now(),
			Severity:    "high",
			Type:        "wmi_exec",
			SrcIP:       srcIP,
			DstIP:       dstIP,
			Description: "WMI remote execution detected",
			Confidence:  0.88,
		}
	}

	return nil
}

// Cleanup removes old tracking data
func (d *LateralMovementDetector) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)

	// Clean scan tracker
	for ip, tracker := range d.scanTracker {
		if len(tracker.Timestamps) == 0 || tracker.Timestamps[len(tracker.Timestamps)-1].Before(cutoff) {
			delete(d.scanTracker, ip)
		}
	}

	// Clean auth tracker
	for key, tracker := range d.authTracker {
		if tracker.LastSeen.Before(cutoff) {
			delete(d.authTracker, key)
		}
	}

	// Clean exec tracker
	for key, tracker := range d.execTracker {
		if tracker.Timestamp.Before(cutoff) {
			delete(d.execTracker, key)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
