package routes

import (
	"net/http"

	"github.com/XEDJK/ToT/handlers" // Import handlers to access APIConfig and handler methods
	"github.com/XEDJK/ToT/middleware"
)

// RegisterRoutes sets up the application's routes.
// It takes the ServeMux, the APIConfig, and rate limiters as parameters.
func RegisterRoutes(mux *http.ServeMux, apiCfg *handlers.APIConfig, authLimiter, genericLimiter *middleware.RateLimiter) {

	// Root endpoint
	mux.Handle("GET /", middleware.RateLimitMiddleware(genericLimiter)(http.HandlerFunc(apiCfg.RootHandler)))

	// Readiness endpoint
	mux.Handle("GET /v1/readiness", middleware.RateLimitMiddleware(genericLimiter)(http.HandlerFunc(apiCfg.ReadinessHandler)))

	// Health check endpoint
	mux.Handle("GET /v1/healthz", middleware.RateLimitMiddleware(genericLimiter)(http.HandlerFunc(apiCfg.HealthzHandler)))

	// Error endpoint
	mux.HandleFunc("GET /v1/err", apiCfg.ErrorHandler)

	// User routes
	mux.Handle("POST /v1/users", middleware.RateLimitMiddleware(authLimiter)(http.HandlerFunc(apiCfg.SignupHandler)))
	mux.Handle("POST /v1/login", middleware.RateLimitMiddleware(authLimiter)(http.HandlerFunc(apiCfg.LoginHandler)))

	// User protected routes
	mux.Handle("GET /v1/me", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.GetMeHandler)))
	mux.Handle("GET /v1/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.GetUserByIDHandler)))
	mux.Handle("GET /v1/users/username/{username}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.GetUserByUsernameHandler)))
	mux.Handle("PUT /v1/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.UpdateUserHandler)))
	mux.Handle("DELETE /v1/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.DeleteUserHandler)))
	mux.Handle("GET /v1/users", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.ListUsersHandler)))
	mux.Handle("POST /v1/users/{id}/profile-picture", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.UploadProfilePictureHandler)))
	mux.Handle("GET /v1/leaderboard", middleware.RateLimitMiddleware(genericLimiter)(http.HandlerFunc(apiCfg.GetLeaderboardHandler)))
}
