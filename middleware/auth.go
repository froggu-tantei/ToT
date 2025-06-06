package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/froggu-tantei/ToT/auth"
	"github.com/froggu-tantei/ToT/models"
)

// Key for storing user claims in request context
type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware authenticates requests using JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No Authorization header
			respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper function to get user claims from context
func GetUserFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	return claims, ok
}

// Helper function to respond with error
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Use the models.ErrorResponse for consistent error formatting
	resp := models.NewErrorResponse(message)
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling error response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
