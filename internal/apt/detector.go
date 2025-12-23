package apt

import (
	"sync"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
)

// Detector detects APT (Advanced Persistent Threat) activities
type Detector struct {
	logger       *logrus.Logger
	killChain    map[string]*KillChain
	iocDatabase  map[string][]string // type -> list of IOCs
	mu           sync.RWMutex
}

// KillChain tracks attack kill chain phases
type KillChain struct {
	Entity    string
	Phases    map[string]*models.APTIndicator
	Score     float64
	FirstSeen time.Time
	LastSeen  time.Time
}

var killChainPhases = []string{
	"reconnaissance",
	"weaponization",
	"delivery",
	"exploitation",
	"installation",
	"command_control",
	"actions_objectives",
}

// NewDetector creates a new APT detector
func NewDetector(logger *logrus.Logger) *Detector {
	return &Detector{
		logger:      logger,
		killChain:   make(map[string]*KillChain),
		iocDatabase: make(map[string][]string),
	}
}

// AnalyzeEvent analyzes event for APT indicators
func (d *Detector) AnalyzeEvent(entity, eventType string, timestamp time.Time) *models.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	chain, exists := d.killChain[entity]
	if !exists {
		chain = &KillChain{
			Entity:    entity,
			Phases:    make(map[string]*models.APTIndicator),
			FirstSeen: timestamp,
		}
		d.killChain[entity] = chain
	}

	chain.LastSeen = timestamp

	// Map event type to kill chain phase
	phase := d.mapEventToPhase(eventType)
	if phase != "" {
		chain.Phases[phase] = &models.APTIndicator{
			Entity:    entity,
			Phase:     phase,
			EventType: eventType,
			Timestamp: timestamp,
			Score:     0.7,
		}
	}

	// Check if multiple phases detected
	if len(chain.Phases) >= 3 {
		return &models.Alert{
			Type:        "apt_kill_chain",
			Severity:    "critical",
			SrcIP:       entity,
			Description: fmt.Sprintf("APT kill chain detected: %d phases", len(chain.Phases)),
			Confidence:  0.95,
			Timestamp:   timestamp,
		}
	}

	return nil
}

func (d *Detector) mapEventToPhase(eventType string) string {
	mapping := map[string]string{
		"port_scan":             "reconnaissance",
		"host_discovery":        "reconnaissance",
		"malware_download":      "weaponization",
		"exploit_attempt":       "exploitation",
		"buffer_overflow":       "exploitation",
		"persistence_mechanism": "installation",
		"registry_modification": "installation",
		"c2_communication":      "command_control",
		"beacon_traffic":        "command_control",
		"data_exfiltration":     "actions_objectives",
		"lateral_movement":      "actions_objectives",
	}

	return mapping[eventType]
}

// HuntIOC searches for indicators of compromise
func (d *Detector) HuntIOC(iocType, value string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	iocs, exists := d.iocDatabase[iocType]
	if !exists {
		return false
	}

	for _, ioc := range iocs {
		if ioc == value {
			return true
		}
	}

	return false
}

// LoadIOCs loads IOC database
func (d *Detector) LoadIOCs(iocs map[string][]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.iocDatabase = iocs
	d.logger.Infof("Loaded %d IOC types", len(iocs))
}

// GetKillChains returns all detected kill chains
func (d *Detector) GetKillChains() []*KillChain {
	d.mu.RLock()
	defer d.mu.RUnlock()

	chains := make([]*KillChain, 0, len(d.killChain))
	for _, chain := range d.killChain {
		chains = append(chains, chain)
	}

	return chains
}

// CleanOldChains removes old kill chain data
func (d *Detector) CleanOldChains(maxAge time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for entity, chain := range d.killChain {
		if chain.LastSeen.Before(cutoff) {
			delete(d.killChain, entity)
		}
	}
}
