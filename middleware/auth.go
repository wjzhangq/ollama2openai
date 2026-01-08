package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"ollama2openai/config"
)

// Context key for alias
type contextKey string

const aliasContextKey contextKey = "alias"

// WithAuth validates API keys from the Authorization header
func WithAuth(handler http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and usage endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/usage" {
			handler.ServeHTTP(w, r)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
			writeError(w, http.StatusUnauthorized, "Missing API key")
			return
		}

		// Check Bearer prefix
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		apiKey := authHeader[7:]

		// Validate API key
		alias := cfg.GetAlias(apiKey)
		if alias == "" {
			writeError(w, http.StatusForbidden, "Invalid API key")
			return
		}

		// Store alias in context for usage tracking
		ctx := context.WithValue(r.Context(), aliasContextKey, alias)
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	})
}

// GetAliasFromContext retrieves the alias from context
func GetAliasFromContext(ctx context.Context) string {
	if alias, ok := ctx.Value(aliasContextKey).(string); ok {
		return alias
	}
	return "unknown"
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "invalid_request_error",
		},
	})
}
