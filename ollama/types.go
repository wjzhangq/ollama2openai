package ollama

// Ollama Chat Request - matches Ollama API spec
type ChatRequest struct {
	Model    string       `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool         `json:"stream,omitempty"`
	Format   string       `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	KeepAlive interface{}  `json:"keep_alive,omitempty"`
}

// Ollama Chat Message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Images  []string `json:"images,omitempty"` // Base64 encoded images
}

// Ollama Chat Response
type ChatResponse struct {
	Model              string   `json:"model"`
	CreatedAt          string   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool     `json:"done"`
	TotalDuration      int64    `json:"total_duration,omitempty"`
	LoadDuration       int64    `json:"load_duration,omitempty"`
	PromptEvalCount    int      `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64    `json:"prompt_eval_duration,omitempty"`
	EvalCount          int      `json:"eval_count,omitempty"`
	EvalDuration       int64    `json:"eval_duration,omitempty"`
}

// Ollama Embedding Request
type EmbeddingRequest struct {
	Model  string   `json:"model"`
	Input  string   `json:"input"`
	Options map[string]interface{} `json:"options,omitempty"`
	KeepAlive interface{} `json:"keep_alive,omitempty"`
}

// Ollama Embedding Response
type EmbeddingResponse struct {
	Model     string     `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
	TotalDuration int64  `json:"total_duration,omitempty"`
	LoadDuration  int64  `json:"load_duration,omitempty"`
}

// Ollama Tags Response (for listing models)
type TagsResponse struct {
	Models []ModelInfo `json:"models"`
}

type ModelInfo struct {
	Name       string   `json:"name"`
	Model      string   `json:"model"`
	ModifiedAt string   `json:"modified_at"`
	Size       int64    `json:"size"`
	Digest     string   `json:"digest"`
	Details    *ModelDetails `json:"details,omitempty"`
}

type ModelDetails struct {
	ParentModel   string   `json:"parent_model,omitempty"`
	Format        string   `json:"format"`
	Family        string   `json:"family"`
	Families      []string `json:"families,omitempty"`
	ParameterSize string   `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

// Ollama Generate Request (alternative to chat)
type GenerateRequest struct {
	Model    string   `json:"model"`
	Prompt   string   `json:"prompt"`
	Stream   bool     `json:"stream,omitempty"`
	Format   string   `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	KeepAlive interface{} `json:"keep_alive,omitempty"`
}

// Ollama Generate Response
type GenerateResponse struct {
	Model              string   `json:"model"`
	CreatedAt          string   `json:"created_at"`
	Response           string   `json:"response"`
	Done               bool     `json:"done"`
	TotalDuration      int64    `json:"total_duration,omitempty"`
	LoadDuration       int64    `json:"load_duration,omitempty"`
	PromptEvalCount    int      `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64    `json:"prompt_eval_duration,omitempty"`
	EvalCount          int      `json:"eval_count,omitempty"`
	EvalDuration       int64    `json:"eval_duration,omitempty"`
}
