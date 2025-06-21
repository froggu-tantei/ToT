package middleware

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/froggu-tantei/ToT/auth"
	"github.com/froggu-tantei/ToT/models"
)

// TokenBucket represents a thread-safe token bucket for rate limiting
type TokenBucket struct {
	mu         sync.Mutex // Protects all fields below
	tokens     float64    // Current number of tokens available
	capacity   int        // Maximum number of tokens the bucket can hold
	rate       float64    // Rate in tokens per second
	lastRefill time.Time  // Last time tokens were refilled
}

// bucketInfo holds bucket and metadata
type bucketInfo struct {
	bucket   *TokenBucket // The actual token bucket
	lastSeen time.Time    // Last time this bucket was accessed
}

// RateLimiter implements a secure token bucket rate limiter
type RateLimiter struct {
	buckets     map[string]*bucketInfo // Map of client IDs to bucket info
	mu          sync.RWMutex           // Protects the buckets map
	rate        float64                // Rate in tokens per second
	capacity    int                    // Capacity of each bucket
	maxBuckets  int                    // Maximum number of buckets allowed
	cleanupDone chan struct{}          // Channel to signal cleanup goroutine to stop
}

// NewRateLimiter creates a new rate limiter with cleanup
func NewRateLimiter(rate float64, capacity int) *RateLimiter {
	return NewRateLimiterWithLimit(rate, capacity, 10000) // Default limit
}

// NewRateLimiterWithLimit creates a rate limiter with custom bucket limit
func NewRateLimiterWithLimit(rate float64, capacity int, maxBuckets int) *RateLimiter {
	rl := &RateLimiter{
		buckets:     make(map[string]*bucketInfo),
		rate:        rate,
		capacity:    capacity,
		maxBuckets:  maxBuckets,
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// Close stops the cleanup goroutine
func (rl *RateLimiter) Close() {
	close(rl.cleanupDone)
}

// cleanup removes expired buckets
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupExpiredBuckets()
		case <-rl.cleanupDone:
			return
		}
	}
}

// cleanupExpiredBuckets removes buckets unused for 10 minutes
func (rl *RateLimiter) cleanupExpiredBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for clientID, info := range rl.buckets {
		if info.lastSeen.Before(cutoff) {
			delete(rl.buckets, clientID)
		}
	}
}

// consume attempts to consume tokens from the bucket - NOW THREAD-SAFE! ✅
func (tb *TokenBucket) consume(tokens int, now time.Time) bool {
	tb.mu.Lock()         // LOCK: Protect all operations
	defer tb.mu.Unlock() // UNLOCK: Always release

	// Calculate token refill since last check
	timePassed := now.Sub(tb.lastRefill).Seconds()
	refill := timePassed * tb.rate

	// Add tokens to bucket, up to capacity
	tb.tokens += refill
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}

	// Update last refill time
	tb.lastRefill = now

	// Check if request can be allowed
	if tb.tokens < float64(tokens) {
		return false
	}

	// Consume tokens
	tb.tokens -= float64(tokens)
	return true
}

// getClientID generates a secure client identifier
func (rl *RateLimiter) getClientID(r *http.Request) string {
	// Try to extract user ID from validated token first
	if userID := rl.extractUserID(r); userID != "" {
		// Hash the user ID/token for privacy - FULL HASH for security
		hash := sha256.Sum256([]byte(userID))
		return fmt.Sprintf("token:%x", hash) // Full 256-bit hash
	}

	// Fallback to IP with proper forwarded header handling
	ip := rl.getRealIP(r)
	return fmt.Sprintf("ip:%s", ip)
}

// extractUserID extracts user ID from Authorization header
func (rl *RateLimiter) extractUserID(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate JWT token using your existing auth package
	claims, err := auth.ValidateToken(token)
	if err != nil {
		// Invalid token - fall back to IP-based rate limiting
		return ""
	}

	// Return the actual user ID from the validated token
	return claims.UserID.String()
}

// getRealIP extracts the real client IP considering proxies
func (rl *RateLimiter) getRealIP(r *http.Request) string {
	// Check X-Forwarded-For first (common behind proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr, strip port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Allow checks if a request is allowed under rate limiting
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.buckets[clientID] // Check if bucket exists for this client

	// If bucket exists, check if it's still valid
	if len(rl.buckets) >= rl.maxBuckets && !exists {
		return false // Reject new clients if too many buckets
	}

	if !exists {
		bucket := &TokenBucket{
			tokens:     float64(rl.capacity - 1), // Start with capacity-1 (first request consumes 1)
			capacity:   rl.capacity,
			rate:       rl.rate,
			lastRefill: now,
		}
		info = &bucketInfo{
			bucket:   bucket,
			lastSeen: now,
		}
		rl.buckets[clientID] = info
		return true
	}

	info.lastSeen = now
	return info.bucket.consume(1, now) // Consume 1 token for this request
}

// RateLimitMiddleware creates middleware that applies rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get secure client ID
			clientID := limiter.getClientID(r)

			// Check rate limit
			if !limiter.Allow(clientID) {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				resp := models.NewErrorResponse("Rate limit exceeded. Please try again later.")
				data, _ := json.Marshal(resp) // Marshal the error response
				w.Write(data)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
