package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/froggu-tantei/ToT/auth"
	"github.com/froggu-tantei/ToT/models"
)

// RateLimiterConfig holds all configuration for the rate limiter
type RateLimiterConfig struct {
	Rate            float64       // Tokens per second
	Capacity        int           // Bucket capacity
	MaxBuckets      int           // Maximum concurrent buckets
	CleanupInterval time.Duration // How often to cleanup old buckets
	BucketTTL       time.Duration // How long before a bucket expires
	MaxRetryAfter   time.Duration // Maximum retry-after time
}

// DefaultConfig returns sensible defaults
func DefaultConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:            10.0,             // 10 requests per second
		Capacity:        20,               // Burst of 20
		MaxBuckets:      10000,            // 10k concurrent clients
		CleanupInterval: 5 * time.Minute,  // Cleanup every 5 minutes
		BucketTTL:       10 * time.Minute, // Expire buckets after 10 minutes
		MaxRetryAfter:   5 * time.Minute,  // Max 5 minute retry
	}
}

// Metrics holds rate limiter statistics
type Metrics struct {
	RequestsAllowed int64 // Total requests allowed
	RequestsDenied  int64 // Total requests denied
	ActiveBuckets   int64 // Current active buckets
	BucketsCreated  int64 // Total buckets created
	BucketsExpired  int64 // Total buckets expired
	LastCleanup     int64 // Unix timestamp of last cleanup
}

// GetMetrics returns current metrics (thread-safe)
func (m *Metrics) GetMetrics() map[string]int64 {
	return map[string]int64{
		"requests_allowed": atomic.LoadInt64(&m.RequestsAllowed),
		"requests_denied":  atomic.LoadInt64(&m.RequestsDenied),
		"active_buckets":   atomic.LoadInt64(&m.ActiveBuckets),
		"buckets_created":  atomic.LoadInt64(&m.BucketsCreated),
		"buckets_expired":  atomic.LoadInt64(&m.BucketsExpired),
		"last_cleanup":     atomic.LoadInt64(&m.LastCleanup),
	}
}

// TokenBucket represents a thread-safe token bucket for rate limiting
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   int
	rate       float64
	lastRefill time.Time
}

// bucketInfo holds bucket and metadata
type bucketInfo struct {
	bucket   *TokenBucket
	lastSeen int64 // atomic access
}

// RateLimiter implements a production-ready token bucket rate limiter
type RateLimiter struct {
	config  RateLimiterConfig
	buckets sync.Map // Use sync.Map for better concurrent access
	metrics Metrics
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewRateLimiter creates a new rate limiter with custom config
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	// Start background cleanup
	go rl.cleanup()
	return rl
}

// NewDefaultRateLimiter creates a rate limiter with default settings
func NewDefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(DefaultConfig())
}

// Close gracefully shuts down the rate limiter
func (rl *RateLimiter) Close() error {
	rl.cancel()
	select {
	case <-rl.done:
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("cleanup goroutine did not stop in time")
	}
}

// cleanup runs the background cleanup process
func (rl *RateLimiter) cleanup() {
	defer close(rl.done)
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupExpiredBuckets()
		case <-rl.ctx.Done():
			return
		}
	}
}

func (rl *RateLimiter) cleanupExpiredBuckets() {
	cutoff := time.Now().Add(-rl.config.BucketTTL).Unix()
	var expired int64
	var remaining int64

	rl.buckets.Range(func(key, value any) bool {
		info := value.(*bucketInfo)
		if atomic.LoadInt64(&info.lastSeen) < cutoff {
			rl.buckets.Delete(key)
			expired++
		} else {
			remaining++
		}
		return true
	})

	// Update metrics atomically
	atomic.AddInt64(&rl.metrics.BucketsExpired, expired)
	atomic.StoreInt64(&rl.metrics.ActiveBuckets, remaining)
	atomic.StoreInt64(&rl.metrics.LastCleanup, time.Now().Unix())
}

// consume attempts to consume tokens from the bucket
func (tb *TokenBucket) consume(tokens int, now time.Time) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	elapsed := now.Sub(tb.lastRefill).Seconds()
	refill := elapsed * tb.rate
	tb.tokens = min(tb.tokens+refill, float64(tb.capacity))
	tb.lastRefill = now

	// Check if we have enough tokens
	if tb.tokens < float64(tokens) {
		return false
	}

	tb.tokens -= float64(tokens)
	return true
}

// getRemainingTokens returns current token count without consuming
func (tb *TokenBucket) getRemainingTokens(now time.Time) float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	elapsed := now.Sub(tb.lastRefill).Seconds()
	refill := elapsed * tb.rate
	return min(tb.tokens+refill, float64(tb.capacity))
}

