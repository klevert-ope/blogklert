package handlers

import (
	"log"
	"net/http"
)

// RootHandler handles requests to the root path
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Welcome to the root route!")); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
