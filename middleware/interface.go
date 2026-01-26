package middleware

// UsageTracker defines the interface for tracking API usage statistics
// This allows for different implementations (in-memory, Redis, database, etc.)
type UsageTracker interface {
	// RecordCompletion records token usage for a completion request
	RecordCompletion(alias string, promptTokens, completionTokens int64)

	// RecordEmbedding records token usage for an embedding request
	RecordEmbedding(alias string, tokens int64)

	// GetStats returns all usage statistics
	GetStats() map[string]*UsageRecord

	// Reset resets all statistics (useful for testing)
	Reset()
}

// Ensure UsageStats implements UsageTracker
var _ UsageTracker = (*UsageStats)(nil)
