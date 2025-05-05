package routes

import (
	"net/http"

	"github.com/XEDJK/ToT/handlers" // Import handlers to access APIConfig and handler methods
	"github.com/XEDJK/ToT/middleware"
)

// RegisterRoutes sets up the application's routes.
// It takes the ServeMux and the APIConfig as dependencies.
func RegisterRoutes(mux *http.ServeMux, apiCfg *handlers.APIConfig) {

	// Root endpoint
	mux.HandleFunc("GET /", apiCfg.RootHandler)

	// Readiness endpoint
	mux.HandleFunc("GET /v1/readiness", apiCfg.ReadinessHandler)

	// Health check endpoint
	mux.HandleFunc("GET /v1/healthz", apiCfg.HealthzHandler)

	// Error endpoint
	mux.HandleFunc("GET /v1/err", apiCfg.ErrorHandler)

	// User routes
	mux.HandleFunc("POST /v1/users", apiCfg.SignupHandler)
	mux.HandleFunc("POST /v1/login", apiCfg.LoginHandler)

	// User protected routes
	mux.Handle("GET /v1/me", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.GetMeHandler)))
	mux.Handle("GET /v1/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.GetUserByIDHandler)))
	mux.Handle("GET /v1/users/username/{username}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.GetUserByUsernameHandler)))
	mux.Handle("PUT /v1/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.UpdateUserHandler)))
	mux.Handle("DELETE /v1/users/{id}", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.DeleteUserHandler)))
	mux.Handle("GET /v1/users", middleware.AuthMiddleware(http.HandlerFunc(apiCfg.ListUsersHandler)))

}
