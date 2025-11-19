package models

import (
	"time"
)

// ResponseStatus represents standard response status codes
type ResponseStatus int

const (
	StatusSuccess      ResponseStatus = 200
	StatusBadRequest   ResponseStatus = 400
	StatusUnauthorized ResponseStatus = 401
	StatusNotFound     ResponseStatus = 404
	StatusError        ResponseStatus = 500
)

// StandardAPIResponse represents a standard API response format (IPC_02: 统一响应格式)
type StandardAPIResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message"`
	Data      interface{}    `json:"data,omitempty"`
	Error     *string        `json:"error,omitempty"`
	Code      ResponseStatus `json:"code,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(message string, data interface{}) *StandardAPIResponse {
	return &StandardAPIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Code:      StatusSuccess,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(message string, err error) *StandardAPIResponse {
	errorMsg := message
	if err != nil {
		errorMsg = message + ": " + err.Error()
	}

	return &StandardAPIResponse{
		Success:   false,
		Message:   errorMsg,
		Error:     &errorMsg,
		Code:      StatusError,
		Timestamp: time.Now(),
	}
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(message string) *StandardAPIResponse {
	return &StandardAPIResponse{
		Success:   false,
		Message:   message,
		Error:     &message,
		Code:      StatusBadRequest,
		Timestamp: time.Now(),
	}
}

// NewNotFoundResponse creates a not found response
func NewNotFoundResponse(message string) *StandardAPIResponse {
	return &StandardAPIResponse{
		Success:   false,
		Message:   message,
		Error:     &message,
		Code:      StatusNotFound,
		Timestamp: time.Now(),
	}
}

// NewUnauthorizedResponse creates an unauthorized response
func NewUnauthorizedResponse(message string) *StandardAPIResponse {
	return &StandardAPIResponse{
		Success:   false,
		Message:   message,
		Error:     &message,
		Code:      StatusUnauthorized,
		Timestamp: time.Now(),
	}
}

// PaginatedResponse wraps paginated data with standard format
type PaginatedResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message"`
	Data      interface{}    `json:"data,omitempty"`
	Error     *string        `json:"error,omitempty"`
	Code      ResponseStatus `json:"code,omitempty"`
	Timestamp time.Time      `json:"timestamp"`

	// Pagination specific fields
	Pagination *PaginationParams `json:"pagination,omitempty"`
	Total      int               `json:"total,omitempty"`
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(message string, data interface{}, pagination *PaginationParams, total int) *PaginatedResponse {
	return &PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Total:      total,
		Code:       StatusSuccess,
		Timestamp:  time.Now(),
	}
}
