package middleware

import (
	"net/http"
)

// RouteWithMiddleware wraps the given handler with middleware chain
func RouteWithMiddleware(handler http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range mws {
		handler = mw(handler)
	}
	return handler
}
