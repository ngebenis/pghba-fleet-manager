package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// loginAttempts tracks failed login attempts per client IP so brute-forcing
// a password (particularly the published default Super Admin credentials
// seeded in database/migrations/001_init.sql) costs an attacker a lockout
// instead of unlimited guesses. Keyed by IP rather than IP+username: the
// username itself is part of what an attacker is guessing.
type loginAttempts struct {
	mu          sync.Mutex
	failures    map[string]int
	windowStart map[string]time.Time
	lockedUntil map[string]time.Time
}

const (
	loginMaxFailures  = 10
	loginWindow       = 15 * time.Minute
	loginLockDuration = 15 * time.Minute
)

var loginLimiter = &loginAttempts{
	failures:    make(map[string]int),
	windowStart: make(map[string]time.Time),
	lockedUntil: make(map[string]time.Time),
}

// LoginRateLimit blocks further attempts from an IP once it has
// loginMaxFailures failed logins within loginWindow, for loginLockDuration.
// It inspects the downstream handler's response status rather than
// requiring AuthHandler.Login to call into it directly, so a lockout
// change here can never accidentally affect how login itself decides
// success/failure.
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		loginLimiter.mu.Lock()
		if until, locked := loginLimiter.lockedUntil[ip]; locked && now.Before(until) {
			retryAfter := until.Sub(now)
			loginLimiter.mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many failed login attempts — try again later",
			})
			return
		}
		loginLimiter.mu.Unlock()

		c.Next()

		status := c.Writer.Status()
		loginLimiter.mu.Lock()
		defer loginLimiter.mu.Unlock()
		if status == http.StatusOK {
			delete(loginLimiter.failures, ip)
			delete(loginLimiter.windowStart, ip)
			delete(loginLimiter.lockedUntil, ip)
			return
		}
		if status == http.StatusUnauthorized {
			start, ok := loginLimiter.windowStart[ip]
			if !ok || now.Sub(start) > loginWindow {
				loginLimiter.windowStart[ip] = now
				loginLimiter.failures[ip] = 0
			}
			loginLimiter.failures[ip]++
			if loginLimiter.failures[ip] >= loginMaxFailures {
				loginLimiter.lockedUntil[ip] = now.Add(loginLockDuration)
			}
		}
	}
}
