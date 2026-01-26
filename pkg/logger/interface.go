package logger

// Logger defines the interface for structured logging
// This allows for different logging implementations (stdlib, zap, zerolog, etc.)
type Logger interface {
	// Debug logs a debug message with optional fields
	Debug(msg string, fields ...Field)

	// Info logs an info message with optional fields
	Info(msg string, fields ...Field)

	// Warn logs a warning message with optional fields
	Warn(msg string, fields ...Field)

	// Error logs an error message with optional fields
	Error(msg string, fields ...Field)

	// With returns a new logger with the given fields added to all log messages
	With(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions to create fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
