package api

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/adevcorn/ensemble/internal/server/metrics"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// requestIDKey is the context key for request ID
type contextKey string

const requestIDKey contextKey = "request_id"

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request
		duration := time.Since(start)
		requestID := getRequestID(r.Context())

		log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP request")

		metrics.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration.Seconds())
	})
}

// recoveryMiddleware recovers from panics and returns 500 error
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestID(r.Context())

				log.Error().
					Str("request_id", requestID).
					Interface("error", err).
					Str("stack", string(debug.Stack())).
					Msg("Panic recovered")

				respondError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requestIDMiddleware adds a unique request ID to the context
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getRequestID extracts the request ID from context
func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper function to ensure status code is captured even if WriteHeader is not called
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// Error wraps an error response
type Error struct {
	Error string `json:"error"`
}

// respondError sends an error response
func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, Error{Error: message})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		if err := writeJSON(w, data); err != nil {
			log.Error().Err(err).Msg("Failed to write JSON response")
		}
	}
}

// writeJSON writes JSON data to the response writer
func writeJSON(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
