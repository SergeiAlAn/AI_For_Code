// Package respond contains helpers for writing JSON HTTP responses.
package respond

import (
	"encoding/json"
	"log"
	"net/http"

	"ai-for-code/go-server/internal/models"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, models.ErrorResponse{Error: message})
}

// InternalError logs err and hides its details from the client.
func InternalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	Error(w, http.StatusInternalServerError, "internal server error")
}
