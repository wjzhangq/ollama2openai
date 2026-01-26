package errors

import (
	"encoding/json"
	"net/http"
)

// APIError represents a structured API error
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// ErrorResponse represents the OpenAI-compatible error response format
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail contains the error details
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// Predefined error types
const (
	TypeInvalidRequest = "invalid_request_error"
	TypeAuthentication = "authentication_error"
	TypePermission     = "permission_error"
	TypeNotFound       = "not_found_error"
	TypeRateLimit      = "rate_limit_error"
	TypeServer         = "server_error"
	TypeTimeout        = "timeout_error"
)

// Predefined errors
var (
	ErrInvalidRequest = &APIError{
		Code:       "invalid_request",
		Message:    "Invalid request",
		Type:       TypeInvalidRequest,
		StatusCode: http.StatusBadRequest,
	}

	ErrMissingAPIKey = &APIError{
		Code:       "missing_api_key",
		Message:    "Missing API key",
		Type:       TypeAuthentication,
		StatusCode: http.StatusUnauthorized,
	}

	ErrInvalidAPIKey = &APIError{
		Code:       "invalid_api_key",
		Message:    "Invalid API key",
		Type:       TypeAuthentication,
		StatusCode: http.StatusForbidden,
	}

	ErrModelNotFound = &APIError{
		Code:       "model_not_found",
		Message:    "Model not found",
		Type:       TypeNotFound,
		StatusCode: http.StatusNotFound,
	}

	ErrMethodNotAllowed = &APIError{
		Code:       "method_not_allowed",
		Message:    "Method not allowed",
		Type:       TypeInvalidRequest,
		StatusCode: http.StatusMethodNotAllowed,
	}

	ErrInternalServer = &APIError{
		Code:       "internal_server_error",
		Message:    "Internal server error",
		Type:       TypeServer,
		StatusCode: http.StatusInternalServerError,
	}

	ErrOllamaConnection = &APIError{
		Code:       "ollama_connection_error",
		Message:    "Failed to connect to Ollama",
		Type:       TypeServer,
		StatusCode: http.StatusServiceUnavailable,
	}

	ErrRequestTimeout = &APIError{
		Code:       "request_timeout",
		Message:    "Request timeout",
		Type:       TypeTimeout,
		StatusCode: http.StatusGatewayTimeout,
	}
)

// New creates a new APIError with a custom message
func New(code, message, errType string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Type:       errType,
		StatusCode: statusCode,
	}
}

// WithMessage returns a copy of the error with a custom message
func (e *APIError) WithMessage(message string) *APIError {
	return &APIError{
		Code:       e.Code,
		Message:    message,
		Type:       e.Type,
		StatusCode: e.StatusCode,
	}
}

// WriteError writes an error response in OpenAI format
func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)

	response := ErrorResponse{
		Error: &ErrorDetail{
			Message: err.Message,
			Type:    err.Type,
			Code:    err.Code,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// WriteErrorMessage writes an error response with a custom message
func WriteErrorMessage(w http.ResponseWriter, statusCode int, message string) {
	errType := TypeServer
	switch statusCode {
	case http.StatusBadRequest:
		errType = TypeInvalidRequest
	case http.StatusUnauthorized, http.StatusForbidden:
		errType = TypeAuthentication
	case http.StatusNotFound:
		errType = TypeNotFound
	case http.StatusTooManyRequests:
		errType = TypeRateLimit
	}

	err := &APIError{
		Code:       "error",
		Message:    message,
		Type:       errType,
		StatusCode: statusCode,
	}

	WriteError(w, err)
}
