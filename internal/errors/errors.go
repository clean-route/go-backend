package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "NOT_FOUND"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "INTERNAL_ERROR"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "EXTERNAL_ERROR"
	// ErrorTypeUnauthorized represents unauthorized errors
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	// ErrorTypeBadRequest represents bad request errors
	ErrorTypeBadRequest ErrorType = "BAD_REQUEST"
)

// AppError represents a custom application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Err        error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewExternalError creates a new external service error
func NewExternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeExternal,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Err:        err,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// WithContext adds context to an error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError returns the AppError if the error is one, otherwise nil
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with formatted additional context
func Wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// Cause returns the underlying cause of an error
func Cause(err error) error {
	// For simplicity, just return the error itself
	// In a more sophisticated implementation, you might want to unwrap recursively
	return err
}
