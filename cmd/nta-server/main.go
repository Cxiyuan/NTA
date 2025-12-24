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
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/pkg/notification"
	"github.com/Cxiyuan/NTA/pkg/pcap"
	"github.com/Cxiyuan/NTA/pkg/report"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	configFile = flag.String("config", "/opt/nta-probe/config/nta.yaml", "Configuration file path")
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
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Warnf("Failed to load config file, using defaults: %v", err)
		cfg = config.DefaultConfig()
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