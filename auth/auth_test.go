package auth

import (
	"os"
	"testing"

	"github.com/froggu-tantei/ToT/db/database"
	"github.com/google/uuid"
)

func TestGenerateToken(t *testing.T) {
	// Setup environment for all tests
	os.Setenv("JWT_SECRET", "test_secret_key")
	os.Setenv("JWT_EXPIRATION", "1h")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRATION")
	}()

	tests := []struct {
		name        string
		user        database.User
		expectError bool
	}{
		{
			name: "valid_user_with_all_fields",
			user: database.User{
				ID:       uuid.New(),
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectError: false,
		},
		{
			name: "valid_user_with_empty_fields",
			user: database.User{
				ID:       uuid.New(),
				Username: "",
				Email:    "",
			},
			expectError: false,
		},
		{
			name: "user_with_special_characters",
			user: database.User{
				ID:       uuid.New(),
				Username: "test@user#123",
				Email:    "test+tag@example.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.user)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if token != "" {
					t.Error("Expected empty token on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if token == "" {
					t.Error("Expected non-empty token")
				}

				// Validate that we can parse the token back
				claims, err := ValidateToken(token)
				if err != nil {
					t.Errorf("Generated token failed validation: %v", err)
				}
				if claims.UserID != tt.user.ID {
					t.Errorf("Expected user ID %v, got %v", tt.user.ID, claims.UserID)
				}
			}
		})
	}
}

func TestGenerateTokenEnvironmentErrors(t *testing.T) {
	// Save original environment
	originalSecret := os.Getenv("JWT_SECRET")
	originalExpiry := os.Getenv("JWT_EXPIRY")

	defer func() {
		// Restore original environment
		os.Setenv("JWT_SECRET", originalSecret)
		os.Setenv("JWT_EXPIRY", originalExpiry)
	}()

	tests := []struct {
		name          string
		jwtSecret     string
		jwtExpiry     string
		expectedError bool
	}{
		{
			name:          "missing_secret",
			jwtSecret:     "",
			jwtExpiry:     "24h",
			expectedError: true,
		},
		{
			name:          "invalid_expiration",
			jwtSecret:     "test-secret",
			jwtExpiry:     "invalid-duration", // This should cause an error
			expectedError: true,
		},
	}

	// Create a mock user for testing
	mockUser := database.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test environment
			os.Setenv("JWT_SECRET", tt.jwtSecret)
			os.Setenv("JWT_EXPIRY", tt.jwtExpiry)

			_, err := GenerateToken(mockUser)

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	// Setup environment
	os.Setenv("JWT_SECRET", "test_secret_key")
	defer os.Unsetenv("JWT_SECRET")

	// Generate a valid token first
	testUser := database.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	validToken, err := GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
		checkClaims bool
	}{
		{
			name:        "valid_token",
			token:       validToken,
			expectError: false,
			checkClaims: true,
		},
		{
			name:        "empty_token",
			token:       "",
			expectError: true,
			checkClaims: false,
		},
		{
			name:        "invalid_format",
			token:       "invalid.token",
			expectError: true,
			checkClaims: false,
		},
		{
			name:        "malformed_jwt",
			token:       "not.a.jwt.token.at.all",
			expectError: true,
			checkClaims: false,
		},
		{
			name:        "random_string",
			token:       "randomstring",
			expectError: true,
			checkClaims: false,
		},
		{
			name:        "jwt_with_wrong_signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectError: true,
			checkClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if claims != nil {
					t.Error("Expected nil claims on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if claims == nil {
					t.Error("Expected non-nil claims")
				}

				if tt.checkClaims && claims != nil {
					if claims.UserID != testUser.ID {
						t.Errorf("Expected user ID %v, got %v", testUser.ID, claims.UserID)
					}
					if claims.Username != testUser.Username {
						t.Errorf("Expected username %q, got %q", testUser.Username, claims.Username)
					}
					if claims.Email != testUser.Email {
						t.Errorf("Expected email %q, got %q", testUser.Email, claims.Email)
					}
				}
			}
		})
	}
}

func TestValidateTokenWithDifferentSecrets(t *testing.T) {
	// Generate token with one secret
	os.Setenv("JWT_SECRET", "original_secret")
	testUser := database.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	token, err := GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name      string
		newSecret string
		setSecret bool
	}{
		{
			name:      "different_secret",
			newSecret: "different_secret",
			setSecret: true,
		},
		{
			name:      "no_secret",
			setSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setSecret {
				os.Setenv("JWT_SECRET", tt.newSecret)
			} else {
				os.Unsetenv("JWT_SECRET")
			}

			claims, err := ValidateToken(token)
			if err == nil {
				t.Error("Expected error when validating with different/no secret")
			}
			if claims != nil {
				t.Error("Expected nil claims when validation fails")
			}
		})
	}
}
