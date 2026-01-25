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
	"ollama2openai/tokenizer"
	"time"
)

// ResponseHandler handles Response API requests (simplified implementation)
// The Response API is a newer OpenAI API that combines chat, tools, and vision
func ResponseHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, client *ollama.Client) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	var req openai.ResponseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Use default model if not specified
	if req.Model == "" {
		req.Model = defaultModel
	}

	alias := getAliasFromRequest(r, cfg)

	// For now, convert Response API request to chat completion
	// This is a simplified implementation
	messages := extractMessagesFromInput(req.Input)

	// Convert to chat completion request
	chatReq := &openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   req.Stream,
	}

	if req.MaxOutputTokens != nil {
		chatReq.MaxTokens = req.MaxOutputTokens
	}
	if req.Temperature != nil {
		chatReq.Temperature = req.Temperature
	}

	ctx := r.Context()

	if req.Stream {
		// For streaming, we'll redirect to chat handler logic
		ollamaReq, err := convertChatRequest(chatReq)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to convert request: %v", err))
			return
		}
		handleStreamingChat(ctx, w, cfg, client, chatReq, ollamaReq, alias)
		return
	}

	// Non-streaming response
	ollamaReq, err := convertChatRequest(chatReq)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to convert request: %v", err))
		return
	}

	resp, err := client.Chat(ctx, ollamaReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Ollama error: %v", err))
		return
	}

	// Calculate tokens
	promptTokens := estimatePromptTokens(chatReq)
	completionTokens := tokenizer.EstimateTokenCount(resp.Message.Content)

	middleware.GetGlobalStats().RecordCompletion(alias, int64(promptTokens), int64(completionTokens))

	// Build Response API format
	response := openai.ResponseResponse{
		ID:      fmt.Sprintf("resp_%s", generateID()),
		Object:  "response",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Output: []openai.OutputItem{
			{
				Type: "message",
				Content: []openai.ContentPart{
					{
						Type: "text",
						Text: resp.Message.Content,
					},
				},
				Role: "assistant",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractMessagesFromInput(input interface{}) []openai.ChatMessage {
	// Extract messages from input - this is a simplified implementation
	// In the Response API, input can be a string or array of content

	messages := []openai.ChatMessage{}

	switch v := input.(type) {
	case string:
		messages = append(messages, openai.ChatMessage{
			Role:    "user",
			Content: v,
		})
	case []interface{}:
		for _, item := range v {
			if msgMap, ok := item.(map[string]interface{}); ok {
				if content, ok := msgMap["content"].(string); ok {
					role := "user"
					if r, ok := msgMap["role"].(string); ok {
						role = r
					}
					messages = append(messages, openai.ChatMessage{
						Role:    role,
						Content: content,
					})
				}
			}
		}
	}

	return messages
}
