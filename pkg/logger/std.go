package logger

import (
	"fmt"
	"log"
	"strings"
)

// Level represents the log level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// ParseLevel parses a string into a Level
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// StdLogger is a Logger implementation using the standard library log package
type StdLogger struct {
	level  Level
	fields []Field
}

// NewStdLogger creates a new standard library logger
func NewStdLogger(level Level) *StdLogger {
	return &StdLogger{
		level:  level,
		fields: make([]Field, 0),
	}
}

// Debug logs a debug message
func (l *StdLogger) Debug(msg string, fields ...Field) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, fields...)
	}
}

// Info logs an info message
func (l *StdLogger) Info(msg string, fields ...Field) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, fields...)
	}
}

// Warn logs a warning message
func (l *StdLogger) Warn(msg string, fields ...Field) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, fields...)
	}
}

// Error logs an error message
func (l *StdLogger) Error(msg string, fields ...Field) {
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, fields...)
	}
}

// With returns a new logger with the given fields
func (l *StdLogger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)
	return &StdLogger{
		level:  l.level,
		fields: newFields,
	}
}

// log is the internal logging function
func (l *StdLogger) log(level, msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	if len(allFields) == 0 {
		log.Printf("[%s] %s", level, msg)
		return
	}

	// Format fields as key=value pairs
	var fieldStrs []string
	for _, f := range allFields {
		fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", f.Key, f.Value))
	}

	log.Printf("[%s] %s | %s", level, msg, strings.Join(fieldStrs, " "))
}

// Ensure StdLogger implements Logger
var _ Logger = (*StdLogger)(nil)
