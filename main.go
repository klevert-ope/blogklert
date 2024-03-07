package main

import (
	"blogklert/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"

	"blogklert/handlers"
	"blogklert/middleware"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("Database URL (DB_URL) environment variable is not set")
	}

	// Load the Bearer token from environment variable
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		log.Fatal("Bearer Token environment variable (BEARER_TOKEN) is not set")
	}

	storageAccount := os.Getenv("STORAGE_ACCOUNT_ENDPOINT")
	if storageAccount == "" {
		log.Fatal("Storage Account endpoint environment variable (STORAGE_ACCOUNT_ENDPOINT) is not set")
	}

	// Load the container name from environment variable
	containerName := os.Getenv("CONTAINER_NAME")
	if containerName == "" {
		log.Fatal("Container Name environment variable (CONTAINER_NAME) is not set")
	}

	// Load the CDN endpoint URL from environment variable
	cdnEndpoint := os.Getenv("CDN_ENDPOINT_URL")
	if cdnEndpoint == "" {
		log.Fatal("CDN Endpoint URL environment variable (CDN_ENDPOINT_URL) is not set")
	}

	err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	router := mux.NewRouter()
	handlers.SetupRootRoute(router)
	handlers.SetupPostRoutes(router)
	handlers.SetupUploadRoute(router, storageAccount, cdnEndpoint, containerName)

	// Apply middleware to all requests by wrapping the handler with middleware chain
	http.Handle("/",
		middleware.RouteWithMiddleware(
			router,
			middleware.Cors,
			middleware.NewRateLimiter().Limit,
			middleware.ValidateBearerToken(bearerToken),
		),
	)

	srv := &http.Server{
		Addr:           ":8000",
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1500, // 1.5 KB
		IdleTimeout:    120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
