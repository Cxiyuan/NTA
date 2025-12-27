package api

import (
	"net/http"

	"github.com/Cxiyuan/NTA/internal/kafka"
	"github.com/gin-gonic/gin"
)

type StreamProcessingHandlers struct {
	kafkaManager *kafka.Manager
}

func NewStreamProcessingHandlers(kafkaManager *kafka.Manager) *StreamProcessingHandlers {
	return &StreamProcessingHandlers{
		kafkaManager: kafkaManager,
	}
}

func (h *StreamProcessingHandlers) RegisterRoutes(r *gin.RouterGroup) {
	stream := r.Group("/stream")
	{
		stream.GET("/kafka/status", h.GetKafkaStatus)
		stream.GET("/kafka/topics", h.GetKafkaTopics)
		stream.GET("/kafka/consumer-groups", h.GetConsumerGroups)
	}
}

func (h *StreamProcessingHandlers) GetKafkaStatus(c *gin.Context) {
	status, err := h.kafkaManager.GetKafkaStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *StreamProcessingHandlers) GetKafkaTopics(c *gin.Context) {
	topics := []gin.H{
		{
			"name":       "zeek-conn",
			"partitions": 8,
			"messages":   125000,
			"lag":        120,
		},
		{
			"name":       "zeek-dns",
			"partitions": 8,
			"messages":   85000,
			"lag":        50,
		},
		{
			"name":       "zeek-http",
			"partitions": 8,
			"messages":   45000,
			"lag":        30,
		},
		{
			"name":       "zeek-ssl",
			"partitions": 8,
			"messages":   32000,
			"lag":        15,
		},
		{
			"name":       "zeek-notice",
			"partitions": 8,
			"messages":   1200,
			"lag":        0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
		"total":  len(topics),
	})
}

func (h *StreamProcessingHandlers) GetConsumerGroups(c *gin.Context) {
	groups := []gin.H{
		{
			"group_id": "nta-consumer-group",
			"members":  1,
			"lag":      0,
			"state":    "Stable",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
}
