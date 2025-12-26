package api

import (
	"github.com/Cxiyuan/NTA/services/intel-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type IntelHandler struct {
	service *service.IntelService
	logger  *logrus.Logger
}

func NewIntelHandler(service *service.IntelService, logger *logrus.Logger) *IntelHandler {
	return &IntelHandler{
		service: service,
		logger:  logger,
	}
}

func (h *IntelHandler) RegisterRoutes(r *gin.RouterGroup) {
	intel := r.Group("/intel")
	{
		intel.GET("/threats", h.listThreats)
		intel.GET("/threats/:id", h.getThreat)
		intel.POST("/threats/search", h.searchThreats)
		intel.GET("/indicators", h.listIndicators)
	}
}

func (h *IntelHandler) listThreats(c *gin.Context) {
	c.JSON(200, gin.H{"threats": []interface{}{}})
}

func (h *IntelHandler) getThreat(c *gin.Context) {
	c.JSON(200, gin.H{"threat": gin.H{}})
}

func (h *IntelHandler) searchThreats(c *gin.Context) {
	c.JSON(200, gin.H{"results": []interface{}{}})
}

func (h *IntelHandler) listIndicators(c *gin.Context) {
	c.JSON(200, gin.H{"indicators": []interface{}{}})
}
