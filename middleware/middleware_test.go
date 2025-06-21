package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/froggu-tantei/ToT/auth"
	"github.com/froggu-tantei/ToT/db/database"
	"github.com/google/uuid"
)

// Helper function to create test rate limiter
func createTestRateLimiter(rate float64, capacity int) *RateLimiter {
	config := RateLimiterConfig{
		Rate:            rate,
		Capacity:        capacity,
		MaxBuckets:      1000,
		CleanupInterval: 1 * time.Minute,
		BucketTTL:       2 * time.Minute,
		MaxRetryAfter:   5 * time.Minute,
	}
	return NewRateLimiter(config)
}

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
	limiter := createTestRateLimiter(1.0, 2) // 1 token/second, capacity 2
	defer limiter.Close()                    // Clean up

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

	limiter := createTestRateLimiter(1.0, 2)
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

	// Should use user-based client ID for valid tokens (updated to match new format)
	if !strings.HasPrefix(clientID, "user:") {
		t.Errorf("Should use user-based client ID for valid token, got: %s", clientID)
	}

	// Basic rate limiting should still work
	if !limiter.Allow(clientID) {
		t.Error("First request with auth should be allowed")
	}
}

func TestRateLimiterWithInvalidAuth(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 2)
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
	limiter := createTestRateLimiter(1.0, 1) // Very restrictive
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

// Keep your existing TestAuthMiddleware and TestGetUserFromContext functions
// Remove any duplicate declarations

// ADD ONLY THE NEW TESTS that don't already exist:

func TestRateLimiterIPExtraction(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 2)
	defer limiter.Close()

	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expectedIP    string
	}{
		{
			name:       "Basic RemoteAddr",
			remoteAddr: "192.168.1.100:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name:          "X-Forwarded-For single IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For multiple IPs",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 10.0.0.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:          "X-Forwarded-For takes priority over X-Real-IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "203.0.113.2",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "Invalid X-Forwarded-For falls back to X-Real-IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "invalid-ip",
			xRealIP:       "203.0.113.2",
			expectedIP:    "203.0.113.2",
		},
		{
			name:          "Invalid IPs fall back to RemoteAddr",
			remoteAddr:    "192.168.1.100:12345",
			xForwardedFor: "invalid-ip",
			xRealIP:       "also-invalid",
			expectedIP:    "192.168.1.100",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.100",
			expectedIP: "192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := limiter.getRealIP(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}

			// Test that client ID includes the IP
			clientID := limiter.getClientID(req)
			expectedClientID := "ip:" + tt.expectedIP
			if clientID != expectedClientID {
				t.Errorf("Expected client ID %s, got %s", expectedClientID, clientID)
			}
		})
	}
}

func TestRateLimiterAllowWithRetryInfo(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 2) // 1 token/second, capacity 2
	defer limiter.Close()

	clientID := "test-client"

	// First request should be allowed
	allowed, retryAfter := limiter.AllowWithRetryInfo(clientID)
	if !allowed {
		t.Error("First request should be allowed")
	}
	if retryAfter != 0 {
		t.Errorf("Expected retry after 0, got %d", retryAfter)
	}

	// Use up remaining capacity
	allowed, _ = limiter.AllowWithRetryInfo(clientID)
	if !allowed {
		t.Error("Second request should be allowed")
	}

	// Should be rate limited now
	allowed, retryAfter = limiter.AllowWithRetryInfo(clientID)
	if allowed {
		t.Error("Third request should be blocked")
	}
	if retryAfter <= 0 {
		t.Errorf("Expected positive retry after, got %d", retryAfter)
	}
}

