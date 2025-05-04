package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
)

// LoggingMiddleware logs incoming requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// CorsMiddleware sets up and returns a CORS handler.
func CorsMiddleware(next http.Handler) http.Handler {
	// Configure CORS
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // TODO: Replace * with your frontend domain in production
		// AllowedOrigins: []string{"http://localhost:3000", "https://your-frontend-domain.com"}, // Example
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         300, // Maximum value not ignored by any major browsers
	}).Handler(next) // Wrap the next handler with CORS middleware
}
