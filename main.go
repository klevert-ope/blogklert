package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"blogklert/db"
	"blogklert/handlers"
	"blogklert/middleware"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("Database URL (DB_URL) environment variable is not set")
	}

	err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Load the Bearer token from environment variable
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		log.Fatal("Bearer Token environment variable (BEARER_TOKEN) is not set")
	}

	rateLimiter := middleware.NewRateLimiter()

	route := http.NewServeMux()
	route.Handle("/posts", handlers.SetupPostRoutes())
	route.Handle("/posts/{id}", handlers.SetupPostRoutes())
	route.HandleFunc("/", handlers.RootHandler)

	// Apply middleware to all requests by wrapping the handler with middleware chain
	http.Handle("/",
		middleware.RouteWithMiddleware(
			route,
			middleware.Cors,
			rateLimiter.Limit,
			middleware.ValidateBearerToken(bearerToken),
		),
	)

	srv := &http.Server{
		Addr:           ":8000",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1500, // 1.5 KB
		IdleTimeout:    120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
