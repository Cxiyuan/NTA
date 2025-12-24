package api

import (
	"net/http"

	"github.com/Cxiyuan/NTA/internal/detector"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) detectDGA(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	advancedDetector := detector.NewAdvancedDetector(s.logger)
	isDGA, confidence := advancedDetector.DetectDGA(req.Domain)

	c.JSON(http.StatusOK, gin.H{
		"is_dga":     isDGA,
		"confidence": confidence,
		"domain":     req.Domain,
	})
}

func (s *Server) detectDNSTunnel(c *gin.Context) {
	var req struct {
		SrcIP      string `json:"src_ip" binding:"required"`
		TimeWindow int    `json:"time_window"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var connections []models.Connection
	s.db.Where("src_ip = ? AND dst_port = 53", req.SrcIP).
		Order("timestamp DESC").
		Limit(100).
		Find(&connections)

	advancedDetector := detector.NewAdvancedDetector(s.logger)
	isTunnel, confidence := advancedDetector.DetectDNSTunnel(connections)

	c.JSON(http.StatusOK, gin.H{
		"is_tunnel":   isTunnel,
		"confidence":  confidence,
		"query_count": len(connections),
	})
}

func (s *Server) detectC2(c *gin.Context) {
	var req struct {
		SrcIP   string `json:"src_ip" binding:"required"`
		DstIP   string `json:"dst_ip" binding:"required"`
		DstPort int    `json:"dst_port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var conn models.Connection
	err := s.db.Where("src_ip = ? AND dst_ip = ? AND dst_port = ?",
		req.SrcIP, req.DstIP, req.DstPort).
		Order("timestamp DESC").
		First(&conn).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	advancedDetector := detector.NewAdvancedDetector(s.logger)
	isC2, confidence, c2Type := advancedDetector.DetectC2Communication(&conn)

	c.JSON(http.StatusOK, gin.H{
		"is_c2":      isC2,
		"confidence": confidence,
		"c2_type":    c2Type,
		"duration":   conn.Duration,
		"orig_bytes": conn.OrigBytes,
		"resp_bytes": conn.RespBytes,
	})
}

func (s *Server) detectWebShell(c *gin.Context) {
	var req struct {
		URL     string `json:"url" binding:"required"`
		Payload string `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	httpLogs := []string{req.URL, req.Payload}

	advancedDetector := detector.NewAdvancedDetector(s.logger)
	isWebShell, confidence := advancedDetector.DetectWebShell(httpLogs)

	c.JSON(http.StatusOK, gin.H{
		"is_webshell": isWebShell,
		"confidence":  confidence,
	})
}
