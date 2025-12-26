package client

import (
	"fmt"
	"net/http"
	"time"
)

type MicroserviceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMicroserviceClient(baseURL string) *MicroserviceClient {
	return &MicroserviceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *MicroserviceClient) Get(endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	return c.httpClient.Get(url)
}

func (c *MicroserviceClient) Post(endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	return c.httpClient.Post(url, "application/json", nil)
}

var (
	AuthServiceURL   = "http://localhost:8081"
	AssetServiceURL  = "http://localhost:8082"
	DetectionServiceURL = "http://localhost:8083"
	AlertServiceURL  = "http://localhost:8084"
	ReportServiceURL = "http://localhost:8085"
	NotificationServiceURL = "http://localhost:8086"
	ProbeServiceURL  = "http://localhost:8087"
	IntelServiceURL  = "http://localhost:8088"
)
