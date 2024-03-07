package handlers

import (
	"github.com/gorilla/mux"
	"net/http"
)

// SetupRootRoute sets up routes for the application
func SetupRootRoute(router *mux.Router) {
	// Define routes here
	router.HandleFunc("/", RootHandler).Methods("GET")
}

// RootHandler handles requests to the root path
func RootHandler(w http.ResponseWriter, r *http.Request) {
	// Set response status to 200 OK
	w.WriteHeader(http.StatusOK)

	// Write response body
	if _, err := w.Write([]byte("Welcome to the root route!")); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
