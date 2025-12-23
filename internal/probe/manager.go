package probe

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Manager manages multiple probe instances
type Manager struct {
	db            *gorm.DB
	redis         *redis.Client
	logger        *logrus.Logger
	probes        map[string]*models.Probe
	mu            sync.RWMutex
	probeTimeout  time.Duration
	heartbeatChan chan string
}

// NewManager creates a new probe manager
func NewManager(db *gorm.DB, rdb *redis.Client, logger *logrus.Logger) *Manager {
	return &Manager{
		db:            db,
		redis:         rdb,
		logger:        logger,
		probes:        make(map[string]*models.Probe),
		probeTimeout:  2 * time.Minute,
		heartbeatChan: make(chan string, 100),
	}
}

// RegisterProbe registers a new probe
func (m *Manager) RegisterProbe(ctx context.Context, probe *models.Probe) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	probe.Status = "online"
	probe.LastHeartbeat = time.Now()

	// Save to database
	if err := m.db.Create(probe).Error; err != nil {
		return err
	}

	// Store in memory
	m.probes[probe.ProbeID] = probe

	// Publish to Redis
	data, _ := json.Marshal(probe)
	m.redis.Set(ctx, "probe:"+probe.ProbeID, data, m.probeTimeout)
	m.redis.Publish(ctx, "probe:register", probe.ProbeID)

	m.logger.Infof("Probe registered: %s (%s)", probe.ProbeID, probe.Hostname)

	return nil
}

// UpdateHeartbeat updates probe heartbeat
func (m *Manager) UpdateHeartbeat(ctx context.Context, probeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	probe, exists := m.probes[probeID]
	if !exists {
		// Load from database
		var dbProbe models.Probe
		if err := m.db.Where("probe_id = ?", probeID).First(&dbProbe).Error; err != nil {
			return err
		}
		probe = &dbProbe
		m.probes[probeID] = probe
	}

	probe.LastHeartbeat = time.Now()
	probe.Status = "online"

	// Update database
	m.db.Model(probe).Updates(map[string]interface{}{
		"last_heartbeat": probe.LastHeartbeat,
		"status":         probe.Status,
	})

	// Extend TTL in Redis
	data, _ := json.Marshal(probe)
	m.redis.Set(ctx, "probe:"+probeID, data, m.probeTimeout)

	return nil
}

// GetProbe retrieves probe information
func (m *Manager) GetProbe(probeID string) (*models.Probe, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	probe, exists := m.probes[probeID]
	if exists {
		return probe, nil
	}

	// Load from database
	var dbProbe models.Probe
	if err := m.db.Where("probe_id = ?", probeID).First(&dbProbe).Error; err != nil {
		return nil, err
	}

	return &dbProbe, nil
}

// ListProbes returns all registered probes
func (m *Manager) ListProbes() []*models.Probe {
	m.mu.RLock()
	defer m.mu.RUnlock()

	probes := make([]*models.Probe, 0, len(m.probes))
	for _, probe := range m.probes {
		probes = append(probes, probe)
	}

	return probes
}

// CheckProbeHealth checks health of all probes
func (m *Manager) CheckProbeHealth(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for probeID, probe := range m.probes {
		if now.Sub(probe.LastHeartbeat) > m.probeTimeout {
			probe.Status = "offline"
			m.db.Model(probe).Update("status", "offline")
			m.logger.Warnf("Probe offline: %s", probeID)
		}
	}
}

// DistributeAlert sends alert to all probes
func (m *Manager) DistributeAlert(ctx context.Context, alert *models.Alert) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	// Publish to Redis channel
	return m.redis.Publish(ctx, "alerts", data).Err()
}

// SubscribeAlerts subscribes to alert channel
func (m *Manager) SubscribeAlerts(ctx context.Context, callback func(*models.Alert)) error {
	pubsub := m.redis.Subscribe(ctx, "alerts")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			var alert models.Alert
			if err := json.Unmarshal([]byte(msg.Payload), &alert); err != nil {
				m.logger.Errorf("Failed to unmarshal alert: %v", err)
				continue
			}
			callback(&alert)
		}
	}
}

// StartHealthCheck starts periodic health check
func (m *Manager) StartHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.CheckProbeHealth(ctx)
		}
	}
}

// RemoveProbe removes a probe from registry
func (m *Manager) RemoveProbe(ctx context.Context, probeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from memory
	delete(m.probes, probeID)

	// Remove from Redis
	m.redis.Del(ctx, "probe:"+probeID)

	// Mark as offline in database
	m.db.Model(&models.Probe{}).Where("probe_id = ?", probeID).Update("status", "offline")

	m.logger.Infof("Probe removed: %s", probeID)

	return nil
}
