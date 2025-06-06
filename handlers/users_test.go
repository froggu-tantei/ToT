package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/froggu-tantei/ToT/models"
	"github.com/froggu-tantei/ToT/storage"
)

// Simple tests that don't require database
func TestSignupHandlerValidation(t *testing.T) {
	fileStorage := storage.NewLocalStorage("test_uploads", "")
	apiCfg := &APIConfig{
		FileStorage: fileStorage,
		DB:          nil,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing_username",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "testpass123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and username are required", // Updated to match actual handler
		},
		{
			name: "missing_email",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "testpass123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and username are required",
		},
		{
			name: "missing_password",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and username are required",
		},
		{
			name: "bio_too_long",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "testpass123",
				"bio":      strings.Repeat("a", 201), // 201 characters
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bio cannot exceed 200 characters",
		},
		{
			name:           "invalid_json",
			requestBody:    "invalid json string",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer

			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest("POST", "/signup", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiCfg.SignupHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response models.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			if response.Error != tt.expectedError {
				t.Errorf("Expected error %q, got %q", tt.expectedError, response.Error)
			}
		})
	}
}

func TestLoginHandlerValidation(t *testing.T) {
	fileStorage := storage.NewLocalStorage("test_uploads", "")
	apiCfg := &APIConfig{
		FileStorage: fileStorage,
		DB:          nil,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing_email",
			requestBody: map[string]string{
				"password": "testpass123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required", // Changed from "Invalid request format"
		},
		{
			name: "missing_password",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required", // Changed from "Invalid request format"
		},
		{
			name: "empty_fields",
			requestBody: map[string]string{
				"email":    "",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name:           "invalid_json",
			requestBody:    "invalid json string",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer

			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest("POST", "/login", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiCfg.LoginHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response models.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			if response.Error != tt.expectedError {
				t.Errorf("Expected error %q, got %q", tt.expectedError, response.Error)
			}
		})
	}
}
