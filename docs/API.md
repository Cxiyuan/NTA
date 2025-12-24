# NTA API Documentation

## Overview
NTA (Network Traffic Analysis) provides RESTful APIs for security monitoring, threat detection, and network asset management.

**Base URL**: `http://your-server:8080/api/v1`

**Authentication**: All API endpoints (except `/health`) require JWT Bearer token authentication.

## Authentication

### Generate Token
Contact your administrator to obtain a JWT token. Include it in all requests:

```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Health Check

#### GET /health
Check service health status.

**No authentication required**

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-01T12:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": {"status": "ok"},
    "redis": {"status": "ok"}
  }
}
```

---

### Alerts

#### GET /api/v1/alerts
List security alerts with pagination and filtering.

**Required Role:** `admin`, `analyst`, `viewer`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `page_size` (int, default: 50, max: 100) - Items per page
- `severity` (string) - Filter by severity: `critical`, `high`, `medium`, `low`
- `status` (string) - Filter by status: `new`, `investigating`, `resolved`, `false_positive`

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "timestamp": "2025-01-01T12:00:00Z",
      "severity": "high",
      "type": "lateral_scan",
      "src_ip": "192.168.1.100",
      "dst_ip": "192.168.1.200",
      "description": "Lateral movement scan detected",
      "confidence": 0.9,
      "status": "new"
    }
  ],
  "page": 1,
  "page_size": 50,
  "total": 150
}
```

#### GET /api/v1/alerts/:id
Get alert details by ID.

**Required Role:** `admin`, `analyst`, `viewer`

**Response:**
```json
{
  "id": 1,
  "timestamp": "2025-01-01T12:00:00Z",
  "severity": "critical",
  "type": "pass_the_hash",
  "src_ip": "192.168.1.100",
  "dst_ip": "192.168.1.200",
  "description": "Pass-the-Hash attack detected",
  "confidence": 0.95,
  "details": "{\"hash\": \"ntlm:...\", \"targets\": 5}",
  "status": "investigating"
}
```

#### PUT /api/v1/alerts/:id
Update alert status.

**Required Role:** `admin`, `analyst`

**Request Body:**
```json
{
  "status": "resolved"
}
```

**Valid Status Values:** `new`, `investigating`, `resolved`, `false_positive`

**Response:**
```json
{
  "status": "updated"
}
```

---

### Assets

#### GET /api/v1/assets
List discovered network assets.

**Required Role:** `admin`, `analyst`, `viewer`

**Response:**
```json
[
  {
    "id": 1,
    "ip": "192.168.1.100",
    "mac": "00:11:22:33:44:55",
    "hostname": "workstation-01",
    "vendor": "Dell Inc.",
    "os": "Windows 10",
    "services": "[\"http\", \"smb\", \"rdp\"]",
    "first_seen": "2025-01-01T08:00:00Z",
    "last_seen": "2025-01-01T12:00:00Z"
  }
]
```

#### GET /api/v1/assets/:ip
Get asset details by IP address.

**Required Role:** `admin`, `analyst`, `viewer`

**Response:**
```json
{
  "id": 1,
  "ip": "192.168.1.100",
  "mac": "00:11:22:33:44:55",
  "hostname": "workstation-01",
  "vendor": "Dell Inc.",
  "os": "Windows 10",
  "services": "[\"http\", \"smb\", \"rdp\"]",
  "first_seen": "2025-01-01T08:00:00Z",
  "last_seen": "2025-01-01T12:00:00Z"
}
```

---

### Threat Intelligence

#### GET /api/v1/threat-intel/check
Check if an IOC (Indicator of Compromise) is malicious.

**Required Role:** `admin`, `analyst`

**Query Parameters:**
- `type` (string, required) - IOC type: `ip`, `domain`, `hash`
- `value` (string, required) - IOC value to check

**Example:** `GET /api/v1/threat-intel/check?type=ip&value=1.2.3.4`

**Response:**
```json
{
  "type": "ip",
  "value": "1.2.3.4",
  "severity": "high",
  "source": "threatfox",
  "tags": "[\"malware\", \"botnet\"]",
  "valid_until": "2025-12-31T23:59:59Z"
}
```

#### POST /api/v1/threat-intel/update
Manually trigger threat intelligence feed update.

**Required Role:** `admin`

**Response:**
```json
{
  "status": "updated"
}
```

---

### Probes

#### POST /api/v1/probes/register
Register a new probe instance.

**Required Role:** `admin`

**Request Body:**
```json
{
  "probe_id": "probe-001",
  "hostname": "nta-probe-01",
  "ip_address": "10.0.1.100",
  "version": "1.0.0",
  "capabilities": "[\"packet_capture\", \"threat_detection\"]"
}
```

**Response:**
```json
{
  "id": 1,
  "probe_id": "probe-001",
  "hostname": "nta-probe-01",
  "ip_address": "10.0.1.100",
  "version": "1.0.0",
  "status": "online",
  "last_heartbeat": "2025-01-01T12:00:00Z"
}
```

#### POST /api/v1/probes/:id/heartbeat
Send probe heartbeat.

**Required Role:** All authenticated users

**Response:**
```json
{
  "status": "ok"
}
```

#### GET /api/v1/probes
List all registered probes.

**Required Role:** `admin`, `analyst`

**Response:**
```json
[
  {
    "id": 1,
    "probe_id": "probe-001",
    "hostname": "nta-probe-01",
    "ip_address": "10.0.1.100",
    "status": "online",
    "last_heartbeat": "2025-01-01T12:00:00Z"
  }
]
```

---

### Audit Logs

#### GET /api/v1/audit
Query audit logs.

**Required Role:** `admin`

**Query Parameters:**
- `user` (string) - Filter by username
- `action` (string) - Filter by action type

**Response:**
```json
[
  {
    "id": 1,
    "timestamp": "2025-01-01T12:00:00Z",
    "user": "admin",
    "action": "update_alert",
    "resource": "alert:123",
    "details": "{\"status\": \"resolved\"}",
    "result": "success",
    "checksum": "abc123..."
  }
]
```

---

### License

#### GET /api/v1/license
Get license information.

**Required Role:** `admin`

**Response:**
```json
{
  "customer": "ACME Corporation",
  "product": "NTA Enterprise",
  "max_probes": 50,
  "max_bandwidth_mbps": 10000,
  "issue_date": "2025-01-01T00:00:00Z",
  "expiry_date": "2026-01-01T00:00:00Z",
  "features": ["threat_intel", "apt_detection", "encryption_analysis"]
}
```

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "invalid request parameters"
}
```

### 401 Unauthorized
```json
{
  "error": "missing authorization header"
}
```

### 403 Forbidden
```json
{
  "error": "insufficient permissions"
}
```

### 404 Not Found
```json
{
  "error": "resource not found"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate limit exceeded",
  "retry_after": 30.5
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Rate Limiting

API requests are rate-limited to **100 requests per minute** per IP address. Exceeding this limit will result in HTTP 429 responses.

---

## Metrics

Prometheus metrics are exposed at `/metrics` endpoint (no authentication required).

**Example metrics:**
- `nta_http_requests_total` - Total HTTP requests
- `nta_alerts_total` - Total alerts generated  
- `nta_active_probes` - Number of active probes
- `nta_packets_processed_total` - Total packets processed

---

## Versioning

The API uses URL versioning (`/api/v1`). Breaking changes will be released under a new version (`/api/v2`).

---

## Support

For API support, contact: support@nta.example.com
