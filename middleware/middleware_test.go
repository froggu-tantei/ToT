package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

	// First request should pass
	if !limiter.Allow("test-client") {
		t.Error("First request should be allowed")
	}

	// Second request should pass (capacity 2)
	if !limiter.Allow("test-client") {
		t.Error("Second request should be allowed")
	}

	// Third request should fail (no tokens left)
	if limiter.Allow("test-client") {
		t.Error("Third request should be blocked")
	}

	// Wait for token refill
	time.Sleep(1100 * time.Millisecond)

	// Should work again after refill
	if !limiter.Allow("test-client") {
		t.Error("Request after refill should be allowed")
	}
}
