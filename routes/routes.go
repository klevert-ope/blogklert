package routes

import (
	"blogklert/controllers"
	"blogklert/middlewares"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Config interface represents the configuration needed for setting up routes.
type Config interface {
	GetBearerToken() string
}

// SetupRoutes sets up the application routes and middlewares.
func SetupRoutes(config Config) http.Handler {
	router := mux.NewRouter()
	controllers.SetupRootRoute(router)
	controllers.SetupPostRoutes(router)

	// Create a CorsConfig instance
	corsConfig := &middlewares.CorsConfig{
		AllowedOrigins:   []string{"http://0.0.0.0:3000", "http://localhost:8000", "https://www.klevertopee.app", "https://klevert-dev.koyeb.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}

	// Apply Cors middlewares to all requests by wrapping the router
	router.Use(middlewares.CorsMiddleware(corsConfig))

	// Initialize rate limiter with limit, window duration, and cleanup interval
	rateLimiter := middlewares.NewRateLimiter(15, 1*time.Minute, 1*time.Minute, 1)

	// Create the middlewares chain
	middlewareChain := rateLimiter.Limit(middlewares.ValidateBearerToken(config.GetBearerToken())(router))
	middlewareChain = middlewares.LoggingMiddleware(middlewareChain)

	return middlewareChain
}
