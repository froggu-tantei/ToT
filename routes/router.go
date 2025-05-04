// filepath: /home/wst/Documents/Code/ToT/routes/router.go
package routes

import (
	"net/http"

	"github.com/XEDJK/ToT/handlers" // Import handlers to access APIConfig and handler methods
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

}
