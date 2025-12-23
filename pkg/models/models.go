package models

import (
	"time"
)

// Alert represents a security alert
type Alert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // critical, high, medium, low
	Type        string    `json:"type"`
	SrcIP       string    `json:"src_ip"`
	DstIP       string    `json:"dst_ip"`
	SrcPort     int       `json:"src_port"`
	DstPort     int       `json:"dst_port"`
	Protocol    string    `json:"protocol"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	Details     string    `json:"details" gorm:"type:text"`
	Status      string    `json:"status"` // new, investigating, resolved, false_positive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Asset represents a discovered network asset
type Asset struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	IP          string    `json:"ip" gorm:"uniqueIndex"`
	MAC         string    `json:"mac"`
	Hostname    string    `json:"hostname"`
	Vendor      string    `json:"vendor"`
	OS          string    `json:"os"`
	Services    string    `json:"services" gorm:"type:text"` // JSON array
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ThreatIntel represents threat intelligence data
type ThreatIntel struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Type       string    `json:"type"` // ip, domain, hash
	Value      string    `json:"value" gorm:"uniqueIndex"`
	Severity   string    `json:"severity"`
	Source     string    `json:"source"`
	Tags       string    `json:"tags"` // JSON array
	ValidUntil time.Time `json:"valid_until"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Probe represents a deployed probe instance
type Probe struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ProbeID      string    `json:"probe_id" gorm:"uniqueIndex"`
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	Version      string    `json:"version"`
	Status       string    `json:"status"` // online, offline, error
	Capabilities string    `json:"capabilities" gorm:"type:text"` // JSON array
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APTIndicator represents APT detection indicator
type APTIndicator struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Entity      string    `json:"entity"` // IP or user
	Phase       string    `json:"phase"` // Kill Chain phase
	EventType   string    `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	Score       float64   `json:"score"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditLog represents audit trail
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details" gorm:"type:text"`
	Result    string    `json:"result"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
}

// Connection represents Zeek connection data
type Connection struct {
	UID       string    `json:"uid"`
	Timestamp time.Time `json:"ts"`
	SrcIP     string    `json:"src_ip"`
	SrcPort   int       `json:"src_port"`
	DstIP     string    `json:"dst_ip"`
	DstPort   int       `json:"dst_port"`
	Protocol  string    `json:"proto"`
	Service   string    `json:"service"`
	Duration  float64   `json:"duration"`
	OrigBytes int64     `json:"orig_bytes"`
	RespBytes int64     `json:"resp_bytes"`
	ConnState string    `json:"conn_state"`
}

// TLSHandshake represents TLS connection metadata
type TLSHandshake struct {
	Timestamp   time.Time `json:"ts"`
	UID         string    `json:"uid"`
	SrcIP       string    `json:"src_ip"`
	DstIP       string    `json:"dst_ip"`
	SrcPort     int       `json:"src_port"`
	DstPort     int       `json:"dst_port"`
	Version     string    `json:"version"`
	CipherSuite string    `json:"cipher"`
	ServerName  string    `json:"server_name"`
	JA3         string    `json:"ja3"`
	JA3S        string    `json:"ja3s"`
}