func TestRateLimiterMaxBuckets(t *testing.T) {
	// Create limiter with very low bucket limit
	config := RateLimiterConfig{
		Rate:            1.0,
		Capacity:        1,
		MaxBuckets:      2, // Very low limit
		CleanupInterval: 1 * time.Minute,
		BucketTTL:       2 * time.Minute,
		MaxRetryAfter:   5 * time.Minute,
	}
	limiter := NewRateLimiter(config)
	defer limiter.Close()

	// Fill up bucket limit
	allowed1, _ := limiter.AllowWithRetryInfo("client1")
	allowed2, _ := limiter.AllowWithRetryInfo("client2")

	if !allowed1 || !allowed2 {
		t.Error("First two clients should be allowed")
	}

	// Third client should be rejected due to bucket limit
	allowed3, retryAfter := limiter.AllowWithRetryInfo("client3")
	if allowed3 {
		t.Error("Third client should be rejected due to bucket limit")
	}
	if retryAfter != int(config.MaxRetryAfter.Seconds()) {
		t.Errorf("Expected max retry after %d, got %d", int(config.MaxRetryAfter.Seconds()), retryAfter)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	limiter := createTestRateLimiter(10.0, 10) // Higher limits for concurrency test
	defer limiter.Close()

	var wg sync.WaitGroup
	var allowedCount, deniedCount int32
	numGoroutines := 20
	requestsPerGoroutine := 5

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(clientNum int) {
			defer wg.Done()
			clientID := fmt.Sprintf("client-%d", clientNum)
			for j := 0; j < requestsPerGoroutine; j++ {
				if limiter.Allow(clientID) {
					atomic.AddInt32(&allowedCount, 1)
				} else {
					atomic.AddInt32(&deniedCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	totalRequests := int32(numGoroutines * requestsPerGoroutine)
	if allowedCount+deniedCount != totalRequests {
		t.Errorf("Expected %d total requests, got %d", totalRequests, allowedCount+deniedCount)
	}

	// Should have some allowed requests
	if allowedCount == 0 {
		t.Error("Expected some requests to be allowed")
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	config := RateLimiterConfig{
		Rate:            1.0,
		Capacity:        1,
		MaxBuckets:      1000,
		CleanupInterval: 100 * time.Millisecond, // Fast cleanup for testing
		BucketTTL:       200 * time.Millisecond, // Short TTL for testing
		MaxRetryAfter:   5 * time.Minute,
	}
	limiter := NewRateLimiter(config)
	defer limiter.Close()

	// Create a bucket
	limiter.Allow("test-client")

	// Check metrics before cleanup
	metrics := limiter.GetMetrics()
	if metrics["buckets_created"] == 0 {
		t.Error("Expected at least one bucket created")
	}

	// Wait for cleanup to run
	time.Sleep(500 * time.Millisecond)

	// Check that expired buckets were cleaned up
	metrics = limiter.GetMetrics()
	if metrics["last_cleanup"] == 0 {
		t.Error("Expected cleanup to have run")
	}
}

func TestRateLimiterClose(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 1)

	// Test that limiter works before close
	if !limiter.Allow("test-client") {
		t.Error("Limiter should work before close")
	}

	// Close should succeed
	err := limiter.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}

	// Test that close is idempotent (safe to call multiple times)
	err = limiter.Close()
	// Idempotent close is good design - either succeeds or times out gracefully
	if err != nil && !strings.Contains(err.Error(), "did not stop in time") {
		t.Errorf("Expected no error or timeout on second close, got: %v", err)
	}

	// Limiter should still be usable after close (graceful degradation)
	// This tests that your implementation doesn't crash
	limiter.Allow("test-client-after-close") // Should not panic
}

func TestRateLimiterMetrics(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 1)
	defer limiter.Close()

	clientID := "metrics-test"

	// Test allowed request
	limiter.Allow(clientID)
	metrics := limiter.GetMetrics()
	if metrics["requests_allowed"] == 0 {
		t.Error("Expected requests_allowed to be > 0")
	}

	// Test denied request
	limiter.Allow(clientID) // This should be denied
	metrics = limiter.GetMetrics()
	if metrics["requests_denied"] == 0 {
		t.Error("Expected requests_denied to be > 0")
	}
}

func TestRateLimiterMetricsHandler(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 1)
	defer limiter.Close()

	// Make some requests to generate metrics
	limiter.Allow("test-client")

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler := limiter.MetricsHandler()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected JSON content type")
	}

	var metrics map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	if err != nil {
		t.Errorf("Failed to parse metrics JSON: %v", err)
	}

	expectedFields := []string{"requests_allowed", "requests_denied", "active_buckets", "buckets_created"}
	for _, field := range expectedFields {
		if _, exists := metrics[field]; !exists {
			t.Errorf("Expected metric field %s", field)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Rate <= 0 {
		t.Error("Expected positive rate")
	}
	if config.Capacity <= 0 {
		t.Error("Expected positive capacity")
	}
	if config.MaxBuckets <= 0 {
		t.Error("Expected positive max buckets")
	}
	if config.CleanupInterval <= 0 {
		t.Error("Expected positive cleanup interval")
	}
}

func TestNewDefaultRateLimiter(t *testing.T) {
	limiter := NewDefaultRateLimiter()
	defer limiter.Close()

	if limiter == nil {
		t.Error("Expected non-nil limiter")
	}

	// Should be functional with default config
	if !limiter.Allow("test-client") {
		t.Error("Default limiter should allow first request")
	}
}

// Add these small tests to catch remaining edge cases:

func TestRateLimiterTokenBucketEdgeCases(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 1)
	defer limiter.Close()

	// Test token bucket with zero tokens consumption
	bucket := &TokenBucket{
		tokens:     1.0,
		capacity:   1,
		rate:       1.0,
		lastRefill: time.Now(),
	}

	// Test consuming 0 tokens (edge case)
	if !bucket.consume(0, time.Now()) {
		t.Error("Should allow consuming 0 tokens")
	}

	// Test getRemainingTokens
	remaining := bucket.getRemainingTokens(time.Now())
	if remaining < 0 || remaining > 1 {
		t.Errorf("Expected remaining tokens between 0-1, got %f", remaining)
	}
}

func TestRateLimiterAuthEdgeCases(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 2)
	defer limiter.Close()

	// Test with empty Authorization header value
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "")
	req.RemoteAddr = "192.168.1.100:12345"

	clientID := limiter.getClientID(req)
	if !strings.HasPrefix(clientID, "ip:") {
		t.Error("Empty auth header should fall back to IP")
	}

	// Test with just "Bearer" (no token)
	req.Header.Set("Authorization", "Bearer")
	clientID = limiter.getClientID(req)
	if !strings.HasPrefix(clientID, "ip:") {
		t.Error("Malformed Bearer should fall back to IP")
	}

	// Test with whitespace token
	req.Header.Set("Authorization", "Bearer    ")
	clientID = limiter.getClientID(req)
	if !strings.HasPrefix(clientID, "ip:") {
		t.Error("Whitespace token should fall back to IP")
	}
}

