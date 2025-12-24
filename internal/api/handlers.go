package api

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) listReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reports, total, err := s.reportService.List(page, pageSize)
	if err != nil {
		s.logger.Errorf("Failed to list reports: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      reports,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func (s *Server) generateReport(c *gin.Context) {
	var req struct {
		Type      string    `json:"type" binding:"required"`
		StartTime time.Time `json:"start_time" binding:"required"`
		EndTime   time.Time `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	report, err := s.reportService.Generate(req.Type, req.StartTime, req.EndTime, username.(string))
	if err != nil {
		s.logger.Errorf("Failed to generate report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (s *Server) downloadReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	filePath, err := s.reportService.GetFilePath(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.FileAttachment(filePath, "report.txt")
}

func (s *Server) getNotificationConfig(c *gin.Context) {
	config := s.notifyService.GetConfig()
	c.JSON(http.StatusOK, config)
}

func (s *Server) updateNotificationConfig(c *gin.Context) {
	var config models.NotificationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.notifyService.UpdateConfig(&config); err != nil {
		s.logger.Errorf("Failed to update notification config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) testNotification(c *gin.Context) {
	var req struct {
		Channel string `json:"channel" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.notifyService.TestNotification(req.Channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

func (s *Server) listPCAPSessions(c *gin.Context) {
	srcIP := c.Query("src_ip")
	dstIP := c.Query("dst_ip")
	
	var startTime, endTime time.Time
	if st := c.Query("start_time"); st != "" {
		startTime, _ = time.Parse(time.RFC3339, st)
	}
	if et := c.Query("end_time"); et != "" {
		endTime, _ = time.Parse(time.RFC3339, et)
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	sessions, err := s.pcapStorage.SearchSessions(srcIP, dstIP, startTime, endTime, limit)
	if err != nil {
		s.logger.Errorf("Failed to search PCAP sessions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func (s *Server) downloadPCAP(c *gin.Context) {
	sessionID := c.Param("id")

	filePath, err := s.pcapStorage.GetSessionFile(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.FileAttachment(filePath, sessionID+".pcap")
}

func (s *Server) searchPCAP(c *gin.Context) {
	var req struct {
		SrcIP     string    `json:"src_ip"`
		DstIP     string    `json:"dst_ip"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		Limit     int       `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessions, err := s.pcapStorage.SearchSessions(req.SrcIP, req.DstIP, req.StartTime, req.EndTime, req.Limit)
	if err != nil {
		s.logger.Errorf("Failed to search PCAP: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
