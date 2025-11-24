package httpresp

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// Error writes an error response payload.
func Error(w http.ResponseWriter, status int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	JSON(w, status, errorResponse{Error: message})
}
