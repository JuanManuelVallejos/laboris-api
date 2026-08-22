package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// bucket is a simple token bucket: tokens refill continuously at rate
// tokensPerSecond up to max, and each allowed request consumes one.
type bucket struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	refill   float64
	lastSeen time.Time
}

func (b *bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.lastSeen = now

	b.tokens += elapsed * b.refill
	if b.tokens > b.max {
		b.tokens = b.max
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// ipLimiter hands out one token bucket per client IP, sweeping buckets that
// have been idle for a while so the map doesn't grow unbounded on a
// long-running process.
type ipLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	max     float64
	refill  float64
}

func newIPLimiter(requestsPerSecond, burst float64) *ipLimiter {
	l := &ipLimiter{
		buckets: make(map[string]*bucket),
		max:     burst,
		refill:  requestsPerSecond,
	}
	go l.cleanupLoop()
	return l
}

func (l *ipLimiter) cleanupLoop() {
	for {
		time.Sleep(10 * time.Minute)
		cutoff := time.Now().Add(-30 * time.Minute)
		l.mu.Lock()
		for ip, b := range l.buckets {
			b.mu.Lock()
			stale := b.lastSeen.Before(cutoff)
			b.mu.Unlock()
			if stale {
				delete(l.buckets, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	b, ok := l.buckets[ip]
	if !ok {
		b = &bucket{tokens: l.max, max: l.max, refill: l.refill, lastSeen: time.Now()}
		l.buckets[ip] = b
	}
	l.mu.Unlock()
	return b.allow()
}

// RateLimit caps each client IP to requestsPerSecond sustained requests,
// allowing short bursts up to burst. Keyed by IP (not by authenticated
// user) since it needs to protect endpoints before/without any identity in
// scope — this is basic abuse protection for an MVP, not a precision tool.
func RateLimit(requestsPerSecond, burst float64) gin.HandlerFunc {
	l := newIPLimiter(requestsPerSecond, burst)
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
