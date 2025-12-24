package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string]*ClientLimit
	mu       sync.RWMutex
	maxReqs  int
	window   time.Duration
}

type ClientLimit struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*ClientLimit),
		maxReqs:  maxRequests,
		window:   window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		client, exists := rl.requests[clientIP]

		if !exists || now.After(client.resetTime) {
			rl.requests[clientIP] = &ClientLimit{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			c.Next()
			return
		}

		if client.count >= rl.maxReqs {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": client.resetTime.Sub(now).Seconds(),
			})
			c.Abort()
			return
		}

		client.count++
		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.requests {
			if now.After(client.resetTime.Add(1 * time.Minute)) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}
