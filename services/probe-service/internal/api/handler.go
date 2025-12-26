package api

import (
	"github.com/Cxiyuan/NTA/services/probe-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProbeHandler struct {
	service *service.ProbeService
	logger  *logrus.Logger
}

func NewProbeHandler(service *service.ProbeService, logger *logrus.Logger) *ProbeHandler {
	return &ProbeHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProbeHandler) RegisterRoutes(r *gin.RouterGroup) {
	probes := r.Group("/probes")
	{
		probes.GET("", h.listProbes)
		probes.POST("", h.createProbe)
		probes.GET("/:id", h.getProbe)
		probes.GET("/:id/status", h.getProbeStatus)
	}
}

func (h *ProbeHandler) listProbes(c *gin.Context) {
	c.JSON(200, gin.H{"probes": []interface{}{}})
}

func (h *ProbeHandler) createProbe(c *gin.Context) {
	c.JSON(201, gin.H{"message": "Probe created"})
}

func (h *ProbeHandler) getProbe(c *gin.Context) {
	c.JSON(200, gin.H{"probe": gin.H{}})
}

func (h *ProbeHandler) getProbeStatus(c *gin.Context) {
	c.JSON(200, gin.H{"status": "active"})
}
