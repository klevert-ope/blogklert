package handlers

import (
	"fmt"
	"github.com/gorilla/csrf"
	"net/http"
)

// CsrfHandler handles requests to generate crsftoken
func CsrfHandler(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintf(w, `{"csrfToken": "%s"}`, token)
	if err != nil {
		return
	}
}