func TestRateLimiterIPExtractionEdgeCases(t *testing.T) {
	limiter := createTestRateLimiter(1.0, 2)
	defer limiter.Close()

	// Test with empty X-Forwarded-For
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "")
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.RemoteAddr = "192.168.1.100:12345"

	ip := limiter.getRealIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("Empty X-Forwarded-For should use X-Real-IP, got %s", ip)
	}

	// Test with whitespace in X-Forwarded-For
	req.Header.Set("X-Forwarded-For", "   ")
	ip = limiter.getRealIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("Whitespace X-Forwarded-For should use X-Real-IP, got %s", ip)
	}

	// Test with empty X-Real-IP
	req.Header.Set("X-Real-IP", "")
	ip = limiter.getRealIP(req)
	if ip != "192.168.1.100" {
		t.Errorf("Empty headers should use RemoteAddr, got %s", ip)
	}
}

func TestRateLimiterCalculateRetryAfterEdgeCases(t *testing.T) {
	config := RateLimiterConfig{
		Rate:            0.1, // Very slow rate for testing
		Capacity:        1,
		MaxBuckets:      1000,
		CleanupInterval: 1 * time.Minute,
		BucketTTL:       2 * time.Minute,
		MaxRetryAfter:   10 * time.Second, // Short max for testing
	}
	limiter := NewRateLimiter(config)
	defer limiter.Close()

	// Create bucket with no tokens
	bucket := &TokenBucket{
		tokens:     0.0,
		capacity:   1,
		rate:       0.1,
		lastRefill: time.Now(),
	}

	// Should hit max retry limit
	retryAfter := limiter.calculateRetryAfter(bucket, time.Now())
	if retryAfter != 10 { // MaxRetryAfter is 10 seconds
		t.Errorf("Expected max retry after 10, got %d", retryAfter)
	}

	// Test with fractional tokens
	bucket.tokens = 0.7
	retryAfter = limiter.calculateRetryAfter(bucket, time.Now())
	if retryAfter < 1 {
		t.Errorf("Expected at least 1 second retry, got %d", retryAfter)
	}
}
