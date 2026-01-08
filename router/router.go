package router

import (
	"encoding/json"
	"log"
	"net/http"

	"ollama2openai/config"
	"ollama2openai/ollama"
)

// Global client and config instances
var globalClient *ollama.Client
var globalConfig *config.Config

// SetupRoutes initializes all routes
func SetupRoutes(mux *http.ServeMux, cfg *config.Config) {
	// Store config globally for alias lookup
	globalConfig = cfg

	// Create Ollama client
	globalClient = ollama.NewClient(cfg.OllamaURL, cfg.GetTimeout())

	// Health check
	mux.HandleFunc("/health", HealthHandler)

	// Usage statistics
	mux.HandleFunc("/usage", func(w http.ResponseWriter, r *http.Request) {
		UsageHandler(w, r, cfg)
	})

	// OpenAI-compatible endpoints
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		ChatHandler(w, r, cfg, globalClient)
	})

	mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		EmbeddingHandler(w, r, cfg, globalClient)
	})

	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		ModelsHandler(w, r, cfg, globalClient)
	})

	mux.HandleFunc("/v1/models/", func(w http.ResponseWriter, r *http.Request) {
		ModelHandler(w, r, cfg, globalClient)
	})

	mux.HandleFunc("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		ResponseHandler(w, r, cfg, globalClient)
	})

	log.Printf("Routes configured successfully")
}

// writeError writes an error response in OpenAI format
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

// getAliasFromRequest extracts the API key alias from the request
func getAliasFromRequest(r *http.Request) string {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return "unknown"
	}

	apiKey := authHeader[7:]

	// Look up alias using config
	if globalConfig != nil {
		if alias := globalConfig.GetAlias(apiKey); alias != "" {
			return alias
		}
	}

	return "default"
}
