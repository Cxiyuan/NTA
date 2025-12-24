package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nta_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nta_http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	AlertsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nta_alerts_total",
			Help: "Total number of alerts generated",
		},
		[]string{"severity", "type"},
	)

	ActiveProbes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "nta_active_probes",
			Help: "Number of active probes",
		},
	)

	ThreatIntelCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nta_threat_intel_cache_hits_total",
			Help: "Total number of threat intelligence cache hits",
		},
	)

	ThreatIntelCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nta_threat_intel_cache_misses_total",
			Help: "Total number of threat intelligence cache misses",
		},
	)

	PacketsProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nta_packets_processed_total",
			Help: "Total number of packets processed",
		},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nta_database_query_duration_seconds",
			Help:    "Database query latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	LateralMovementDetections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nta_lateral_movement_detections_total",
			Help: "Total number of lateral movement detections",
		},
		[]string{"type"},
	)
)
