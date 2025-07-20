package sync

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// ErrorType represents different types of sync errors
type ErrorType int

const (
	ErrorTypeNetwork ErrorType = iota
	ErrorTypeAuth
	ErrorTypePermission
	ErrorTypeQuota
	ErrorTypeConflict
	ErrorTypeValidation
	ErrorTypeTimeout
	ErrorTypeUnknown
)

// SyncError represents a sync operation error with additional context
type SyncError struct {
	Type      ErrorType
	Message   string
	Operation string
	FilePath  string
	Cause     error
	Retryable bool
	Timestamp time.Time
}

func (e *SyncError) Error() string {
	if e.FilePath != "" {
		return fmt.Sprintf("%s operation failed for %s: %s", e.Operation, e.FilePath, e.Message)
	}
	return fmt.Sprintf("%s operation failed: %s", e.Operation, e.Message)
}

func (e *SyncError) Unwrap() error {
	return e.Cause
}

// NewSyncError creates a new sync error with context
func NewSyncError(errType ErrorType, operation, message string, cause error) *SyncError {
	return &SyncError{
		Type:      errType,
		Message:   message,
		Operation: operation,
		Cause:     cause,
		Retryable: isRetryable(errType, cause),
		Timestamp: time.Now(),
	}
}

// NewSyncErrorWithFile creates a sync error with file context
func NewSyncErrorWithFile(errType ErrorType, operation, filePath, message string, cause error) *SyncError {
	err := NewSyncError(errType, operation, message, cause)
	err.FilePath = filePath
	return err
}

// isRetryable determines if an error should be retried
func isRetryable(errType ErrorType, cause error) bool {
	switch errType {
	case ErrorTypeNetwork, ErrorTypeTimeout:
		return true
	case ErrorTypeQuota:
		return false // Don't retry quota errors immediately
	case ErrorTypeAuth:
		return false // Auth errors need manual intervention
	case ErrorTypePermission:
		return false // Permission errors won't resolve automatically
	case ErrorTypeValidation:
		return false // Validation errors are permanent
	case ErrorTypeConflict:
		return true // Conflicts can be resolved
	default:
		// Check specific error types
		if cause != nil {
			return isNetworkError(cause) || isTemporaryError(cause)
		}
		return false
	}
}

// isNetworkError checks if an error is network-related
func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	
	var syscallErr *net.OpError
	if errors.As(err, &syscallErr) {
		return true
	}
	
	// Check for specific syscall errors
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	
	return false
}

// isTemporaryError checks if an error is temporary
func isTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	
	if temp, ok := err.(temporary); ok {
		return temp.Temporary()
	}
	
	return false
}

// ClassifyHTTPError classifies HTTP response errors
func ClassifyHTTPError(statusCode int, operation string, cause error) *SyncError {
	switch statusCode {
	case http.StatusUnauthorized:
		return NewSyncError(ErrorTypeAuth, operation, "Authentication failed", cause)
	case http.StatusForbidden:
		return NewSyncError(ErrorTypePermission, operation, "Permission denied", cause)
	case http.StatusTooManyRequests:
		return NewSyncError(ErrorTypeQuota, operation, "Rate limit exceeded", cause)
	case http.StatusConflict:
		return NewSyncError(ErrorTypeConflict, operation, "Conflict detected", cause)
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return NewSyncError(ErrorTypeTimeout, operation, "Request timeout", cause)
	case http.StatusBadRequest:
		return NewSyncError(ErrorTypeValidation, operation, "Invalid request", cause)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return NewSyncError(ErrorTypeNetwork, operation, "Server error", cause)
	default:
		if statusCode >= 500 {
			return NewSyncError(ErrorTypeNetwork, operation, fmt.Sprintf("Server error: %d", statusCode), cause)
		}
		return NewSyncError(ErrorTypeUnknown, operation, fmt.Sprintf("HTTP error: %d", statusCode), cause)
	}
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts    int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	RetryableTypes []ErrorType
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableTypes: []ErrorType{
			ErrorTypeNetwork,
			ErrorTypeTimeout,
			ErrorTypeConflict,
		},
	}
}

// ShouldRetry determines if an error should be retried based on config
func (rc *RetryConfig) ShouldRetry(err *SyncError, attempt int) bool {
	if attempt >= rc.MaxAttempts {
		return false
	}
	
	if !err.Retryable {
		return false
	}
	
	// Check if error type is in retryable list
	for _, retryableType := range rc.RetryableTypes {
		if err.Type == retryableType {
			return true
		}
	}
	
	return false
}

// GetDelay calculates delay before next retry attempt
func (rc *RetryConfig) GetDelay(attempt int) time.Duration {
	delay := float64(rc.InitialDelay) * pow(rc.BackoffFactor, float64(attempt))
	
	if delay > float64(rc.MaxDelay) {
		return rc.MaxDelay
	}
	
	return time.Duration(delay)
}

// pow is a simple power function for backoff calculation
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// ErrorRecovery provides strategies for recovering from specific errors
type ErrorRecovery struct {
	retryConfig *RetryConfig
}

// NewErrorRecovery creates a new error recovery instance
func NewErrorRecovery(config *RetryConfig) *ErrorRecovery {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &ErrorRecovery{
		retryConfig: config,
	}
}

// HandleError processes an error and determines recovery strategy
func (er *ErrorRecovery) HandleError(err *SyncError, attempt int) (shouldRetry bool, delay time.Duration) {
	if !er.retryConfig.ShouldRetry(err, attempt) {
		return false, 0
	}
	
	delay = er.retryConfig.GetDelay(attempt)
	
	// Special handling for specific error types
	switch err.Type {
	case ErrorTypeQuota:
		// For quota errors, use longer delays
		delay = delay * 5
	case ErrorTypeConflict:
		// For conflicts, shorter delays might be better
		delay = delay / 2
	}
	
	return true, delay
}