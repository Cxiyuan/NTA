package api

import (
	"net/http"
	"strconv"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/services/alert-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AlertHandler struct {
	service *service.AlertService
	logger  *logrus.Logger
}

func NewAlertHandler(service *service.AlertService, logger *logrus.Logger) *AlertHandler {
	return &AlertHandler{service: service, logger: logger}
}

func (h *AlertHandler) RegisterRoutes(r *gin.RouterGroup) {
	alerts := r.Group("/alerts")
	{
		alerts.GET("", h.ListAlerts)
		alerts.GET("/:id", h.GetAlert)
		alerts.POST("", h.CreateAlert)
		alerts.PUT("/:id", h.UpdateAlert)
		alerts.DELETE("/:id", h.DeleteAlert)
	}
}

func (h *AlertHandler) ListAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	filters := make(map[string]interface{})
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	alerts, total, err := h.service.ListAlerts(c.Request.Context(), pageSize, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      alerts,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func (h *AlertHandler) GetAlert(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	alert, err := h.service.GetAlertByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}
	c.JSON(http.StatusOK, alert)
}

func (h *AlertHandler) CreateAlert(c *gin.Context) {
	var alert models.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateAlert(c.Request.Context(), &alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, alert)
}

func (h *AlertHandler) UpdateAlert(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateAlert(c.Request.Context(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert updated"})
}

func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.DeleteAlert(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert deleted"})
}
