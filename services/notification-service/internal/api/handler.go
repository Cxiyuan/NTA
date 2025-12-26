package api

import (
	"github.com/Cxiyuan/NTA/services/notification-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type NotificationHandler struct {
	service *service.NotificationService
	logger  *logrus.Logger
}

func NewNotificationHandler(service *service.NotificationService, logger *logrus.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *NotificationHandler) RegisterRoutes(r *gin.RouterGroup) {
	notifications := r.Group("/notifications")
	{
		notifications.GET("", h.listNotifications)
		notifications.POST("", h.createNotification)
		notifications.PUT("/:id/read", h.markAsRead)
	}
}

func (h *NotificationHandler) listNotifications(c *gin.Context) {
	c.JSON(200, gin.H{"notifications": []interface{}{}})
}

func (h *NotificationHandler) createNotification(c *gin.Context) {
	c.JSON(201, gin.H{"message": "Notification created"})
}

func (h *NotificationHandler) markAsRead(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Marked as read"})
}
