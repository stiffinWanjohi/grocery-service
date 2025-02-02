package postgres

import "errors"

// Repository errors
var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrCustomerNotFound  = errors.New("customer not found")
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderItemNotFound = errors.New("order item not found")
)