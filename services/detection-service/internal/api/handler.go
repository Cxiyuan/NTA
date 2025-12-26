package api

import (
	"net/http"

	"github.com/Cxiyuan/NTA/services/detection-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DetectionHandler struct {
	service *service.DetectionService
	logger  *logrus.Logger
}

func NewDetectionHandler(service *service.DetectionService, logger *logrus.Logger) *DetectionHandler {
	return &DetectionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *DetectionHandler) RegisterRoutes(r *gin.RouterGroup) {
	detection := r.Group("/detection")
	{
		detection.POST("/dga", h.DetectDGA)
		detection.POST("/c2", h.DetectC2)
		detection.POST("/dns-tunnel", h.DetectDNSTunnel)
		detection.POST("/webshell", h.DetectWebShell)
	}
}

type DGARequest struct {
	Domain string `json:"domain" binding:"required"`
}

type C2Request struct {
	SrcIP       string  `json:"src_ip" binding:"required"`
	DstIP       string  `json:"dst_ip" binding:"required"`
	PacketCount int     `json:"packet_count"`
	AvgInterval float64 `json:"avg_interval"`
}

type DNSTunnelRequest struct {
	Query     string `json:"query" binding:"required"`
	QueryLen  int    `json:"query_len"`
}

type WebShellRequest struct {
	URI    string `json:"uri" binding:"required"`
	Method string `json:"method" binding:"required"`
}

func (h *DetectionHandler) DetectDGA(c *gin.Context) {
	var req DGARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.DetectDGA(c.Request.Context(), req.Domain)
	c.JSON(http.StatusOK, result)
}

func (h *DetectionHandler) DetectC2(c *gin.Context) {
	var req C2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.DetectC2(c.Request.Context(), req.SrcIP, req.DstIP, req.PacketCount, req.AvgInterval)
	c.JSON(http.StatusOK, result)
}

func (h *DetectionHandler) DetectDNSTunnel(c *gin.Context) {
	var req DNSTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.QueryLen == 0 {
		req.QueryLen = len(req.Query)
	}

	result := h.service.DetectDNSTunnel(c.Request.Context(), req.Query, req.QueryLen)
	c.JSON(http.StatusOK, result)
}

func (h *DetectionHandler) DetectWebShell(c *gin.Context) {
	var req WebShellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.DetectWebShell(c.Request.Context(), req.URI, req.Method)
	c.JSON(http.StatusOK, result)
}
