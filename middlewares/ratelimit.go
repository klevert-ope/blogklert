package middlewares

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter struct to store request counts and manage locking.
type RateLimiter struct {
	limits     map[string]*clientData
	mu         sync.Mutex
	limit      int
	window     time.Duration
	cleanupInt time.Duration
}

type clientData struct {
	sync.Mutex
	requests int
	timer    *time.Timer
}

// NewRateLimiter initializes a RateLimiter with specified request limit and time window.
func NewRateLimiter(limit int, window time.Duration, cleanupInt time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limits:     make(map[string]*clientData),
		limit:      limit,
		window:     window,
		cleanupInt: cleanupInt,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.cleanupInt)
		rl.mu.Lock()
		for ip, data := range rl.limits {
			data.Lock()
			if data.requests == 0 {
				data.timer.Stop()
				delete(rl.limits, ip)
			}
			data.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP address from the request, considering proxies.
func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, the first one is the original client IP.
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Fallback to the original remote address if there's an error.
	}
	return host
}

// Limit middlewares to limit requests based on client IP.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		rl.mu.Lock()
		data, exists := rl.limits[clientIP]
		if !exists {
			data = &clientData{
				requests: 0,
				timer: time.AfterFunc(rl.window, func() {
					rl.resetRequests(clientIP)
				}),
			}
			rl.limits[clientIP] = data
		}
		rl.mu.Unlock()

		data.Lock()
		defer data.Unlock()

		if data.requests >= rl.limit {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		data.requests++
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) resetRequests(clientIP string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if data, exists := rl.limits[clientIP]; exists {
		data.Lock()
		defer data.Unlock()
		data.requests = 0
		// Only reset the timer if it's not already reset by another call.
		if !data.timer.Stop() {
			<-data.timer.C
		}
		data.timer.Reset(rl.window)
	}
}
