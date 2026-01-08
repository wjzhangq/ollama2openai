package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an Ollama API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Ollama client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat sends a chat completion request to Ollama
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/api/chat", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// ChatStream is a channel that receives streaming chat responses
type ChatStream struct {
	responses chan ChatResponse
	err       error
	done      chan struct{}
}

// Chat sends a streaming chat request to Ollama and returns a stream reader
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error) {
	url := fmt.Sprintf("%s/api/chat", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	stream := &ChatStream{
		responses: make(chan ChatResponse, 10),
		done:      make(chan struct{}),
	}

	go func() {
		defer close(stream.responses)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chatResp ChatResponse
			if err := decoder.Decode(&chatResp); err != nil {
				if err == io.EOF || ctx.Err() != nil {
					return
				}
				stream.err = err
				return
			}
			select {
			case stream.responses <- chatResp:
			case <-ctx.Done():
				return
			case <-stream.done:
				return
			}
			if chatResp.Done {
				return
			}
		}
	}()

	return stream, nil
}

// ReadResponse reads a single response from the stream
func (s *ChatStream) ReadResponse() (ChatResponse, error) {
	resp, ok := <-s.responses
	if !ok {
		return ChatResponse{}, s.err
	}
	return resp, nil
}

// Close closes the stream
func (s *ChatStream) Close() {
	close(s.done)
}

// Embedding sends an embedding request to Ollama
func (c *Client) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	url := fmt.Sprintf("%s/api/embeddings", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &embedResp, nil
}

// Tags lists all available models
func (c *Client) Tags(ctx context.Context) (*TagsResponse, error) {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	var tagsResp TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tagsResp, nil
}

// Generate sends a generate request to Ollama (non-streaming)
func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	url := fmt.Sprintf("%s/api/generate", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &genResp, nil
}
