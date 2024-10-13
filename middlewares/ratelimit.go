package middlewares

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RateLimiter struct {
	limits     sync.Map
	limit      int
	window     time.Duration
	cleanupInt time.Duration
	logLevel   int // 1 for info, 2 for debug
}

type clientData struct {
	requests int32
	timer    *time.Timer
}

// NewRateLimiter initializes a new RateLimiter instance.
func NewRateLimiter(limit int, window time.Duration, cleanupInt time.Duration, logLevel int) *RateLimiter {
	rl := &RateLimiter{
		limit:      limit,
		window:     window,
		cleanupInt: cleanupInt,
		logLevel:   logLevel,
	}

	go rl.cleanup()

	return rl
}

// SetLimit allows dynamic changing of the request limit.
func (rl *RateLimiter) SetLimit(limit int) {
	rl.limit = limit
}

// SetWindow allows dynamic changing of the time window.
func (rl *RateLimiter) SetWindow(window time.Duration) {
	rl.window = window
}

// cleanup periodically removes entries from the rate limiter.
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.cleanupInt)
		rl.limits.Range(func(key, value interface{}) bool {
			data := value.(*clientData)
			if atomic.LoadInt32(&data.requests) == 0 {
				data.timer.Stop()
				rl.limits.Delete(key)
				rl.log("Cleaned up rate limiter entry for client", key)
			}
			return true
		})
	}
}

// getClientIP retrieves the client's IP address.
func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				return ip
			}
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		return ip
	}
	return ""
}

// Limit implements the rate-limiting middleware.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		// Hash the IP address for logging to anonymize it.
		hashedIP := hashIP(clientIP)

		data, _ := rl.limits.LoadOrStore(hashedIP, &clientData{
			requests: 0,
			timer: time.AfterFunc(rl.window, func() {
				rl.resetRequests(hashedIP)
			}),
		})
		clientData := data.(*clientData)

		if atomic.AddInt32(&clientData.requests, 1) > int32(rl.limit) {
			http.Error(w, "You have exceeded the allowed number of requests. Please try again later.", http.StatusTooManyRequests)
			rl.log("Blocked request from client due to rate limiting", hashedIP)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// resetRequests resets the request count for a client.
func (rl *RateLimiter) resetRequests(hashedIP string) {
	data, ok := rl.limits.Load(hashedIP)
	if !ok {
		return
	}
	clientData := data.(*clientData)
	atomic.StoreInt32(&clientData.requests, 0)
	clientData.timer.Reset(rl.window)
}

// hashIP hashes the IP address for anonymization.
func hashIP(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(hash[:])
}

// log logs messages based on the log level.
func (rl *RateLimiter) log(message string, key interface{}) {
	if rl.logLevel != 1 {
		if rl.logLevel == 2 {
			log.Printf(message+": %v (debug)", key) // debug level logging
		}
	} else {
		log.Printf(message+": %v", key) // info level logging
	}
}
