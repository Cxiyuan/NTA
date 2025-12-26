package api

import (
	"github.com/Cxiyuan/NTA/services/report-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ReportHandler struct {
	service *service.ReportService
	logger  *logrus.Logger
}

func NewReportHandler(service *service.ReportService, logger *logrus.Logger) *ReportHandler {
	return &ReportHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReportHandler) RegisterRoutes(r *gin.RouterGroup) {
	reports := r.Group("/reports")
	{
		reports.GET("", h.listReports)
		reports.POST("", h.createReport)
		reports.GET("/:id", h.getReport)
	}
}

func (h *ReportHandler) listReports(c *gin.Context) {
	c.JSON(200, gin.H{"reports": []interface{}{}})
}

func (h *ReportHandler) createReport(c *gin.Context) {
	c.JSON(201, gin.H{"message": "Report created"})
}

func (h *ReportHandler) getReport(c *gin.Context) {
	c.JSON(200, gin.H{"report": gin.H{}})
}
