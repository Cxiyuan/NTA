package api

import (
	"net/http"

	"github.com/Cxiyuan/NTA/internal/asset"
	"github.com/Cxiyuan/NTA/internal/audit"
	"github.com/Cxiyuan/NTA/internal/license"
	"github.com/Cxiyuan/NTA/internal/probe"
	"github.com/Cxiyuan/NTA/internal/threatintel"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Server represents the API server
type Server struct {
	router        *gin.Engine
	db            *gorm.DB
	logger        *logrus.Logger
	assetScanner  *asset.Scanner
	threatIntel   *threatintel.Service
	probeManager  *probe.Manager
	licenseService *license.Service
	auditService  *audit.Service
}

// NewServer creates a new API server
func NewServer(
	db *gorm.DB,
	logger *logrus.Logger,
	assetScanner *asset.Scanner,
	threatIntel *threatintel.Service,
	probeManager *probe.Manager,
	licenseService *license.Service,
	auditService *audit.Service,
) *Server {
	router := gin.Default()

	s := &Server{
		router:         router,
		db:             db,
		logger:         logger,
		assetScanner:   assetScanner,
		threatIntel:    threatIntel,
		probeManager:   probeManager,
		licenseService: licenseService,
		auditService:   auditService,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")

	// Health check
	s.router.GET("/health", s.healthCheck)

	// Assets
	api.GET("/assets", s.listAssets)
	api.GET("/assets/:ip", s.getAsset)

	// Alerts
	api.GET("/alerts", s.listAlerts)
	api.GET("/alerts/:id", s.getAlert)
	api.PUT("/alerts/:id", s.updateAlert)

	// Threat Intelligence
	api.GET("/threat-intel/check", s.checkThreatIntel)
	api.POST("/threat-intel/update", s.updateThreatIntel)

	// Probes
	api.POST("/probes/register", s.registerProbe)
	api.POST("/probes/:id/heartbeat", s.probeHeartbeat)
	api.GET("/probes", s.listProbes)
	api.GET("/probes/:id", s.getProbe)

	// Audit logs
	api.GET("/audit", s.queryAuditLogs)

	// License
	api.GET("/license", s.getLicenseInfo)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) listAssets(c *gin.Context) {
	assets := s.assetScanner.GetAssets()
	c.JSON(http.StatusOK, assets)
}

func (s *Server) getAsset(c *gin.Context) {
	ip := c.Param("ip")
	
	var asset models.Asset
	if err := s.db.Where("ip = ?", ip).First(&asset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

func (s *Server) listAlerts(c *gin.Context) {
	var alerts []models.Alert
	
	query := s.db.Model(&models.Alert{})
	
	// Filters
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Order("timestamp DESC").Limit(100).Find(&alerts)
	
	c.JSON(http.StatusOK, alerts)
}

func (s *Server) getAlert(c *gin.Context) {
	id := c.Param("id")
	
	var alert models.Alert
	if err := s.db.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

func (s *Server) updateAlert(c *gin.Context) {
	id := c.Param("id")
	
	var update struct {
		Status string `json:"status"`
	}
	
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.Model(&models.Alert{}).Where("id = ?", id).Update("status", update.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) checkThreatIntel(c *gin.Context) {
	iocType := c.Query("type")
	value := c.Query("value")

	var result *models.ThreatIntel
	var err error

	switch iocType {
	case "ip":
		result, err = s.threatIntel.CheckIP(c.Request.Context(), value)
	case "domain":
		result, err = s.threatIntel.CheckDomain(c.Request.Context(), value)
	case "hash":
		result, err = s.threatIntel.CheckHash(c.Request.Context(), value)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) updateThreatIntel(c *gin.Context) {
	if err := s.threatIntel.UpdateFeeds(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) registerProbe(c *gin.Context) {
	var probe models.Probe
	if err := c.ShouldBindJSON(&probe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.probeManager.RegisterProbe(c.Request.Context(), &probe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, probe)
}

func (s *Server) probeHeartbeat(c *gin.Context) {
	probeID := c.Param("id")

	if err := s.probeManager.UpdateHeartbeat(c.Request.Context(), probeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) listProbes(c *gin.Context) {
	probes := s.probeManager.ListProbes()
	c.JSON(http.StatusOK, probes)
}

func (s *Server) getProbe(c *gin.Context) {
	probeID := c.Param("id")

	probe, err := s.probeManager.GetProbe(probeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "probe not found"})
		return
	}

	c.JSON(http.StatusOK, probe)
}

func (s *Server) queryAuditLogs(c *gin.Context) {
	filters := make(map[string]interface{})
	
	if user := c.Query("user"); user != "" {
		filters["user"] = user
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	logs, err := s.auditService.Query(filters, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (s *Server) getLicenseInfo(c *gin.Context) {
	info := s.licenseService.GetInfo()
	c.JSON(http.StatusOK, info)
}

// Run starts the API server
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}