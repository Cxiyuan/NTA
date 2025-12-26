package zeek

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Manager struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewManager(db *gorm.DB, logger *logrus.Logger) *Manager {
	return &Manager{
		db:     db,
		logger: logger,
	}
}

func (m *Manager) GetBuiltinProbe(ctx context.Context) (*models.ZeekProbe, error) {
	var probe models.ZeekProbe
	if err := m.db.Where("probe_id = ?", "builtin-zeek").First(&probe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &probe, nil
}

func (m *Manager) CreateBuiltinProbe(ctx context.Context, probe *models.ZeekProbe) error {
	probe.ProbeID = "builtin-zeek"
	probe.Name = "内置探针"
	probe.Status = "stopped"
	return m.db.Create(probe).Error
}

func (m *Manager) UpdateProbe(ctx context.Context, probe *models.ZeekProbe) error {
	return m.db.Save(probe).Error
}

func (m *Manager) GetProbeStatus(ctx context.Context, probeID string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", "nta-zeek", "zeekctl", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "error", err
	}
	
	if strings.Contains(string(output), "running") {
		return "running", nil
	}
	return "stopped", nil
}

func (m *Manager) StartProbe(ctx context.Context, probeID string) error {
	var probe models.ZeekProbe
	if err := m.db.Where("probe_id = ?", probeID).First(&probe).Error; err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "docker", "restart", "nta-zeek")
	if err := cmd.Run(); err != nil {
		m.logger.Errorf("Failed to start zeek probe: %v", err)
		probe.Status = "error"
		m.db.Save(&probe)
		return err
	}

	probe.Status = "running"
	return m.db.Save(&probe).Error
}

func (m *Manager) StopProbe(ctx context.Context, probeID string) error {
	var probe models.ZeekProbe
	if err := m.db.Where("probe_id = ?", probeID).First(&probe).Error; err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "docker", "exec", "nta-zeek", "zeekctl", "stop")
	if err := cmd.Run(); err != nil {
		m.logger.Errorf("Failed to stop zeek probe: %v", err)
		return err
	}

	probe.Status = "stopped"
	return m.db.Save(&probe).Error
}

func (m *Manager) GetProbeStats(ctx context.Context, probeID string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", "nta-zeek", "zeekctl", "netstats")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["raw_output"] = string(output)
	stats["timestamp"] = time.Now()

	return stats, nil
}

func (m *Manager) UpdateProbeConfig(ctx context.Context, probeID string, config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return m.db.Model(&models.ZeekProbe{}).
		Where("probe_id = ?", probeID).
		Updates(map[string]interface{}{
			"interface":  config["interface"],
			"bpf_filter": config["bpf_filter"],
			"config":     string(configJSON),
		}).Error
}

func (m *Manager) GetLogs(ctx context.Context, probeID string, logType string, limit int) ([]models.ZeekLog, error) {
	var logs []models.ZeekLog
	query := m.db.Where("probe_id = ?", probeID)
	
	if logType != "" {
		query = query.Where("log_type = ?", logType)
	}
	
	if limit == 0 {
		limit = 100
	}
	
	err := query.Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (m *Manager) GetLogStats(ctx context.Context, probeID string, startTime, endTime time.Time) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	var results []struct {
		LogType string
		Count   int64
	}
	
	query := m.db.Model(&models.ZeekLog{}).
		Select("log_type, count(*) as count").
		Where("probe_id = ?", probeID)
	
	if !startTime.IsZero() {
		query = query.Where("timestamp >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("timestamp <= ?", endTime)
	}
	
	if err := query.Group("log_type").Find(&results).Error; err != nil {
		return nil, err
	}
	
	for _, r := range results {
		stats[r.LogType] = r.Count
	}
	
	return stats, nil
}

func (m *Manager) EnableScript(ctx context.Context, probeID string, scriptName string) error {
	var probe models.ZeekProbe
	if err := m.db.Where("probe_id = ?", probeID).First(&probe).Error; err != nil {
		return err
	}

	var scripts []string
	if probe.ScriptsEnabled != "" {
		json.Unmarshal([]byte(probe.ScriptsEnabled), &scripts)
	}

	for _, s := range scripts {
		if s == scriptName {
			return nil
		}
	}

	scripts = append(scripts, scriptName)
	scriptsJSON, _ := json.Marshal(scripts)
	probe.ScriptsEnabled = string(scriptsJSON)

	return m.db.Save(&probe).Error
}

func (m *Manager) DisableScript(ctx context.Context, probeID string, scriptName string) error {
	var probe models.ZeekProbe
	if err := m.db.Where("probe_id = ?", probeID).First(&probe).Error; err != nil {
		return err
	}

	var scripts []string
	if probe.ScriptsEnabled != "" {
		json.Unmarshal([]byte(probe.ScriptsEnabled), &scripts)
	}

	newScripts := []string{}
	for _, s := range scripts {
		if s != scriptName {
			newScripts = append(newScripts, s)
		}
	}

	scriptsJSON, _ := json.Marshal(newScripts)
	probe.ScriptsEnabled = string(scriptsJSON)

	return m.db.Save(&probe).Error
}

func (m *Manager) RestartProbe(ctx context.Context, probeID string) error {
	if err := m.StopProbe(ctx, probeID); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return m.StartProbe(ctx, probeID)
}

func (m *Manager) GetAvailableScripts() []map[string]string {
	return []map[string]string{
		{"name": "lateral-scan", "description": "横向扫描检测", "file": "lateral-scan.zeek"},
		{"name": "lateral-auth", "description": "横向认证检测", "file": "lateral-auth.zeek"},
		{"name": "lateral-exec", "description": "横向执行检测", "file": "lateral-exec.zeek"},
		{"name": "encrypted-traffic", "description": "加密流量分析", "file": "encrypted-traffic.zeek"},
		{"name": "attack-chain", "description": "攻击链分析", "file": "attack-chain.zeek"},
		{"name": "deep-inspection", "description": "深度包检测", "file": "deep-inspection.zeek"},
		{"name": "zeroday-detection", "description": "0day检测", "file": "zeroday-detection.zeek"},
	}
}

func (m *Manager) CleanOldLogs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	result := m.db.Where("created_at < ?", cutoff).Delete(&models.ZeekLog{})
	if result.Error != nil {
		return result.Error
	}
	m.logger.Infof("Cleaned %d old zeek logs", result.RowsAffected)
	return nil
}

func (m *Manager) ImportLog(ctx context.Context, log *models.ZeekLog) error {
	return m.db.Create(log).Error
}

func (m *Manager) GetProbeInterfaces() ([]string, error) {
	cmd := exec.Command("docker", "exec", "nta-zeek", "ip", "link", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	interfaces := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ": ") {
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				ifname := strings.TrimSpace(parts[1])
				if !strings.HasPrefix(ifname, "lo") {
					interfaces = append(interfaces, strings.Split(ifname, ":")[0])
				}
			}
		}
	}

	return interfaces, nil
}

func (m *Manager) ValidateBPFFilter(filter string) error {
	if filter == "" {
		return nil
	}
	
	cmd := exec.Command("tcpdump", "-d", filter)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid BPF filter: %v", err)
	}
	
	return nil
}
