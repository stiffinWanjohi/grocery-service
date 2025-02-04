package utils

import (
	"errors"
	"fmt"
	"log"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   error
	Message string
	Code    string
}

// Application Error Codes
const (
	// Authentication Errors
	ErrCodeInvalidCredentials = "AUTH001"
	ErrCodeTokenExpired       = "AUTH002"
	ErrCodeInvalidToken       = "AUTH003"
	ErrCodeUnauthorized       = "AUTH004"

	// Customer Errors
	ErrCodeCustomerNotFound    = "CUST001"
	ErrCodeInvalidCustomerData = "CUST002"
	ErrCodeDuplicateEmail      = "CUST003"

	// Product Errors
	ErrCodeProductNotFound    = "PROD001"
	ErrCodeInvalidProductData = "PROD002"
	ErrCodeInsufficientStock  = "PROD003"
	ErrCodeProductUnavailable = "PROD004"

	// Order Errors
	ErrCodeOrderNotFound      = "ORD001"
	ErrCodeInvalidOrderData   = "ORD002"
	ErrCodeOrderStatusInvalid = "ORD003"
	ErrCodeEmptyOrder         = "ORD004"

	// Category Errors
	ErrCodeCategoryNotFound    = "CAT001"
	ErrCodeInvalidCategoryData = "CAT002"

	// Database Errors
	ErrCodeDBConnection = "DB001"
	ErrCodeDBQuery      = "DB002"
	ErrCodeDBDuplicate  = "DB003"

	// Notification Errors
	ErrCodeEmailSendFailed = "NOTIF001"
	ErrCodeSMSSendFailed   = "NOTIF002"

	// Validation Errors
	ErrCodeInvalidInput  = "VAL001"
	ErrCodeRequired      = "VAL002"
	ErrCodeInvalidFormat = "VAL003"
)

// Pre-defined application errors
var (
	// Authentication Errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUnauthorized       = errors.New("unauthorized access")

	// Customer Errors
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrInvalidCustomerData = errors.New("invalid customer data")
	ErrDuplicateEmail      = errors.New("email already exists")

	// Product Errors
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductData = errors.New("invalid product data")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrProductUnavailable = errors.New("product is unavailable")

	// Order Errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderData   = errors.New("invalid order data")
	ErrOrderStatusInvalid = errors.New("invalid order status")
	ErrEmptyOrder         = errors.New("order is empty")

	// Category Errors
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidCategoryData = errors.New("invalid category data")

	// Database Errors
	ErrDBConnection = errors.New("database connection error")
	ErrDBQuery      = errors.New("database query error")
	ErrDBDuplicate  = errors.New("duplicate entry in database")
)

// LogError logs an error and returns an ErrorResponse
func LogError(err error, message string, code string) ErrorResponse {
	log.Printf("Error [%s]: %v - %s", code, err, message)
	return ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrCustomerNotFound) ||
		errors.Is(err, ErrProductNotFound) ||
		errors.Is(err, ErrOrderNotFound) ||
		errors.Is(err, ErrCategoryNotFound)
}

// IsDuplicate checks if the error is a duplicate error
func IsDuplicate(err error) bool {
	return errors.Is(err, ErrDuplicateEmail) ||
		errors.Is(err, ErrDBDuplicate)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidCustomerData) ||
		errors.Is(err, ErrInvalidProductData) ||
		errors.Is(err, ErrInvalidOrderData) ||
		errors.Is(err, ErrInvalidCategoryData)
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrUnauthorized)
}
