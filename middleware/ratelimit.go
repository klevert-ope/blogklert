package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	sync.Mutex
	Requests map[string]int
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		Requests: make(map[string]int),
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.Lock()
		defer rl.Unlock()
		// Assuming each IP address is unique for simplicity
		clientIP := r.RemoteAddr
		// Limiting to 10 requests per minute per IP
		limit := 10
		windowDuration := time.Minute
		// Reset counts after window duration
		if rl.Requests[clientIP] == 0 {
			time.AfterFunc(windowDuration, func() {
				rl.Lock()
				delete(rl.Requests, clientIP)
				rl.Unlock()
			})
		}
		if rl.Requests[clientIP] >= limit {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		// Increment request count for this client
		rl.Requests[clientIP]++
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
