package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	ExpiryTime        time.Duration
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	config   RateLimitConfig
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		config:   config,
	}

	// Start cleanup routine
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.config.ExpiryTime {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// RateLimit middleware factory
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	rateLimiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rateLimiter.getVisitor(ip)
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error":       "Too many requests",
				"retry_after": "1s",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
