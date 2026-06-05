package server

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimiter enforces a per-client-IP request budget using a token bucket per
// IP. The budget is expressed as a number of requests allowed in a rolling
// minute: the bucket holds perMinute tokens and refills at perMinute tokens per
// minute, so the first perMinute requests from an IP succeed immediately and
// further requests are rejected until tokens refill.
//
// Stale per-IP buckets are evicted by a background janitor so memory does not
// grow without bound under churn of distinct client IPs.
type rateLimiter struct {
	limit rate.Limit
	burst int
	ttl   time.Duration

	mu      sync.Mutex
	clients map[string]*rateLimiterClient

	now func() time.Time
}

type rateLimiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newRateLimiter builds a limiter allowing perMinute requests per minute per
// IP. It starts a background janitor that evicts buckets untouched for longer
// than the refill window.
func newRateLimiter(perMinute int) *rateLimiter {
	rl := &rateLimiter{
		limit:   rate.Every(time.Minute / time.Duration(perMinute)),
		burst:   perMinute,
		ttl:     2 * time.Minute,
		clients: make(map[string]*rateLimiterClient),
		now:     time.Now,
	}
	go rl.cleanupLoop()
	return rl
}

// limiterFor returns the token bucket for the given client key, creating one on
// first use and refreshing its lastSeen timestamp.
func (rl *rateLimiter) limiterFor(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, ok := rl.clients[key]
	if !ok {
		c = &rateLimiterClient{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.clients[key] = c
	}
	c.lastSeen = rl.now()
	return c.limiter
}

// Middleware returns chi-compatible middleware that rejects requests exceeding
// the per-IP budget with 429 Too Many Requests and a Retry-After header.
func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.limiterFor(clientIP(r))

		res := limiter.Reserve()
		if delay := res.Delay(); delay > 0 {
			// Token not yet available; return it so the budget is not
			// consumed by a rejected request.
			res.Cancel()
			seconds := int(math.Ceil(delay.Seconds()))
			if seconds < 1 {
				seconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// cleanupLoop periodically drops buckets that have not been used within ttl.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.ttl)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := rl.now().Add(-rl.ttl)
		rl.mu.Lock()
		for key, c := range rl.clients {
			if c.lastSeen.Before(cutoff) {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP extracts the host portion of the request's remote address. RealIP
// middleware upstream has already resolved X-Forwarded-For / X-Real-IP into
// RemoteAddr, so this keys on the genuine client where a proxy is trusted.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
