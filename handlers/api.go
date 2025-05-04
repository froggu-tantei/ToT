// filepath: /home/wst/Documents/Code/ToT/handlers/api.go
package handlers

import (
	"net/http"

	"github.com/XEDJK/ToT/db/database" // Import database package
)

// APIConfig holds the dependencies for the API handlers.
type APIConfig struct {
	DB *database.Queries
}

// NewAPIConfig creates a new APIConfig.
func NewAPIConfig(db *database.Queries) *APIConfig {
	return &APIConfig{
		DB: db,
	}
}

// RootHandler handles requests to the root path.
func (cfg *APIConfig) RootHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]string{
		"name":    "Throne of Thorns API",
		"version": "1.0.0",
		"status":  "running",
		"author":  "XEDJK",
	})
}

// ReadinessHandler handles the readiness check endpoint.
func (cfg *APIConfig) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "ok"})
}

// HealthzHandler handles the health check endpoint.
func (cfg *APIConfig) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "ok"}) // Simple health check
}

// ErrorHandler is a simple handler that always returns an error.
func (cfg *APIConfig) ErrorHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
}
