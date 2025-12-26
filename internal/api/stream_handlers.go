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
		stream.GET("/flink/status", h.GetFlinkStatus)
		stream.GET("/flink/jobs", h.GetFlinkJobs)
		stream.GET("/flink/jobs/:jobId", h.GetFlinkJobDetail)
		stream.POST("/flink/jobs/:jobId/cancel", h.CancelFlinkJob)
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
			"members":  5,
			"lag":      215,
			"state":    "Stable",
		},
		{
			"group_id": "flink-c2-detector",
			"members":  4,
			"lag":      0,
			"state":    "Stable",
		},
		{
			"group_id": "flink-dga-detector",
			"members":  4,
			"lag":      0,
			"state":    "Stable",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
}

func (h *StreamProcessingHandlers) GetFlinkStatus(c *gin.Context) {
	status, err := h.kafkaManager.GetFlinkStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *StreamProcessingHandlers) GetFlinkJobs(c *gin.Context) {
	jobs := []gin.H{
		{
			"job_id":     "a1b2c3d4e5f6",
			"name":       "C2 Beacon Detection",
			"status":     "RUNNING",
			"start_time": "2025-12-26T10:00:00Z",
			"duration":   "2h 30m",
			"tasks": gin.H{
				"total":   4,
				"running": 4,
				"failed":  0,
			},
		},
		{
			"job_id":     "b2c3d4e5f6g7",
			"name":       "DGA Detection",
			"status":     "RUNNING",
			"start_time": "2025-12-26T10:00:00Z",
			"duration":   "2h 30m",
			"tasks": gin.H{
				"total":   4,
				"running": 4,
				"failed":  0,
			},
		},
		{
			"job_id":     "c3d4e5f6g7h8",
			"name":       "Data Exfiltration Detection",
			"status":     "RUNNING",
			"start_time": "2025-12-26T10:00:00Z",
			"duration":   "2h 30m",
			"tasks": gin.H{
				"total":   4,
				"running": 4,
				"failed":  0,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

func (h *StreamProcessingHandlers) GetFlinkJobDetail(c *gin.Context) {
	jobID := c.Param("jobId")

	detail := gin.H{
		"job_id":     jobID,
		"name":       "C2 Beacon Detection",
		"status":     "RUNNING",
		"start_time": "2025-12-26T10:00:00Z",
		"plan": gin.H{
			"nodes": []gin.H{
				{
					"id":          "source-1",
					"description": "Kafka Source (zeek-conn)",
					"parallelism": 8,
				},
				{
					"id":          "window-1",
					"description": "Sliding Window Aggregation",
					"parallelism": 4,
				},
				{
					"id":          "sink-1",
					"description": "JDBC Sink (alerts)",
					"parallelism": 2,
				},
			},
		},
		"metrics": gin.H{
			"records_in":  125430,
			"records_out": 23,
			"bytes_in":    "45.2 MB",
			"bytes_out":   "12.5 KB",
		},
	}

	c.JSON(http.StatusOK, detail)
}

func (h *StreamProcessingHandlers) CancelFlinkJob(c *gin.Context) {
	jobID := c.Param("jobId")

	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancellation requested",
		"job_id":  jobID,
	})
}
