package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cxiyuan/NTA/services/auth-service/internal/api"
	"github.com/Cxiyuan/NTA/services/auth-service/internal/repository"
	"github.com/Cxiyuan/NTA/services/auth-service/internal/service"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	port      = flag.Int("port", 8081, "Server port")
	dbHost    = flag.String("db-host", "postgres", "Database host")
	dbPort    = flag.Int("db-port", 5432, "Database port")
	dbUser    = flag.String("db-user", "nta", "Database user")
	dbPass    = flag.String("db-pass", "nta_password", "Database password")
	dbName    = flag.String("db-name", "auth_db", "Database name")
	jwtSecret = flag.String("jwt-secret", "nta-secret-key", "JWT secret")
	logLevel  = flag.String("log-level", "info", "Log level")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	level, _ := logrus.ParseLevel(*logLevel)
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{}, &models.Tenant{}); err != nil {
		logger.Fatalf("Failed to migrate database: %v", err)
	}

	repo := repository.NewAuthRepository(db)
	svc := service.NewAuthService(repo, logger, *jwtSecret)
	handler := api.NewAuthHandler(svc, logger)

	router := gin.Default()
	apiGroup := router.Group("/api/v1")
	handler.RegisterRoutes(apiGroup)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", *port)
	logger.Infof("Starting auth-service on %s", addr)

	go func() {
		if err := router.Run(addr); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down auth-service...")
	_ = context.Background()
}
