package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"ollama2openai/pkg/logger"
)

// RequestIDKey is the context key for request ID
const RequestIDKey contextKey = "request_id"

// RequestID middleware adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// LoggingMiddleware logs HTTP requests with request ID
func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from context
			requestID := GetRequestID(r.Context())

			// Create logger with request context
			reqLogger := log.With(
				logger.String("request_id", requestID),
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("remote_addr", r.RemoteAddr),
			)

			// Log request start
			reqLogger.Info("Request started")

			// Wrap response writer to capture status code
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(lrw, r)

			// Log request completion
			reqLogger.Info("Request completed",
				logger.Int("status", lrw.statusCode),
			)
		})
	}
}

// Recovery middleware recovers from panics and returns a 500 error
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := GetRequestID(r.Context())
					log.Error("Panic recovered",
						logger.String("request_id", requestID),
						logger.Any("error", err),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":{"message":"Internal server error","type":"server_error"}}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
