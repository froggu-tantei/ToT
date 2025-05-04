package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// RespondWithJSON is a helper to send a JSON response.
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// RespondWithError is a helper to send an error response with a JSON message.
func RespondWithError(w http.ResponseWriter, statusCode int, msg string) {
	RespondWithJSON(w, statusCode, struct {
		Error string `json:"error"`
	}{Error: msg})
}
