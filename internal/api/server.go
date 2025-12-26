package api

import (
	"net/http"
	"strconv"

	"github.com/Cxiyuan/NTA/internal/asset"
	"github.com/Cxiyuan/NTA/internal/audit"
	"github.com/Cxiyuan/NTA/internal/license"
	"github.com/Cxiyuan/NTA/internal/probe"
	"github.com/Cxiyuan/NTA/internal/threatintel"
	"github.com/Cxiyuan/NTA/internal/zeek"
	"github.com/Cxiyuan/NTA/pkg/middleware"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/pkg/notification"
	"github.com/Cxiyuan/NTA/pkg/pcap"
	"github.com/Cxiyuan/NTA/pkg/report"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Server represents the API server
type Server struct {
	router         *gin.Engine
	db             *gorm.DB
	logger         *logrus.Logger
	assetScanner   *asset.Scanner
	threatIntel    *threatintel.Service
	probeManager   *probe.Manager
	licenseService *license.Service
	auditService   *audit.Service
	authMiddleware *middleware.AuthMiddleware
	reportService  *report.Service
	notifyService  *notification.Service
	pcapStorage    *pcap.Storage
	zeekManager    *zeek.Manager
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
	reportService *report.Service,
	notifyService *notification.Service,
	pcapStorage *pcap.Storage,
	zeekManager *zeek.Manager,
	jwtSecret string,
) *Server {
	router := gin.Default()

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret, logger)

	s := &Server{
		router:         router,
		db:             db,
		logger:         logger,
		assetScanner:   assetScanner,
		threatIntel:    threatIntel,
		probeManager:   probeManager,
		licenseService: licenseService,
		auditService:   auditService,
		authMiddleware: authMiddleware,
		reportService:  reportService,
		notifyService:  notifyService,
		pcapStorage:    pcapStorage,
		zeekManager:    zeekManager,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.RequestLogger(s.logger))
	s.router.Use(middleware.Metrics())
	
	rateLimiter := middleware.NewRateLimiter(100, 60)
	s.router.Use(rateLimiter.Limit())

	s.router.GET("/health", s.healthCheck)
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := s.router.Group("/api/v1")
	
	auth := api.Group("/auth")
	{
		auth.POST("/login", s.login)
		auth.POST("/logout", s.authMiddleware.Authenticate(), s.logout)
		auth.GET("/me", s.authMiddleware.Authenticate(), s.getCurrentUser)
	}

	api.Use(s.authMiddleware.Authenticate())

	assets := api.Group("/assets")
	{
		assets.GET("", s.listAssets)
		assets.GET("/:ip", s.getAsset)
	}

	alerts := api.Group("/alerts")
	{
		alerts.GET("", s.listAlerts)
		alerts.GET("/:id", s.getAlert)
		alerts.PUT("/:id", s.authMiddleware.RequireRole("admin", "analyst"), s.updateAlert)
	}

	threatIntel := api.Group("/threat-intel")
	{
		threatIntel.GET("/check", s.checkThreatIntel)
		threatIntel.POST("/update", s.authMiddleware.RequireRole("admin"), s.updateThreatIntel)
	}

	probes := api.Group("/probes")
	{
		probes.POST("/register", s.registerProbe)
		probes.POST("/:id/heartbeat", s.probeHeartbeat)
		probes.GET("", s.listProbes)
		probes.GET("/:id", s.getProbe)
	}

	audit := api.Group("/audit")
	audit.Use(s.authMiddleware.RequireRole("admin"))
	{
		audit.GET("", s.queryAuditLogs)
	}

	reports := api.Group("/reports")
	{
		reports.GET("", s.listReports)
		reports.POST("/generate", s.authMiddleware.RequireRole("admin", "analyst"), s.generateReport)
		reports.GET("/:id/download", s.downloadReport)
	}

	notifications := api.Group("/notifications")
	notifications.Use(s.authMiddleware.RequireRole("admin"))
	{
		notifications.GET("/config", s.getNotificationConfig)
		notifications.PUT("/config", s.updateNotificationConfig)
		notifications.POST("/test", s.testNotification)
	}

	pcapAPI := api.Group("/pcap")
	{
		pcapAPI.GET("/sessions", s.listPCAPSessions)
		pcapAPI.GET("/:id/download", s.downloadPCAP)
		pcapAPI.POST("/search", s.searchPCAP)
	}

	detection := api.Group("/detection")
	{
		detection.POST("/dga", s.detectDGA)
		detection.POST("/dns-tunnel", s.detectDNSTunnel)
		detection.POST("/c2", s.detectC2)
		detection.POST("/webshell", s.detectWebShell)
	}

	users := api.Group("/users")
	users.Use(s.authMiddleware.RequireRole("admin"))
	{
		users.GET("", s.listUsers)
		users.POST("", s.createUser)
		users.PUT("/:id", s.updateUser)
		users.DELETE("/:id", s.deleteUser)
		users.POST("/:id/reset-password", s.resetUserPassword)
	}

	roles := api.Group("/roles")
	roles.Use(s.authMiddleware.RequireRole("admin"))
	{
		roles.GET("", s.listRoles)
		roles.POST("", s.createRole)
		roles.PUT("/:id", s.updateRole)
		roles.DELETE("/:id", s.deleteRole)
		roles.PUT("/:id/permissions", s.updateRolePermissions)
	}

	tenants := api.Group("/tenants")
	tenants.Use(s.authMiddleware.RequireRole("admin"))
	{
		tenants.GET("", s.listTenants)
		tenants.POST("", s.createTenant)
		tenants.PUT("/:id", s.updateTenant)
		tenants.DELETE("/:id", s.deleteTenant)
		tenants.GET("/:id/users", s.getTenantUsers)
	}

	config := api.Group("/config")
	config.Use(s.authMiddleware.RequireRole("admin"))
	{
		config.GET("", s.getSystemConfig)
		config.PUT("/detection", s.updateDetectionConfig)
		config.PUT("/backup", s.updateBackupConfig)
	}

	builtinProbe := api.Group("/builtin-probe")
	builtinProbe.Use(s.authMiddleware.RequireRole("admin"))
	{
		builtinProbe.GET("", s.getBuiltinProbe)
		builtinProbe.PUT("", s.updateBuiltinProbe)
		builtinProbe.GET("/status", s.getBuiltinProbeStatus)
		builtinProbe.POST("/start", s.startBuiltinProbe)
		builtinProbe.POST("/stop", s.stopBuiltinProbe)
		builtinProbe.POST("/restart", s.restartBuiltinProbe)
		builtinProbe.GET("/scripts", s.getBuiltinProbeScripts)
		builtinProbe.POST("/scripts/:script/enable", s.enableBuiltinProbeScript)
		builtinProbe.POST("/scripts/:script/disable", s.disableBuiltinProbeScript)
		builtinProbe.GET("/logs", s.getBuiltinProbeLogs)
		builtinProbe.GET("/logs/stats", s.getBuiltinProbeLogStats)
		builtinProbe.GET("/interfaces", s.getBuiltinProbeInterfaces)
	}

	api.GET("/license", s.authMiddleware.RequireRole("admin"), s.getLicenseInfo)
	api.POST("/license/upload", s.authMiddleware.RequireRole("admin"), s.uploadLicense)
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
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if pageSize > 100 {
		pageSize = 100
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := s.db.Model(&models.Alert{})
	
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	if err := query.Order("timestamp DESC").Limit(pageSize).Offset(offset).Find(&alerts).Error; err != nil {
		s.logger.Errorf("Failed to query alerts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query alerts"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":      alerts,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
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
		Status string `json:"status" binding:"required,oneof=new investigating resolved false_positive"`
	}
	
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := s.db.Model(&models.Alert{}).Where("id = ?", id).Update("status", update.Status)
	if result.Error != nil {
		s.logger.Errorf("Failed to update alert %s: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update alert"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_alert", id, map[string]interface{}{
		"status": update.Status,
	})

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