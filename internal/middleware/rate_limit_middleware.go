package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type ipRateLimiter struct {
	mu      sync.RWMutex
	records map[string][]time.Time
	limit   int
	window  time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	limiter := &ipRateLimiter{
		records: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}

	// Background cleaner: evict stale IP entries every window interval
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return limiter
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	timestamps, exists := l.records[ip]
	if !exists {
		l.records[ip] = []time.Time{now}
		return true
	}

	// Filter timestamps within window
	valid := make([]time.Time, 0, len(timestamps)+1)
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.limit {
		l.records[ip] = valid
		return false
	}

	valid = append(valid, now)
	l.records[ip] = valid
	return true
}

func (l *ipRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-l.window)
	for ip, timestamps := range l.records {
		var active []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				active = append(active, t)
			}
		}
		if len(active) == 0 {
			delete(l.records, ip)
		} else {
			l.records[ip] = active
		}
	}
}

// ContactRateLimitMiddleware limits submissions to 5 requests per 5 minutes per IP
func ContactRateLimitMiddleware() gin.HandlerFunc {
	limiter := newIPRateLimiter(5, 5*time.Minute)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !limiter.allow(clientIP) {
			response.Error(c, http.StatusTooManyRequests, "Bạn đã gửi tin nhắn quá thường xuyên. Vui lòng thử lại sau vài phút.", "rate_limit_exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}
