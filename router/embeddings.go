package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ollama2openai/config"
	"ollama2openai/middleware"
	"ollama2openai/openai"
	"ollama2openai/ollama"
	"ollama2openai/pkg/errors"
	"ollama2openai/tokenizer"
)

// EmbeddingHandler handles embedding requests
func EmbeddingHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, client ollama.ClientInterface, usage middleware.UsageTracker) {
	if r.Method != http.MethodPost {
		writeError(w, errors.ErrMethodNotAllowed)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, errors.ErrInvalidRequest.WithMessage("Failed to read request body"))
		return
	}

	var req openai.EmbeddingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, errors.ErrInvalidRequest.WithMessage(fmt.Sprintf("Invalid request body: %v", err)))
		return
	}

	// Use default model if not specified
	if req.Model == "" {
		req.Model = "nomic-embed-text"
	}

	alias := getAliasFromRequest(r, cfg)

	// Handle both string and array inputs
	inputs := parseEmbeddingInput(req.Input)

	ctx := r.Context()

	var embeddings [][]float64
	totalTokens := 0

	for _, input := range inputs {
		ollamaReq := &ollama.EmbeddingRequest{
			Model:  req.Model,
			Input:  input,
		}

		resp, err := client.Embedding(ctx, ollamaReq)
		if err != nil {
			writeError(w, errors.ErrOllamaConnection.WithMessage(fmt.Sprintf("Ollama error: %v", err)))
			return
		}

		// Extract embeddings
		for _, emb := range resp.Embeddings {
			embeddings = append(embeddings, emb)
		}

		// Estimate tokens for this input
		totalTokens += tokenizer.EstimateTokenCount(input)
	}

	// Record usage
	usage.RecordEmbedding(alias, int64(totalTokens))

	// Build response
	response := openai.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai.EmbeddingData, len(embeddings)),
		Model:  req.Model,
		Usage: openai.EmbeddingUsage{
			PromptTokens: totalTokens,
			TotalTokens:  totalTokens,
		},
	}

	for i, emb := range embeddings {
		response.Data[i] = openai.EmbeddingData{
			Object:    "embedding",
			Embedding: emb,
			Index:     i,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func parseEmbeddingInput(input interface{}) []string {
	switch v := input.(type) {
	case string:
		return []string{v}
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return []string{}
	}
}
