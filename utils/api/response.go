package api

import (
	"encoding/json"
	"net/http"
)

// Response represents the standard API response format
// @Description Standard API response structure
type Response struct {
	// Indicates if the request was successful
	Success bool `json:"success" example:"true"`
	// Contains the response data (if any)
	Data interface{} `json:"data,omitempty"`
	// Contains error message (if any)
	Error string `json:"error,omitempty" example:"Invalid request parameters"`
}

// SuccessResponse writes a success response to the http.ResponseWriter
func SuccessResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse writes an error response to the http.ResponseWriter
func ErrorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}
