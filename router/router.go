package router

import (
	"net/http"

	"ollama2openai/config"
	"ollama2openai/middleware"
	"ollama2openai/ollama"
	"ollama2openai/pkg/errors"
	"ollama2openai/pkg/logger"
)

// Router encapsulates the dependencies for handling requests
type Router struct {
	client ollama.ClientInterface
	config *config.Config
	usage  middleware.UsageTracker
	logger logger.Logger
}

// NewRouter creates a new Router instance
func NewRouter(cfg *config.Config, client ollama.ClientInterface, usage middleware.UsageTracker, log logger.Logger) *Router {
	return &Router{
		client: client,
		config: cfg,
		usage:  usage,
		logger: log,
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
		ChatHandler(w, r, rt.config, rt.client, rt.usage)
	})

	mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		EmbeddingHandler(w, r, rt.config, rt.client, rt.usage)
	})

	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		ModelsHandler(w, r, rt.config, rt.client)
	})

	mux.HandleFunc("/v1/models/", func(w http.ResponseWriter, r *http.Request) {
		ModelHandler(w, r, rt.config, rt.client)
	})

	mux.HandleFunc("/v1/responses", func(w http.ResponseWriter, r *http.Request) {
		ResponseHandler(w, r, rt.config, rt.client, rt.usage)
	})

	rt.logger.Info("Routes configured successfully")
}

// writeError writes an error response using the unified error package
func writeError(w http.ResponseWriter, err *errors.APIError) {
	errors.WriteError(w, err)
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
