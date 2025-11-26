package httpresp

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse represents the standardized payload for successful responses.
type SuccessResponse struct {
	StatusCode int         `json:"status_code"`
	Status     bool        `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// ErrorResponse represents the standardized payload for error responses.
type ErrorResponse struct {
	StatusCode int         `json:"status_code"`
	Status     bool        `json:"status"`
	Message    string      `json:"message"`
	Errors     interface{} `json:"errors"`
}

// ListData wraps list responses along with pagination metadata.
type ListData struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination captures paging metadata for collection responses.
type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}

// JSON writes a standardized success response.
func JSON(w http.ResponseWriter, status int, message string, data interface{}) {
	if message == "" {
		message = defaultMessage(status)
	}
	respond(w, status, SuccessResponse{
		StatusCode: status,
		Status:     isSuccess(status),
		Message:    message,
		Data:       data,
	})
}

// List writes a standardized collection response with pagination metadata.
func List(w http.ResponseWriter, status int, message string, items interface{}, pagination Pagination) {
	if pagination.PerPage == 0 {
		pagination.PerPage = 10
	}
	if message == "" {
		message = defaultMessage(status)
	}
	respond(w, status, SuccessResponse{
		StatusCode: status,
		Status:     isSuccess(status),
		Message:    message,
		Data: ListData{
			Data:       items,
			Pagination: pagination,
		},
	})
}

// Error writes an error response payload.
func Error(w http.ResponseWriter, status int, message string, details interface{}) {
	if message == "" {
		message = defaultMessage(status)
	}
	if details == nil {
		details = message
	}
	respond(w, status, ErrorResponse{
		StatusCode: status,
		Status:     false,
		Message:    message,
		Errors:     details,
	})
}

func respond(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func defaultMessage(status int) string {
	if text := http.StatusText(status); text != "" {
		return text
	}
	return "Success"
}

func isSuccess(status int) bool {
	return status >= http.StatusOK && status < http.StatusBadRequest
}
