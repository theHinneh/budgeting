package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/pkg/logger"
	"go.uber.org/zap"
)

// SimpleRateLimiter is a basic rate limiter that tracks requests per IP
type SimpleRateLimiter struct {
	// Map of IP to request count
	requests map[string]int
	// Map of IP to last request time
	lastRequest map[string]time.Time
	// Mutex for thread safety
	mu *sync.Mutex
	// Maximum requests per window
	limit int
	// Time window for rate limiting
	window time.Duration
}

// NewSimpleRateLimiter creates a new simple rate limiter
func NewSimpleRateLimiter(limit int, window time.Duration) *SimpleRateLimiter {
	limiter := &SimpleRateLimiter{
		requests:    make(map[string]int),
		lastRequest: make(map[string]time.Time),
		mu:          &sync.Mutex{},
		limit:       limit,
		window:      window,
	}

	// Start a cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// cleanup periodically removes expired entries
func (s *SimpleRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for ip, lastSeen := range s.lastRequest {
			if now.Sub(lastSeen) > s.window {
				delete(s.requests, ip)
				delete(s.lastRequest, ip)
				logger.Debug("Cleaned up rate limiter entry", zap.String("ip", ip))
			}
		}
		s.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed
func (s *SimpleRateLimiter) Allow(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	lastReq, exists := s.lastRequest[ip]

	// If the IP exists and the window hasn't expired, check the count
	if exists && now.Sub(lastReq) <= s.window {
		if s.requests[ip] >= s.limit {
			return false
		}
		s.requests[ip]++
	} else {
		// Reset the counter for this IP
		s.requests[ip] = 1
	}

	s.lastRequest[ip] = now
	return true
}

// RateLimit returns a middleware that limits request rate based on client IP
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewSimpleRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.Request.RemoteAddr
		}

		if !limiter.Allow(ip) {
			logger.Error("Rate limit exceeded", zap.String("ip", ip))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
