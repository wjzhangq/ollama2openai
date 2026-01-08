package openai

// Chat Completion Request - matches OpenAI API spec
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatMessage           `json:"messages"`
	MaxTokens        *int                    `json:"max_tokens,omitempty"`
	Temperature      *float64                `json:"temperature,omitempty"`
	TopP             *float64                `json:"top_p,omitempty"`
	N                *int                    `json:"n,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Stop             interface{}             `json:"stop,omitempty"`
	PresencePenalty  *float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int          `json:"logit_bias,omitempty"`
	User             string                  `json:"user,omitempty"`
	Tools            []Tool                  `json:"tools,omitempty"`
	ToolChoice       interface{}             `json:"tool_choice,omitempty"`
	ResponseFormat   *ResponseFormat         `json:"response_format,omitempty"`
}

// ChatMessage represents a message in a chat completion
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentPart
	Name    string      `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ContentPart represents a part of message content (for vision/multimodal)
type ContentPart struct {
	Type      string  `json:"type"`
	Text      string  `json:"text,omitempty"`
	ImageURL  *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in message content
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "auto", "low", "high"
}

// Tool represents a tool that can be called
type Tool struct {
	Type     string     `json:"type"`
	Function *ToolFunc  `json:"function,omitempty"`
}

// ToolFunc represents a function tool
type ToolFunc struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function *ToolCallFunction `json:"function,omitempty"`
}

// ToolCallFunction represents a function call in a tool call
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ResponseFormat specifies the response format
type ResponseFormat struct {
	Type string `json:"type"` // "json_object", "text"
}

// Chat Completion Response
type ChatCompletionResponse struct {
	ID                string                `json:"id"`
	Object            string                `json:"object"`
	Created           int64                 `json:"created"`
	Model             string                `json:"model"`
	Choices           []ChatChoice          `json:"choices"`
	Usage             Usage                 `json:"usage"`
	SystemFingerprint string                `json:"system_fingerprint,omitempty"`
}

// ChatChoice represents a choice in a chat completion response
type ChatChoice struct {
	Index        int           `json:"index"`
	Message      ChatMessage   `json:"message"`
	FinishReason string        `json:"finish_reason"`
	Delta        *ChatMessage  `json:"delta,omitempty"` // For streaming
}

// Usage represents token usage in a response
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Embedding Request
type EmbeddingRequest struct {
	Model    string   `json:"model"`
	Input    interface{} `json:"input"` // Can be string or []string
	User     string   `json:"user,omitempty"`
	EncodingFormat *string `json:"encoding_format,omitempty"`
	Dimensions *int     `json:"dimensions,omitempty"`
}

// Embedding Response
type EmbeddingResponse struct {
	Object    string             `json:"object"`
	Data      []EmbeddingData   `json:"data"`
	Model     string             `json:"model"`
	Usage     EmbeddingUsage     `json:"usage"`
}

// EmbeddingData represents a single embedding
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage represents usage for embeddings
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Model represents a model in the OpenAI API
type Model struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	OwnedBy     string `json:"owned_by"`
}

// ModelsResponse represents the response for listing models
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ErrorResponse represents an error from the API
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// StreamChunk represents a streaming chunk (for internal use)
type StreamChunk struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []StreamChoice  `json:"choices"`
}

type StreamChoice struct {
	Index      int        `json:"index"`
	Delta      ChatMessage `json:"delta"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

// Response API types (simplified)
type ResponseRequest struct {
	Model       string         `json:"model"`
	Input       interface{}    `json:"input,omitempty"`
	Tools       []Tool         `json:"tools,omitempty"`
	ToolChoice  interface{}    `json:"tool_choice,omitempty"`
	MaxOutputTokens *int       `json:"max_output_tokens,omitempty"`
	Temperature *float64       `json:"temperature,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
}

type ResponseResponse struct {
	ID          string          `json:"id"`
	Object      string          `json:"object"`
	Created     int64           `json:"created"`
	Model       string          `json:"model"`
	Output      []OutputItem    `json:"output"`
	Usage       Usage           `json:"usage"`
}

type OutputItem struct {
	Type    string        `json:"type"`
	Content []ContentPart `json:"content,omitempty"`
	Role    string        `json:"role,omitempty"`
}
