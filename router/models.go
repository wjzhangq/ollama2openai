package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ollama2openai/config"
	"ollama2openai/openai"
	"ollama2openai/ollama"
)

// ModelsHandler handles list models requests
func ModelsHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, client *ollama.Client) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := r.Context()

	resp, err := client.Tags(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get models: %v", err))
		return
	}

	// Convert to OpenAI format
	models := make([]openai.Model, 0, len(resp.Models))

	for _, m := range resp.Models {
		model := openai.Model{
			ID:      m.Name,
			Object:  "model",
			Created: parseTimestamp(m.ModifiedAt),
			OwnedBy: "ollama",
		}
		models = append(models, model)
	}

	response := openai.ModelsResponse{
		Object: "list",
		Data:   models,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ModelHandler handles get model details requests
func ModelHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, client *ollama.Client) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract model name from URL path
	// Path format: /v1/models/{model}
	parts := splitPath(r.URL.Path)
	if len(parts) < 2 {
		writeError(w, http.StatusNotFound, "Model not found")
		return
	}

	modelName := parts[1]

	ctx := r.Context()

	resp, err := client.Tags(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get models: %v", err))
		return
	}

	// Find the model
	for _, m := range resp.Models {
		if m.Name == modelName {
			response := openai.Model{
				ID:      m.Name,
				Object:  "model",
				Created: parseTimestamp(m.ModifiedAt),
				OwnedBy: "ollama",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	writeError(w, http.StatusNotFound, "Model not found")
}

func splitPath(path string) []string {
	// Remove leading slash and split
	path = trimLeadingSlash(path)
	if path == "" {
		return []string{}
	}
	result := []string{}
	current := ""

	for _, c := range path {
		if c == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

func trimLeadingSlash(s string) string {
	if len(s) > 0 && s[0] == '/' {
		return s[1:]
	}
	return s
}

func parseTimestamp(s string) int64 {
	// Parse RFC3339 format timestamp from Ollama
	if s == "" {
		return time.Now().Unix()
	}

	// Try to parse as RFC3339
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// If parsing fails, return current time
		return time.Now().Unix()
	}

	return t.Unix()
}
