package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/XEDJK/ToT/models"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	rate       float64              // Tokens per second
	capacity   int                  // Maximum bucket capacity
	tokens     map[string]float64   // Current tokens by client ID
	lastRefill map[string]time.Time // Last time tokens were added
	mu         sync.Mutex           // For thread safety
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     make(map[string]float64),
		lastRefill: make(map[string]time.Time),
	}
}

// Allow checks if a request is allowed under rate limiting
// Allow checks if a request is allowed under rate limiting
func (r *RateLimiter) Allow(clientID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Initialize new client
	if _, exists := r.tokens[clientID]; !exists {
		r.tokens[clientID] = float64(r.capacity) - 1.0 // Consume one token immediately
		r.lastRefill[clientID] = now
		return true
	}

	// Calculate token refill since last check
	timePassed := now.Sub(r.lastRefill[clientID]).Seconds()
	refill := timePassed * r.rate

	// Add tokens to bucket, up to capacity
	currentTokens := r.tokens[clientID] + refill
	if currentTokens > float64(r.capacity) {
		currentTokens = float64(r.capacity)
	}

	// Update last refill time
	r.lastRefill[clientID] = now

	// Check if request can be allowed
	if currentTokens < 1.0 {
		r.tokens[clientID] = currentTokens // Update even if not allowing
		return false
	}

	// Consume token and update state
	r.tokens[clientID] = currentTokens - 1.0
	return true
}

// RateLimitMiddleware creates middleware that applies rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as client ID by default
			clientID := r.RemoteAddr

			// Use Authorization header as client ID if available
			if auth := r.Header.Get("Authorization"); auth != "" {
				clientID = auth
			}

			// Check rate limit
			if !limiter.Allow(clientID) {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
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
