package middleware

import (
	"github.com/gorilla/csrf"
	"net/http"
)

// Cors Custom middleware to handle CORS
func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "false")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Set CSRF token in the response headers
		csrfToken := csrf.Token(r)
		w.Header().Set("X-CSRF-Token", csrfToken)

		next.ServeHTTP(w, r)
	})
}

// RouteWithMiddleware wraps the given handler with middleware chain
func RouteWithMiddleware(handler http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range mws {
		handler = mw(handler)
	}
	return handler
}
