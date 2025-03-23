package server

import (
	"net/http"
)

const (
	// APIKeyHeader is the header name for the API key
	APIKeyHeader = "X-Ticketfu-Key"
)

// APIKeyMiddleware creates a middleware that validates the API key in the request header
func APIKeyMiddleware(apiKey string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			requestKey := r.Header.Get(APIKeyHeader)

			// Validate API key
			if requestKey == "" || requestKey != apiKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid or missing API key"}`))
				return
			}

			// API key is valid, proceed with the request
			next.ServeHTTP(w, r)
		}
	}
}
