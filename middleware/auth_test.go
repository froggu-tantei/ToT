package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/XEDJK/ToT/auth"
	"github.com/XEDJK/ToT/db/database"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup test environment
	os.Setenv("JWT_SECRET", "test_secret_key")
	defer os.Unsetenv("JWT_SECRET")

	// Create test user and generate valid token
	testUser := database.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	validToken, err := auth.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Test handler that requires authentication
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "No user in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(claims.Username))
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		checkBody      bool
	}{
		{
			name:           "valid_bearer_token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "testuser",
			checkBody:      true,
		},
		{
			name:           "missing_authorization_header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
		{
			name:           "invalid_format_no_bearer",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
		{
			name:           "invalid_format_wrong_prefix",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
		{
			name:           "invalid_token",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
		{
			name:           "empty_bearer_token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
		{
			name:           "malformed_token",
			authHeader:     "Bearer notajwttoken",
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
		{
			name:           "bearer_with_extra_parts",
			authHeader:     "Bearer " + validToken + " extra",
			expectedStatus: http.StatusUnauthorized,
			checkBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()

			// Apply auth middleware and call handler
			AuthMiddleware(testHandler).ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body for successful cases
			if tt.checkBody && w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}

			// For error cases, check that we get JSON response
			if tt.expectedStatus == http.StatusUnauthorized {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected JSON content type for error response, got %q", contentType)
				}
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name           string
		contextValue   interface{}
		expectedOK     bool
		expectedClaims *auth.Claims
	}{
		{
			name: "valid_claims_in_context",
			contextValue: &auth.Claims{
				UserID:   uuid.New(),
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectedOK: true,
		},
		{
			name:         "no_value_in_context",
			contextValue: nil,
			expectedOK:   false,
		},
		{
			name:         "wrong_type_in_context",
			contextValue: "not_claims",
			expectedOK:   false,
		},
		{
			name:         "int_in_context",
			contextValue: 12345,
			expectedOK:   false,
		},
		{
			name:         "map_in_context",
			contextValue: map[string]string{"key": "value"},
			expectedOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.contextValue != nil {
				ctx = context.WithValue(ctx, UserContextKey, tt.contextValue)
			}

			claims, ok := GetUserFromContext(ctx)

			if ok != tt.expectedOK {
				t.Errorf("Expected ok=%v, got ok=%v", tt.expectedOK, ok)
			}

			if tt.expectedOK {
				if claims == nil {
					t.Error("Expected non-nil claims when ok=true")
				} else {
					expectedClaims := tt.contextValue.(*auth.Claims)
					if claims.UserID != expectedClaims.UserID {
						t.Errorf("Expected user ID %v, got %v", expectedClaims.UserID, claims.UserID)
					}
					if claims.Username != expectedClaims.Username {
						t.Errorf("Expected username %q, got %q", expectedClaims.Username, claims.Username)
					}
				}
			} else {
				if claims != nil {
					t.Error("Expected nil claims when ok=false")
				}
			}
		})
	}
}

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		message        string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "bad_request_error",
			statusCode:     http.StatusBadRequest,
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  true,
		},
		{
			name:           "unauthorized_error",
			statusCode:     http.StatusUnauthorized,
			message:        "Unauthorized access",
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  true,
		},
		{
			name:           "internal_server_error",
			statusCode:     http.StatusInternalServerError,
			message:        "Something went wrong",
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  true,
		},
		{
			name:           "empty_message",
			statusCode:     http.StatusBadRequest,
			message:        "",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			respondWithError(w, tt.statusCode, tt.message)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse {
				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected JSON content type, got %q", contentType)
				}

				// Check that response is not empty
				if w.Body.Len() == 0 {
					t.Error("Expected non-empty response body")
				}

				// Check that response contains some form of error indication
				responseBody := w.Body.String()
				if len(responseBody) < 10 {
					t.Error("Expected substantial error response")
				}
			}
		})
	}
}
