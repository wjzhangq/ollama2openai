package ollama

import "context"

// ClientInterface defines the interface for Ollama API operations
// This allows for easy mocking and testing
type ClientInterface interface {
	// Chat sends a non-streaming chat completion request
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream sends a streaming chat completion request
	ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error)

	// Embedding sends an embedding request
	Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// Tags lists all available models
	Tags(ctx context.Context) (*TagsResponse, error)

	// Generate sends a generate request (alternative to chat)
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
