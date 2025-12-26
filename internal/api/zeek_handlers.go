package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) getBuiltinProbe(c *gin.Context) {
	probe, err := s.zeekManager.GetBuiltinProbe(c.Request.Context())
	if err != nil {
		s.logger.Errorf("Failed to get builtin probe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get probe"})
		return
	}

	if probe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "builtin probe not initialized"})
		return
	}

	c.JSON(http.StatusOK, probe)
}

func (s *Server) updateBuiltinProbe(c *gin.Context) {
	var req struct {
		Interface  string `json:"interface"`
		BPFFilter  string `json:"bpf_filter"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.zeekManager.ValidateBPFFilter(req.BPFFilter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid BPF filter", "message": err.Error()})
		return
	}

	config := map[string]interface{}{
		"interface":  req.Interface,
		"bpf_filter": req.BPFFilter,
	}

	if err := s.zeekManager.UpdateProbeConfig(c.Request.Context(), "builtin-zeek", config); err != nil {
		s.logger.Errorf("Failed to update probe config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	s.auditService.Log(c.GetString("username"), "update_builtin_probe", "", map[string]interface{}{
		"interface":  req.Interface,
		"bpf_filter": req.BPFFilter,
	})

	c.JSON(http.StatusOK, gin.H{"message": "configuration updated"})
}

func (s *Server) getBuiltinProbeStatus(c *gin.Context) {
	status, err := s.zeekManager.GetProbeStatus(c.Request.Context(), "builtin-zeek")
	if err != nil {
		s.logger.Errorf("Failed to get probe status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status"})
		return
	}

	stats, _ := s.zeekManager.GetProbeStats(c.Request.Context(), "builtin-zeek")

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"stats":  stats,
	})
}

func (s *Server) startBuiltinProbe(c *gin.Context) {
	if err := s.zeekManager.StartProbe(c.Request.Context(), "builtin-zeek"); err != nil {
		s.logger.Errorf("Failed to start probe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start probe", "message": err.Error()})
		return
	}

	s.auditService.Log(c.GetString("username"), "start_builtin_probe", "", nil)

	c.JSON(http.StatusOK, gin.H{"message": "probe started"})
}

func (s *Server) stopBuiltinProbe(c *gin.Context) {
	if err := s.zeekManager.StopProbe(c.Request.Context(), "builtin-zeek"); err != nil {
		s.logger.Errorf("Failed to stop probe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop probe", "message": err.Error()})
		return
	}

	s.auditService.Log(c.GetString("username"), "stop_builtin_probe", "", nil)

	c.JSON(http.StatusOK, gin.H{"message": "probe stopped"})
}

func (s *Server) restartBuiltinProbe(c *gin.Context) {
	if err := s.zeekManager.RestartProbe(c.Request.Context(), "builtin-zeek"); err != nil {
		s.logger.Errorf("Failed to restart probe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to restart probe", "message": err.Error()})
		return
	}

	s.auditService.Log(c.GetString("username"), "restart_builtin_probe", "", nil)

	c.JSON(http.StatusOK, gin.H{"message": "probe restarted"})
}

func (s *Server) getBuiltinProbeScripts(c *gin.Context) {
	scripts := s.zeekManager.GetAvailableScripts()
	
	probe, err := s.zeekManager.GetBuiltinProbe(c.Request.Context())
	if err == nil && probe != nil {
		for i := range scripts {
			scripts[i]["enabled"] = "false"
		}
	}

	c.JSON(http.StatusOK, gin.H{"scripts": scripts})
}

func (s *Server) enableBuiltinProbeScript(c *gin.Context) {
	scriptName := c.Param("script")

	if err := s.zeekManager.EnableScript(c.Request.Context(), "builtin-zeek", scriptName); err != nil {
		s.logger.Errorf("Failed to enable script: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable script"})
		return
	}

	s.auditService.Log(c.GetString("username"), "enable_probe_script", scriptName, nil)

	c.JSON(http.StatusOK, gin.H{"message": "script enabled"})
}

func (s *Server) disableBuiltinProbeScript(c *gin.Context) {
	scriptName := c.Param("script")

	if err := s.zeekManager.DisableScript(c.Request.Context(), "builtin-zeek", scriptName); err != nil {
		s.logger.Errorf("Failed to disable script: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable script"})
		return
	}

	s.auditService.Log(c.GetString("username"), "disable_probe_script", scriptName, nil)

	c.JSON(http.StatusOK, gin.H{"message": "script disabled"})
}

func (s *Server) getBuiltinProbeLogs(c *gin.Context) {
	logType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	logs, err := s.zeekManager.GetLogs(c.Request.Context(), "builtin-zeek", logType, limit)
	if err != nil {
		s.logger.Errorf("Failed to get logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs, "count": len(logs)})
}

func (s *Server) getBuiltinProbeLogStats(c *gin.Context) {
	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
	}
	if endTimeStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endTimeStr)
	}

	stats, err := s.zeekManager.GetLogStats(c.Request.Context(), "builtin-zeek", startTime, endTime)
	if err != nil {
		s.logger.Errorf("Failed to get log stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (s *Server) getBuiltinProbeInterfaces(c *gin.Context) {
	interfaces, err := s.zeekManager.GetProbeInterfaces()
	if err != nil {
		s.logger.Errorf("Failed to get interfaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get interfaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interfaces": interfaces})
}
