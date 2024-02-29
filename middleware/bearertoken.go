package middleware

import (
	"net/http"
)

func ValidateBearerToken(expectedBearerToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the Bearer token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
				return
			}

			// Check if the Bearer token matches the expected token
			if authHeader != "Bearer "+expectedBearerToken {
				http.Error(w, "Invalid Bearer Token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
