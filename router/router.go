package router

import (
	"encoding/json"
	"log"
	"net/http"

	"ollama2openai/config"
	"ollama2openai/ollama"
)

// Router encapsulates the dependencies for handling requests
type Router struct {
	client *ollama.Client
	config *config.Config
}

// NewRouter creates a new Router instance
func NewRouter(cfg *config.Config) *Router {
	return &Router{
		client: ollama.NewClient(cfg.OllamaURL, cfg.GetTimeout()),
		config: cfg,
	}
}

// SetupRoutes initializes all routes
func (rt *Router) SetupRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/health", HealthHandler)

	// Usage statistics
	mux.HandleFunc("/usage", func(w http.ResponseWriter, r *http.Request) {
		UsageHandler(w, r, rt.config)
	})

	// OpenAI-compatible endpoints
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		ChatHandler(w, r, rt.config, rt.client)
	})

	mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		EmbeddingHandler(w, r, rt.config, rt.client)
	})

	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		ModelsHandler(w, r, rt.config, rt.client)
	})

	mux.HandleFunc("/v1/models/", func(w http.ResponseWriter, r *http.Request) {
		ModelHandler(w, r, rt.config, rt.client)
	})

	mux.HandleFunc("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		ResponseHandler(w, r, rt.config, rt.client)
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
func getAliasFromRequest(r *http.Request, cfg *config.Config) string {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return "unknown"
	}

	apiKey := authHeader[7:]

	// Look up alias using config
	if cfg != nil {
		if alias := cfg.GetAlias(apiKey); alias != "" {
			return alias
		}
	}

	return "default"
}
