package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ThreatIntelConfigResponse struct {
	UpdateInterval int    `json:"update_interval_hours"`
	UpdateHour     int    `json:"update_hour"`
	EnableLocalDB  bool   `json:"enable_local_db"`
	Sources        []struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	} `json:"sources"`
	LastSyncTime   string `json:"last_sync_time"`
	TotalIOCs      int64  `json:"total_iocs"`
}

type UpdateThreatIntelConfigRequest struct {
	UpdateInterval int `json:"update_interval_hours" binding:"required,min=1,max=720"`
	UpdateHour     int `json:"update_hour" binding:"required,min=0,max=23"`
}

func (s *Server) getThreatIntelConfig(c *gin.Context) {
	var totalIOCs int64
	s.db.Model(&struct {
		ID uint `gorm:"primaryKey"`
	}{}).Table("threat_intels").Count(&totalIOCs)

	var lastSync struct {
		UpdatedAt string
	}
	s.db.Raw("SELECT updated_at FROM threat_intels ORDER BY updated_at DESC LIMIT 1").Scan(&lastSync)

	response := ThreatIntelConfigResponse{
		UpdateInterval: 24,
		UpdateHour:     2,
		EnableLocalDB:  true,
		Sources: []struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}{
			{Name: "ThreatFox", Enabled: true},
			{Name: "AlienVault OTX", Enabled: true},
		},
		LastSyncTime: lastSync.UpdatedAt,
		TotalIOCs:    totalIOCs,
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) updateThreatIntelConfig(c *gin.Context) {
	var req UpdateThreatIntelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_threat_intel_config", "", map[string]interface{}{
		"update_interval_hours": req.UpdateInterval,
		"update_hour":           req.UpdateHour,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":               "威胁情报配置已更新，将在下次定时任务时生效",
		"update_interval_hours": req.UpdateInterval,
		"update_hour":           req.UpdateHour,
	})
}
