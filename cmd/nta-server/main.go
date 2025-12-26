package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cxiyuan/NTA/internal/analyzer"
	"github.com/Cxiyuan/NTA/internal/api"
	"github.com/Cxiyuan/NTA/internal/asset"
	"github.com/Cxiyuan/NTA/internal/audit"
	"github.com/Cxiyuan/NTA/internal/config"
	"github.com/Cxiyuan/NTA/internal/license"
	"github.com/Cxiyuan/NTA/internal/probe"
	"github.com/Cxiyuan/NTA/internal/threatintel"
	"github.com/Cxiyuan/NTA/internal/zeek"
	"github.com/Cxiyuan/NTA/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"github.com/Cxiyuan/NTA/pkg/notification"
	"github.com/Cxiyuan/NTA/pkg/pcap"
	"github.com/Cxiyuan/NTA/pkg/report"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	configFile = flag.String("config", "/app/config/nta.yaml", "Configuration file path")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
)

func main() {
	flag.Parse()

	// Setup logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.Info("Starting NTA Server...")

	// Load configuration
	logger.Infof("Loading configuration from: %s", *configFile)
	
	var cfg *config.Config
	
	// Check if config file exists
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		logger.Errorf("Configuration file not found: %s", *configFile)
		logger.Warnf("Using default configuration (this will use localhost for database)")
		cfg = config.DefaultConfig()
	} else {
		cfg, err = config.LoadConfig(*configFile)
		if err != nil {
			logger.Warnf("Failed to load config file: %v", err)
			logger.Warnf("Using default configuration")
			cfg = config.DefaultConfig()
		} else {
			logger.Infof("Configuration loaded successfully")
			logger.Infof("Database DSN: %s", cfg.Database.DSN)
			logger.Infof("Redis Address: %s", cfg.Redis.Addr)
		}
	}

	// Initialize database
	var db *gorm.DB
	
	db, err = gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	db.AutoMigrate(
		&models.Alert{},
		&models.Asset{},
		&models.ThreatIntel{},
		&models.Probe{},
		&models.ZeekProbe{},
		&models.ZeekLog{},
		&models.APTIndicator{},
		&models.AuditLog{},
		&models.Tenant{},
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.Report{},
		&models.NotificationConfig{},
		&models.PCAPSession{},
	)

	// Initialize default admin user if not exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		logger.Info("No users found, creating default admin user")
		
		// Create default tenant
		defaultTenant := models.Tenant{
			TenantID:    "default",
			Name:        "Default",
			Description: "Default tenant",
			Status:      models.StatusActive,
			MaxProbes:   100,
			MaxAssets:   10000,
		}
		if err := db.Create(&defaultTenant).Error; err != nil {
			logger.Errorf("Failed to create default tenant: %v", err)
		} else {
			logger.Infof("Created default tenant (ID: %s)", defaultTenant.TenantID)
		}
		
		// Create admin role
		adminRole := models.Role{
			Name:        models.RoleAdmin,
			Description: "Administrator role with full permissions",
			Permissions: "*:*", // All permissions
		}
		if err := db.Create(&adminRole).Error; err != nil {
			logger.Errorf("Failed to create admin role: %v", err)
		} else {
			logger.Infof("Created admin role (ID: %d)", adminRole.ID)
		}
		
		// Create analyst role
		analystRole := models.Role{
			Name:        models.RoleAnalyst,
			Description: "Analyst role with analysis permissions",
			Permissions: "alerts:read,alerts:update,assets:read,threats:read,reports:generate",
		}
		if err := db.Create(&analystRole).Error; err != nil {
			logger.Errorf("Failed to create analyst role: %v", err)
		}
		
		// Create viewer role
		viewerRole := models.Role{
			Name:        models.RoleViewer,
			Description: "Viewer role with read-only permissions",
			Permissions: "alerts:read,assets:read,threats:read,reports:read",
		}
		if err := db.Create(&viewerRole).Error; err != nil {
			logger.Errorf("Failed to create viewer role: %v", err)
		}
		
		// Create default admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("Failed to hash password: %v", err)
		} else {
			adminUser := models.User{
				Username:     "admin",
				Email:        "admin@nta.local",
				PasswordHash: string(hashedPassword),
				TenantID:     defaultTenant.TenantID,
				Status:       models.StatusActive,
			}
			if err := db.Create(&adminUser).Error; err != nil {
				logger.Errorf("Failed to create admin user: %v", err)
			} else {
				logger.Infof("Created admin user (username: admin, password: admin, hash: %s)", string(hashedPassword))
				
				// Assign admin role to admin user
				userRole := models.UserRole{
					UserID:   adminUser.ID,
					RoleID:   adminRole.ID,
					TenantID: defaultTenant.TenantID,
				}
				if err := db.Create(&userRole).Error; err != nil {
					logger.Errorf("Failed to assign admin role: %v", err)
				} else {
					logger.Info("Assigned admin role to admin user")
				}
			}
		}
		
		logger.Info("✓ Database initialization completed successfully")
		logger.Warn("IMPORTANT: Please change the default admin password after first login!")
	}

	// Initialize builtin zeek probe if not exists
	var zeekProbeCount int64
	db.Model(&models.ZeekProbe{}).Where("probe_id = ?", "builtin-zeek").Count(&zeekProbeCount)
	if zeekProbeCount == 0 {
		logger.Info("Initializing builtin zeek probe")
		builtinProbe := models.ZeekProbe{
			ProbeID:        "builtin-zeek",
			Name:           "内置探针",
			Interface:      "eth0",
			BPFFilter:      "",
			ScriptsEnabled: "[]",
			Status:         "stopped",
		}
		if err := db.Create(&builtinProbe).Error; err != nil {
			logger.Errorf("Failed to create builtin zeek probe: %v", err)
		} else {
			logger.Info("Created builtin zeek probe")
		}
	}

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize license service
	licenseService, err := license.NewService(
		cfg.License.LicenseFile,
		cfg.License.PublicKeyFile,
		logger,
	)
	if err != nil {
		logger.Warnf("Failed to load license: %v", err)
	} else {
		if err := licenseService.Verify(); err != nil {
			logger.Fatalf("License verification failed: %v", err)
		}
	}

	// Initialize services
	assetScanner := asset.NewScanner(db, logger)
	
	threatIntelSources := make([]threatintel.Source, 0)
	for _, src := range cfg.ThreatIntel.Sources {
		threatIntelSources = append(threatIntelSources, threatintel.Source{
			Name:    src.Name,
			URL:     src.URL,
			APIKey:  src.APIKey,
			Enabled: src.Enabled,
		})
	}
	threatIntelService := threatintel.NewService(db, rdb, logger, threatIntelSources)
	
	probeManager := probe.NewManager(db, rdb, logger)
	auditService := audit.NewService(db, logger)
	reportService := report.NewService(db, logger, "/opt/nta-probe/reports")
	notifyService := notification.NewService(db, logger)
	pcapStorage := pcap.NewStorage(db, logger, "/app/pcap")
	zeekManager := zeek.NewManager(db, logger)
	zeekParser := zeek.NewLogParser("/opt/zeek/logs", logger)

	// Cleanup old zeek logs periodically
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := zeekManager.CleanOldLogs(ctx, 30*24*time.Hour); err != nil {
				logger.Errorf("Failed to clean old zeek logs: %v", err)
			}
		}
	}()

	// Initialize analyzers
	lateralDetector := analyzer.NewLateralMovementDetector(
		logger,
		cfg.Detection.Scan.Threshold,
		cfg.Detection.Scan.TimeWindow,
	)

	// Start background tasks
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			lateralDetector.Cleanup()
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			assetScanner.SaveAssets()
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			threatIntelService.UpdateFeeds(ctx)
			threatIntelService.CleanCache()
		}
	}()

	go probeManager.StartHealthCheck(ctx, 30*time.Second)

	// Start asset discovery from traffic
	go func() {
		if err := assetScanner.DiscoverFromTraffic(ctx, cfg.Zeek.Interface); err != nil {
			logger.Errorf("Asset discovery error: %v", err)
		}
	}()

	// Start API server
	apiServer := api.NewServer(
		db,
		logger,
		assetScanner,
		threatIntelService,
		probeManager,
		licenseService,
		auditService,
		reportService,
		notifyService,
		pcapStorage,
		zeekManager,
		cfg.Security.JWTSecret,
	)

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		logger.Infof("Starting API server on %s", addr)
		if err := apiServer.Run(addr); err != nil {
			logger.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down gracefully...")
	time.Sleep(2 * time.Second)
}