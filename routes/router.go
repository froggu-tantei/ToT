package routes

import (
	"github.com/XEDJK/ToT/handlers" // Import handlers to access APIConfig and handler methods
	"github.com/XEDJK/ToT/middleware"
	"github.com/go-chi/chi/v5" // Import chi for routing
)

// RegisterRoutes sets up the application's routes.
func RegisterRoutes(apiCfg *handlers.APIConfig, authLimiter, genericLimiter *middleware.RateLimiter) chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.CorsMiddleware)
	r.Use(middleware.LoggingMiddleware)

	// Root endpoint
	r.With(middleware.RateLimitMiddleware(genericLimiter)).Get("/", apiCfg.RootHandler)

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Health endpoints
		r.With(middleware.RateLimitMiddleware(genericLimiter)).Get("/readiness", apiCfg.ReadinessHandler)
		r.With(middleware.RateLimitMiddleware(genericLimiter)).Get("/healthz", apiCfg.HealthzHandler)
		r.Get("/err", apiCfg.ErrorHandler)

		// User authentication routes
		r.With(middleware.RateLimitMiddleware(authLimiter)).Post("/users", apiCfg.SignupHandler)
		r.With(middleware.RateLimitMiddleware(authLimiter)).Post("/login", apiCfg.LoginHandler)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			r.Get("/me", apiCfg.GetMeHandler)
			r.Get("/users", apiCfg.ListUsersHandler)
			r.Get("/users/{id}", apiCfg.GetUserByIDHandler)
			r.Get("/users/username/{username}", apiCfg.GetUserByUsernameHandler)
			r.Put("/users/{id}", apiCfg.UpdateUserHandler)
			r.Delete("/users/{id}", apiCfg.DeleteUserHandler)
			r.Post("/users/{id}/profile-picture", apiCfg.UploadProfilePictureHandler)
		})

		// Leaderboard
		r.With(middleware.RateLimitMiddleware(genericLimiter)).Get("/leaderboard", apiCfg.GetLeaderboardHandler)
	})

	return r
}
