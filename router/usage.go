package router

import (
	"encoding/json"
	"net/http"

	"ollama2openai/config"
	"ollama2openai/middleware"
)

// UsageHandler handles usage statistics requests
func UsageHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats := middleware.GetGlobalStats().GetStats()

	type AliasStats struct {
		PromptTokens      int64 `json:"prompt_tokens"`
		CompletionTokens  int64 `json:"completion_tokens"`
		EmbeddingTokens   int64 `json:"embedding_tokens"`
		TotalRequests     int64 `json:"total_requests"`
		EmbeddingRequests int64 `json:"embedding_requests"`
	}

	result := make(map[string]AliasStats)

	for alias, record := range stats {
		result[alias] = AliasStats{
			PromptTokens:      record.PromptTokens,
			CompletionTokens:  record.CompletionTokens,
			EmbeddingTokens:   record.EmbeddingTokens,
			TotalRequests:     record.TotalRequests,
			EmbeddingRequests: record.EmbeddingRequests,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}
