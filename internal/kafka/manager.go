package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Manager struct {
	kafkaAdminURL string
	flinkURL      string
	logger        *logrus.Logger
}

type KafkaClusterStatus struct {
	Brokers     []BrokerInfo          `json:"brokers"`
	Topics      []TopicInfo           `json:"topics"`
	ConsumerGroups []ConsumerGroupInfo `json:"consumer_groups"`
	Health      string                `json:"health"`
}

type BrokerInfo struct {
	ID   int    `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type TopicInfo struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
	Messages   int64  `json:"messages"`
}

type ConsumerGroupInfo struct {
	GroupID string `json:"group_id"`
	Lag     int64  `json:"lag"`
	Members int    `json:"members"`
}

type FlinkClusterStatus struct {
	JobManager  JobManagerInfo `json:"job_manager"`
	TaskManagers []TaskManagerInfo `json:"task_managers"`
	Jobs        []JobInfo      `json:"jobs"`
	Health      string         `json:"health"`
}

type JobManagerInfo struct {
	Address string `json:"address"`
	Slots   int    `json:"slots"`
}

type TaskManagerInfo struct {
	ID    string `json:"id"`
	Slots int    `json:"slots"`
}

type JobInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	StartTime time.Time `json:"start_time"`
}

func NewManager(kafkaAdminURL, flinkURL string, logger *logrus.Logger) *Manager {
	return &Manager{
		kafkaAdminURL: kafkaAdminURL,
		flinkURL:      flinkURL,
		logger:        logger,
	}
}

func (m *Manager) GetKafkaStatus(ctx context.Context) (*KafkaClusterStatus, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	resp, err := client.Get(m.kafkaAdminURL + "/admin/cluster")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status KafkaClusterStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (m *Manager) GetFlinkStatus(ctx context.Context) (*FlinkClusterStatus, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	resp, err := client.Get(fmt.Sprintf("%s/overview", m.flinkURL))
	if err != nil {
		m.logger.Warnf("Failed to get Flink status: %v", err)
		return &FlinkClusterStatus{
			Health: "unavailable",
		}, nil
	}
	defer resp.Body.Close()

	var overview map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return nil, err
	}

	status := &FlinkClusterStatus{
		Health: "healthy",
	}

	return status, nil
}

func (m *Manager) GetTopicLag(ctx context.Context, topic, groupID string) (int64, error) {
	return 0, nil
}
