package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"ollama2openai/config"
	"ollama2openai/middleware"
	"ollama2openai/openai"
	"ollama2openai/ollama"
	"ollama2openai/tokenizer"
)

const defaultModel = "llama3"

// ChatHandler handles chat completion requests
func ChatHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, client *ollama.Client) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	var req openai.ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Use default model if not specified
	if req.Model == "" {
		req.Model = defaultModel
	}

	// Convert OpenAI request to Ollama format
	ollamaReq, err := convertChatRequest(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to convert request: %v", err))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), cfg.GetTimeout())
	defer cancel()

	// Get alias for usage tracking
	alias := getAliasFromRequest(r, cfg)

	if req.Stream {
		handleStreamingChat(ctx, w, cfg, client, &req, ollamaReq, alias)
		return
	}

	handleNonStreamingChat(ctx, w, cfg, client, &req, ollamaReq, alias)
}

func handleStreamingChat(ctx context.Context, w http.ResponseWriter, cfg *config.Config, client *ollama.Client, req *openai.ChatCompletionRequest, ollamaReq *ollama.ChatRequest, alias string) {
	stream, err := client.ChatStream(ctx, ollamaReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start streaming: %v", err))
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	created := time.Now().Unix()
	chunkID := fmt.Sprintf("chatcmpl-%s", generateID())

	// Accumulate all content for token counting
	var fullContent strings.Builder

	for {
		resp, err := stream.ReadResponse()
		if err != nil {
			break
		}

		// Accumulate content
		if resp.Message.Content != "" {
			fullContent.WriteString(resp.Message.Content)
		}

		// Convert Ollama response to OpenAI format
		chunk := convertToStreamChunk(&resp, req.Model, chunkID, created)

		// Send SSE format
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)

		if resp.Done {
			break
		}
	}

	// Record usage once at the end with accumulated content
	promptTokens := estimatePromptTokens(req)
	completionTokens := tokenizer.EstimateTokenCount(fullContent.String())
	middleware.GetGlobalStats().RecordCompletion(alias, int64(promptTokens), int64(completionTokens))

	// Send [DONE]
	fmt.Fprintf(w, "data: [DONE]\n\n")
}

func handleNonStreamingChat(ctx context.Context, w http.ResponseWriter, cfg *config.Config, client *ollama.Client, req *openai.ChatCompletionRequest, ollamaReq *ollama.ChatRequest, alias string) {
	resp, err := client.Chat(ctx, ollamaReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Ollama error: %v", err))
		return
	}

	// Calculate prompt tokens
	promptTokens := estimatePromptTokens(req)

	// Calculate completion tokens
	completionTokens := tokenizer.EstimateTokenCount(resp.Message.Content)

	// Record usage
	middleware.GetGlobalStats().RecordCompletion(alias, int64(promptTokens), int64(completionTokens))

	// Convert to OpenAI response format
	openaiResp := convertToChatResponse(resp, req.Model, promptTokens, completionTokens)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

func convertChatRequest(req *openai.ChatCompletionRequest) (*ollama.ChatRequest, error) {
	ollamaReq := &ollama.ChatRequest{
		Model:  req.Model,
		Stream: req.Stream,
	}

	// Convert messages
	for _, msg := range req.Messages {
		ollamaMsg := ollama.ChatMessage{
			Role: msg.Role,
		}

		// Handle content
		switch c := msg.Content.(type) {
		case string:
			ollamaMsg.Content = c
		case []interface{}:
			content := buildContentFromParts(c)
			ollamaMsg.Content = content.text
			ollamaMsg.Images = content.images
		}

		ollamaReq.Messages = append(ollamaReq.Messages, ollamaMsg)
	}

	// Handle options
	if len(req.Messages) > 0 {
		options := make(map[string]interface{})

		if req.Temperature != nil {
			options["temperature"] = *req.Temperature
		}
		if req.TopP != nil {
			options["top_p"] = *req.TopP
		}
		if req.MaxTokens != nil {
			options["num_predict"] = *req.MaxTokens
		}

		if len(options) > 0 {
			ollamaReq.Options = options
		}
	}

	return ollamaReq, nil
}

type contentBuilder struct {
	text   string
	images []string
}

func buildContentFromParts(parts []interface{}) contentBuilder {
	builder := contentBuilder{}

	for _, part := range parts {
		partMap, ok := part.(map[string]interface{})
		if !ok {
			continue
		}

		partType, _ := partMap["type"].(string)

		switch partType {
		case "text":
			if text, ok := partMap["text"].(string); ok {
				builder.text += text
			}
		case "image_url":
			if imageURL, ok := partMap["image_url"].(map[string]interface{}); ok {
				if url, ok := imageURL["url"].(string); ok {
					// Extract base64 data from data URL
					if strings.HasPrefix(url, "data:image") {
						parts := strings.Split(url, ",")
						if len(parts) == 2 {
							builder.images = append(builder.images, parts[1])
						}
					}
				}
			}
		}
	}

	return builder
}

func convertToChatResponse(resp *ollama.ChatResponse, model string, promptTokens, completionTokens int) openai.ChatCompletionResponse {
	created := time.Now().Unix()

	finishReason := "stop"
	if resp.Done {
		finishReason = "stop"
	}

	return openai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", generateID()),
		Object:  "chat.completion",
		Created: created,
		Model:   model,
		Choices: []openai.ChatChoice{
			{
				Index: 0,
				Message: openai.ChatMessage{
					Role:    "assistant",
					Content: resp.Message.Content,
				},
				FinishReason: finishReason,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

func convertToStreamChunk(resp *ollama.ChatResponse, model, chunkID string, created int64) openai.StreamChunk {
	finishReason := ""
	if resp.Done {
		finishReason = "stop"
	}

	return openai.StreamChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []openai.StreamChoice{
			{
				Index: 0,
				Delta: openai.ChatMessage{
					Role:    resp.Message.Role,
					Content: resp.Message.Content,
				},
				FinishReason: finishReason,
			},
		},
	}
}

func estimatePromptTokens(req *openai.ChatCompletionRequest) int {
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		m := map[string]interface{}{
			"role": msg.Role,
		}
		m["content"] = msg.Content
		messages[i] = m
	}
	return tokenizer.EstimateMessagesTokenCount(messages)
}

func generateID() string {
	return uuid.New().String()
}
