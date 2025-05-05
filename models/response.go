// models/response.go
package models

// SuccessResponse wraps successful responses with metadata
type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

// ErrorResponse provides consistent error format
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// NewSuccessResponse creates a standard success response
func NewSuccessResponse(data any) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates a standard error response
func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error:   message,
	}
}
