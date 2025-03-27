package errors

import (
	"errors"
	"fmt"
	"log"
)

type ErrorResponse struct {
	Error   error
	Message string
	Code    string
}

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
	ErrInvalidCredentials = errors.New(
		"invalid credentials",
	)
	ErrTokenExpired  = errors.New("token has expired")
	ErrInvalidToken  = errors.New("invalid token")
	ErrUnauthorized  = errors.New("unauthorized access")
	ErrForbidden     = errors.New("forbidden access")
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenRevoked  = errors.New("token has been revoked")

	// Customer Errors
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrInvalidCustomerData = errors.New("invalid customer data")
	ErrCustomerExists      = errors.New("customer already exists")

	// Product Errors
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductData = errors.New("invalid product data")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrProductUnavailable = errors.New("product is unavailable")

	// Order Errors
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidOrderData = errors.New("invalid order data")

	ErrOrderStatusInvalid   = errors.New("invalid order status")
	ErrEmptyOrder           = errors.New("order is empty")
	ErrOrderItemNotFound    = errors.New("order item not found")
	ErrInvalidOrderItemData = errors.New("invalid order item data")

	// Category Errors
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidCategoryData = errors.New("invalid category data")

	// User Errors
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateEmail  = errors.New("email already exists")
	ErrInvalidUserData = errors.New("invalid user data")

	// Database Errors
	ErrDBConnection   = errors.New("database connection error")
	ErrDBQuery        = errors.New("database query error")
	ErrDBDuplicate    = errors.New("duplicate entry in database")
	ErrInternalServer = errors.New("internal server error")
)

func LogError(
	err error,
	message string,
	code string,
) ErrorResponse {
	log.Printf("Error [%s]: %v - %s", code, err, message)
	return ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	}
}

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrCustomerNotFound) ||
		errors.Is(err, ErrProductNotFound) ||
		errors.Is(err, ErrOrderNotFound) ||
		errors.Is(err, ErrCategoryNotFound) ||
		errors.Is(err, ErrOrderItemNotFound) ||
		errors.Is(err, ErrUserNotFound)
}

func IsDuplicate(err error) bool {
	return errors.Is(err, ErrDuplicateEmail) ||
		errors.Is(err, ErrDBDuplicate)
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidCustomerData) ||
		errors.Is(err, ErrInvalidProductData) ||
		errors.Is(err, ErrInvalidOrderData) ||
		errors.Is(err, ErrInvalidCategoryData) ||
		errors.Is(err, ErrInvalidUserData)
}

func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrUnauthorized)
}
