package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cxiyuan/NTA/services/detection-service/internal/api"
	"github.com/Cxiyuan/NTA/services/detection-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	port     = flag.Int("port", 8083, "Server port")
	logLevel = flag.String("log-level", "info", "Log level")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	level, _ := logrus.ParseLevel(*logLevel)
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	svc := service.NewDetectionService(logger)
	handler := api.NewDetectionHandler(svc, logger)

	router := gin.Default()
	apiGroup := router.Group("/api/v1")
	handler.RegisterRoutes(apiGroup)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", *port)
	logger.Infof("Starting detection-service on %s", addr)

	go func() {
		if err := router.Run(addr); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down detection-service...")
	_ = context.Background()
}