// getClientID generates a client identifier with configurable privacy
func (rl *RateLimiter) getClientID(r *http.Request) string {
	// Try JWT-based identification first
	if userID := rl.extractUserID(r); userID != "" {
		// Use first 16 bytes of hash for memory efficiency while maintaining security
		hash := sha256.Sum256([]byte(userID))
		return fmt.Sprintf("user:%x", hash[:16]) // 128-bit hash is plenty
	}

	// Fallback to IP-based identification
	ip := rl.getRealIP(r)
	return fmt.Sprintf("ip:%s", ip)
}

// extractUserID extracts user ID from JWT token
func (rl *RateLimiter) extractUserID(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateToken(token)
	if err != nil {
		return ""
	}

	return claims.UserID.String()
}

// getRealIP extracts the real client IP with validation
func (rl *RateLimiter) getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (rl *RateLimiter) AllowWithRetryInfo(clientID string) (allowed bool, retryAfterSeconds int) {
	now := time.Now() // Single source of truth for this request

	if value, exists := rl.buckets.Load(clientID); exists {
		info := value.(*bucketInfo)
		atomic.StoreInt64(&info.lastSeen, now.Unix())

		if info.bucket.consume(1, now) {
			atomic.AddInt64(&rl.metrics.RequestsAllowed, 1)
			return true, 0
		}

		// Pass the SAME timestamp to ensure consistency
		retryAfter := rl.calculateRetryAfter(info.bucket, now)
		atomic.AddInt64(&rl.metrics.RequestsDenied, 1)
		return false, retryAfter
	}

	// Check if we're at capacity (simple protection)
	activeCount := atomic.LoadInt64(&rl.metrics.ActiveBuckets)
	if activeCount >= int64(rl.config.MaxBuckets) {
		atomic.AddInt64(&rl.metrics.RequestsDenied, 1)
		return false, int(rl.config.MaxRetryAfter.Seconds())
	}

	// Create new bucket
	bucket := &TokenBucket{
		tokens:     float64(rl.config.Capacity),
		capacity:   rl.config.Capacity,
		rate:       rl.config.Rate,
		lastRefill: now,
	}

	info := &bucketInfo{
		bucket:   bucket,
		lastSeen: now.Unix(),
	}

	// Store the bucket
	rl.buckets.Store(clientID, info)

	// Update metrics
	atomic.AddInt64(&rl.metrics.BucketsCreated, 1)
	atomic.AddInt64(&rl.metrics.ActiveBuckets, 1)

	// Allow the first request
	bucket.consume(1, now)
	atomic.AddInt64(&rl.metrics.RequestsAllowed, 1)
	return true, 0
}

// Allow is a simple wrapper for backward compatibility
func (rl *RateLimiter) Allow(clientID string) bool {
	allowed, _ := rl.AllowWithRetryInfo(clientID)
	return allowed
}

func (rl *RateLimiter) calculateRetryAfter(bucket *TokenBucket, now time.Time) int {
	// Use the passed timestamp, don't call getRemainingTokens with a new time
	currentTokens := bucket.getRemainingTokensAtTime(now)

	// This check can now be removed or turned into a defensive assertion
	if currentTokens >= 1.0 {
		// This really shouldn't happen now, but if it does, something's wrong
		return 1 // Or log an error
	}

	tokensNeeded := 1.0 - currentTokens
	secondsNeeded := tokensNeeded / rl.config.Rate
	retryAfter := int(math.Ceil(secondsNeeded))

	return max(1, min(retryAfter, int(rl.config.MaxRetryAfter.Seconds())))
}

func (tb *TokenBucket) getRemainingTokensAtTime(now time.Time) float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	elapsed := now.Sub(tb.lastRefill).Seconds()
	refill := elapsed * tb.rate
	return min(tb.tokens+refill, float64(tb.capacity))
}

// GetMetrics returns current rate limiter metrics
func (rl *RateLimiter) GetMetrics() map[string]int64 {
	return rl.metrics.GetMetrics()
}

// RateLimitMiddleware creates HTTP middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := limiter.getClientID(r)
			allowed, retryAfter := limiter.AllowWithRetryInfo(clientID)

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", limiter.config.Rate))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusTooManyRequests)

				resp := models.NewErrorResponse("Rate limit exceeded. Please try again later.")
				data, _ := json.Marshal(resp)
				w.Write(data)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsHandler provides an HTTP endpoint for metrics
func (rl *RateLimiter) MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metrics := rl.GetMetrics()
		json.NewEncoder(w).Encode(metrics)
	}
}
