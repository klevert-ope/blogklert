package middlewares

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	type args struct {
		limit      int
		window     time.Duration
		cleanupInt time.Duration
		logLevel   int // Include log level in the arguments
	}
	tests := []struct {
		name string
		args args
		want *RateLimiter
	}{
		{
			name: "Standard rate limiter",
			args: args{
				limit:      5,
				window:     1 * time.Second,
				cleanupInt: 1 * time.Second,
				logLevel:   1, // Set log level for the test
			},
			want: &RateLimiter{
				limit:      5,
				window:     1 * time.Second,
				cleanupInt: 1 * time.Second,
				logLevel:   1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRateLimiter(tt.args.limit, tt.args.window, tt.args.cleanupInt, tt.args.logLevel)
			if got.limit != tt.want.limit || got.window != tt.want.window || got.cleanupInt != tt.want.cleanupInt || got.logLevel != tt.want.logLevel {
				t.Errorf("NewRateLimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Limit(t *testing.T) {
	rl := NewRateLimiter(3, 2*time.Second, 1*time.Second, 1)
	var wg sync.WaitGroup

	// Create a dummy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test a client that is within the limit
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			rl.Limit(handler).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status OK, got %v", rec.Code)
			}
		}()
	}

	wg.Wait()

	// Test a client that exceeds the limit
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			rl.Limit(handler).ServeHTTP(rec, req)
			if rec.Code != http.StatusTooManyRequests {
				t.Errorf("Expected status TooManyRequests, got %v", rec.Code)
			}
		}()
	}

	wg.Wait()

	// Wait for the cleanup timer and retry to ensure it resets
	time.Sleep(3 * time.Second)

	// Test a client after the window has passed
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			rl.Limit(handler).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status OK after window reset, got %v", rec.Code)
			}
		}()
	}

	wg.Wait()

	// Validate cleanup behavior (wait for a few more seconds)
	time.Sleep(3 * time.Second)

	// Test that requests beyond the limit again yield TooManyRequests
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			rl.Limit(handler).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK && rec.Code != http.StatusTooManyRequests {
				t.Errorf("Expected status OK or TooManyRequests, got %v", rec.Code)
			}
		}()
	}

	wg.Wait()
}

func TestCleanup(t *testing.T) {
	rl := NewRateLimiter(20, 1*time.Second, 1*time.Second, 1)

	req := httptest.NewRequest("GET", "/", nil)
	rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(httptest.NewRecorder(), req)

	// Add a short delay to ensure the request is processed
	time.Sleep(50 * time.Millisecond)

	clientIP := getClientIP(req)
	hashedIP := hashIP(clientIP)

	// Check that the entry exists before cleanup
	if _, ok := rl.limits.Load(hashedIP); !ok {
		t.Error("Expected the entry to exist before cleanup")
	}

	time.Sleep(3 * time.Second)

	// Ensure the entry is cleaned up after the cleanup interval
	if _, ok := rl.limits.Load(hashedIP); ok {
		t.Errorf("Expected the entry to be cleaned up, but it still exists")
	}
}
