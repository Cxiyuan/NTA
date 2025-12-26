package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Cxiyuan/NTA/internal/kafka"
	"github.com/Cxiyuan/NTA/internal/threatintel"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	kafkaBrokers = flag.String("kafka-brokers", getEnv("KAFKA_BROKERS", "localhost:9092"), "Kafka broker addresses")
	redisAddr    = flag.String("redis-addr", getEnv("REDIS_ADDR", "localhost:6379"), "Redis address")
	postgresHost = flag.String("pg-host", getEnv("POSTGRES_HOST", "localhost"), "PostgreSQL host")
	postgresPort = flag.String("pg-port", getEnv("POSTGRES_PORT", "5432"), "PostgreSQL port")
	postgresDB   = flag.String("pg-db", getEnv("POSTGRES_DB", "nta"), "PostgreSQL database")
	postgresUser = flag.String("pg-user", getEnv("POSTGRES_USER", "nta"), "PostgreSQL user")
	postgresPass = flag.String("pg-pass", getEnv("POSTGRES_PASSWORD", "nta_password"), "PostgreSQL password")
	logLevel     = flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Log level")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	level, _ := logrus.ParseLevel(*logLevel)
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	dsn := "host=" + *postgresHost +
		" port=" + *postgresPort +
		" user=" + *postgresUser +
		" password=" + *postgresPass +
		" dbname=" + *postgresDB +
		" sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	logger.Info("Connected to PostgreSQL")

	rdb := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Info("Connected to Redis")

	threatIntelService := threatintel.NewService(db, rdb, logger, []threatintel.Source{})

	brokers := strings.Split(*kafkaBrokers, ",")

	consumerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topics := []string{
		"zeek-conn",
		"zeek-dns",
		"zeek-http",
		"zeek-ssl",
		"zeek-notice",
	}

	for _, topic := range topics {
		consumer := kafka.NewConsumer(brokers, topic, "nta-consumer-group", db, logger, threatIntelService)
		go func(t string, c *kafka.Consumer) {
			logger.Infof("Starting consumer for topic: %s", t)
			if err := c.Start(consumerCtx); err != nil {
				logger.Errorf("Consumer error for topic %s: %v", t, err)
			}
		}(topic, consumer)
	}

	logger.Info("Kafka consumers started successfully")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	logger.Info("Shutting down consumers...")
	cancel()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}