package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"

	"github.com/froggu-tantei/ToT/models"
)

// isValidEmail validates email format using Go's standard library
func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	// Ensure it's just an email address, not "Name <email@domain.com>" format
	return addr.Address == email
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		// Return JSON error even in error cases for consistency
		w.Write([]byte(`{"error":"Internal Server Error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

// RespondWithError sends a JSON error response using models.ErrorResponse
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	// Check for common client errors and adjust message if needed
	if code > 399 && code < 500 {
		log.Printf("Client error %d: %s", code, msg)
	}
	// Check for server errors and log potentially more details
	if code > 499 {
		log.Printf("Server error %d: %s", code, msg)
	}

	// Use the models.ErrorResponse for consistent error formatting
	resp := models.NewErrorResponse(msg)
	data, err := json.Marshal(resp)
	if err != nil {
		// Log the marshalling error and send a generic server error
		log.Printf("Error marshalling error response: %v", err)
		w.Header().Set("Content-Type", "application/json") // Still try to set content type
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
