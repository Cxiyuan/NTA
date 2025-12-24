package health

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Status struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Checker struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewChecker(db *gorm.DB, redis *redis.Client) *Checker {
	return &Checker{
		db:    db,
		redis: redis,
	}
}

func (c *Checker) Check(ctx context.Context) *Status {
	status := &Status{
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Checks:    make(map[string]Check),
	}

	dbCheck := c.checkDatabase(ctx)
	status.Checks["database"] = dbCheck

	redisCheck := c.checkRedis(ctx)
	status.Checks["redis"] = redisCheck

	if dbCheck.Status == "ok" && redisCheck.Status == "ok" {
		status.Status = "healthy"
	} else {
		status.Status = "unhealthy"
	}

	return status
}

func (c *Checker) checkDatabase(ctx context.Context) Check {
	sqlDB, err := c.db.DB()
	if err != nil {
		return Check{Status: "error", Message: err.Error()}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return Check{Status: "error", Message: err.Error()}
	}

	return Check{Status: "ok"}
}

func (c *Checker) checkRedis(ctx context.Context) Check {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := c.redis.Ping(ctx).Err(); err != nil {
		return Check{Status: "error", Message: err.Error()}
	}

	return Check{Status: "ok"}
}
