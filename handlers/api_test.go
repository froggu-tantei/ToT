package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/froggu-tantei/ToT/models"
	"github.com/froggu-tantei/ToT/storage"
)

func TestRootHandler(t *testing.T) {
	// Setup
	fileStorage := storage.NewLocalStorage("test_uploads", "")
	apiCfg := &APIConfig{
		FileStorage: fileStorage,
		// DB is nil for this test since RootHandler doesn't use it
	}

	tests := []struct {
		name               string
		method             string
		expectedStatusCode int
		expectedName       string
		expectedVersion    string
		expectedStatus     string
	}{
		{
			name:               "get_root_endpoint",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
			expectedName:       "Throne of Thorns API",
			expectedVersion:    "1.0.0",
			expectedStatus:     "running",
		},
		{
			name:               "post_root_endpoint",
			method:             "POST",
			expectedStatusCode: http.StatusOK,
			expectedName:       "Throne of Thorns API",
			expectedVersion:    "1.0.0",
			expectedStatus:     "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			apiCfg.RootHandler(w, req)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected content type 'application/json', got %q", contentType)
			}

			// Parse and check response
			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			if response["name"] != tt.expectedName {
				t.Errorf("Expected name %q, got %q", tt.expectedName, response["name"])
			}

			if response["version"] != tt.expectedVersion {
				t.Errorf("Expected version %q, got %q", tt.expectedVersion, response["version"])
			}

			if response["status"] != tt.expectedStatus {
				t.Errorf("Expected status %q, got %q", tt.expectedStatus, response["status"])
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"message": "test"}
	RespondWithJSON(w, 200, data)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", contentType)
	}
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()

	RespondWithError(w, 400, "test error")

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestReadinessHandler(t *testing.T) {
	fileStorage := storage.NewLocalStorage("test_uploads", "")
	apiCfg := &APIConfig{FileStorage: fileStorage}

	tests := []struct {
		name               string
		method             string
		expectedStatusCode int
		expectedResult     string
	}{
		{
			name:               "get_readiness",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
			expectedResult:     "ok",
		},
		{
			name:               "post_readiness",
			method:             "POST",
			expectedStatusCode: http.StatusOK,
			expectedResult:     "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/readiness", nil)
			w := httptest.NewRecorder()

			apiCfg.ReadinessHandler(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectedStatusCode, w.Code)
			}

			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			if response["status"] != tt.expectedResult {
				t.Errorf("Expected status %q, got %q", tt.expectedResult, response["status"])
			}
		})
	}
}

func TestHealthzHandler(t *testing.T) {
	fileStorage := storage.NewLocalStorage("test_uploads", "")
	apiCfg := &APIConfig{FileStorage: fileStorage}

	tests := []struct {
		name               string
		method             string
		expectedStatusCode int
	}{
		{
			name:               "get_health_check",
			method:             "GET",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "head_health_check",
			method:             "HEAD",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "post_health_check",
			method:             "POST",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/healthz", nil)
			w := httptest.NewRecorder()

			apiCfg.HealthzHandler(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// For HEAD requests, we shouldn't check body
			if tt.method != "HEAD" {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if response["status"] != "ok" {
					t.Errorf("Expected status 'ok', got %q", response["status"])
				}
			}
		})
	}
}

func TestErrorHandler(t *testing.T) {
	fileStorage := storage.NewLocalStorage("test_uploads", "")
	apiCfg := &APIConfig{
		FileStorage: fileStorage,
		DB:          nil,
	}

	req := httptest.NewRequest("GET", "/v1/err", nil)
	w := httptest.NewRecorder()

	apiCfg.ErrorHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response.Error != "Internal Server Error" {
		t.Errorf("Expected error 'Internal Server Error', got %q", response.Error)
	}
}
