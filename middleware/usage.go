package middleware

import (
	"net/http"
	"sync"
	"time"
)

// UsageStats tracks token usage per API key alias
type UsageStats struct {
	mu              sync.RWMutex
	usage           map[string]*UsageRecord
	lastReset       time.Time
}

type UsageRecord struct {
	PromptTokens       int64
	CompletionTokens   int64
	EmbeddingTokens    int64
	TotalRequests      int64
	EmbeddingRequests  int64
}

// Global usage stats instance
var globalStats = NewUsageStats()

// NewUsageStats creates a new usage stats tracker
func NewUsageStats() *UsageStats {
	return &UsageStats{
		usage:     make(map[string]*UsageRecord),
		lastReset: time.Now(),
	}
}

// RecordCompletion records tokens for a completion request
func (s *UsageStats) RecordCompletion(alias string, promptTokens, completionTokens int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.usage[alias]; !ok {
		s.usage[alias] = &UsageRecord{}
	}
	s.usage[alias].PromptTokens += promptTokens
	s.usage[alias].CompletionTokens += completionTokens
	s.usage[alias].TotalRequests++
}

// RecordEmbedding records tokens for an embedding request
func (s *UsageStats) RecordEmbedding(alias string, tokens int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.usage[alias]; !ok {
		s.usage[alias] = &UsageRecord{}
	}
	s.usage[alias].EmbeddingTokens += tokens
	s.usage[alias].EmbeddingRequests++
}

// GetStats returns all usage statistics
func (s *UsageStats) GetStats() map[string]*UsageRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*UsageRecord)
	for k, v := range s.usage {
		result[k] = &UsageRecord{
			PromptTokens:      v.PromptTokens,
			CompletionTokens:  v.CompletionTokens,
			EmbeddingTokens:   v.EmbeddingTokens,
			TotalRequests:     v.TotalRequests,
			EmbeddingRequests: v.EmbeddingRequests,
		}
	}
	return result
}

// WithUsage is middleware for recording usage stats
func WithUsage(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let the handler process the request first
		// Usage will be recorded by the router handlers
		handler.ServeHTTP(w, r)
	})
}

// GetGlobalStats returns the global usage stats
func GetGlobalStats() *UsageStats {
	return globalStats
}

// Reset resets all usage statistics
func (s *UsageStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.usage = make(map[string]*UsageRecord)
	s.lastReset = time.Now()
}
