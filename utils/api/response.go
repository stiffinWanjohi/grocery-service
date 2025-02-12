package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Response represents the standard API response format
// @Description Standard API response structure
type Response struct {
	Success bool        `json:"success"         example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:"Invalid request parameters"`
}

var ErrResponseEncoding = errors.New("failed to encode response")

func SuccessResponse(
	w http.ResponseWriter,
	data interface{},
	status int,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	}); err != nil {
		return fmt.Errorf(
			"%w: %v",
			ErrResponseEncoding,
			err,
		)
	}
	return nil
}

func ErrorResponse(
	w http.ResponseWriter,
	message string,
	status int,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	}); err != nil {
		return fmt.Errorf(
			"%w: %v",
			ErrResponseEncoding,
			err,
		)
	}
	return nil
}
