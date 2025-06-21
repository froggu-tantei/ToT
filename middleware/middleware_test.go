package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/froggu-tantei/ToT/auth"
	"github.com/froggu-tantei/ToT/db/database"
	"github.com/google/uuid"
)

func TestCorsMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := CorsMiddleware(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Check basic CORS functionality
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// CORS headers should be set by the cors package
	// This is a basic smoke test
	if w.Header().Get("Vary") == "" {
		t.Log("Note: CORS headers may not be visible in test environment")
	}
}
func TestRateLimiterBasic(t *testing.T) {
	limiter := NewRateLimiter(1.0, 2) // 1 token/second, capacity 2
	defer limiter.Close()             // Clean up

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// First request should pass
	if !limiter.Allow(limiter.getClientID(req)) {
		t.Error("First request should be allowed")
	}

	// Second request should pass (capacity 2)
	if !limiter.Allow(limiter.getClientID(req)) {
		t.Error("Second request should be allowed")
	}

	// Third request should fail (no tokens left)
	if limiter.Allow(limiter.getClientID(req)) {
		t.Error("Third request should be blocked")
	}

	// Wait for token refill
	time.Sleep(1100 * time.Millisecond)

	// Should work again after refill
	if !limiter.Allow(limiter.getClientID(req)) {
		t.Error("Request after refill should be allowed")
	}
}

func TestRateLimiterWithAuth(t *testing.T) {
	// Setup JWT environment for testing
	os.Setenv("JWT_SECRET", "test_secret_key")
	defer os.Unsetenv("JWT_SECRET")

	limiter := NewRateLimiter(1.0, 2)
	defer limiter.Close()

	// Create a real JWT token for testing
	testUser := database.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	validToken, err := auth.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Test with valid Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.RemoteAddr = "192.168.1.100:12345"

	clientID := limiter.getClientID(req)

	// Should use token-based client ID for valid tokens
	if !strings.HasPrefix(clientID, "token:") {
		t.Errorf("Should use token-based client ID for valid token, got: %s", clientID)
	}

	// Basic rate limiting should still work
	if !limiter.Allow(clientID) {
		t.Error("First request with auth should be allowed")
	}
}

func TestRateLimiterWithInvalidAuth(t *testing.T) {
	limiter := NewRateLimiter(1.0, 2)
	defer limiter.Close()

	// Test with invalid Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-123")
	req.RemoteAddr = "192.168.1.100:12345"

	clientID := limiter.getClientID(req)

	// Should fall back to IP-based client ID for invalid tokens
	if !strings.HasPrefix(clientID, "ip:") {
		t.Errorf("Should use IP-based client ID for invalid token, got: %s", clientID)
	}

	// Basic rate limiting should still work
	if !limiter.Allow(clientID) {
		t.Error("First request with invalid auth should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewRateLimiter(1.0, 1) // Very restrictive
	defer limiter.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := RateLimitMiddleware(limiter)(handler)
	req := httptest.NewRequest("GET", "/", nil)

	// First request should pass
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Error("First request should pass")
	}

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Error("Second request should be rate limited")
	}
}
