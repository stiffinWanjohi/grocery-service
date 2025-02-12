package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusPreparing OrderStatus = "PREPARING"
	OrderStatusReady     OrderStatus = "READY"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRefunded  OrderStatus = "REFUNDED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type Order struct {
	ID         uuid.UUID   `json:"id"                 gorm:"type:uuid;primary_key"`
	CustomerID uuid.UUID   `json:"customer_id"        gorm:"type:uuid;not null"`
	Customer   *Customer   `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Items      []OrderItem `json:"items"              gorm:"foreignKey:OrderID"`
	Status     OrderStatus `json:"status"             gorm:"not null"`
	TotalPrice float64     `json:"total_price"        gorm:"not null"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"                gorm:"type:uuid;primary_key"`
	OrderID   uuid.UUID `json:"order_id"          gorm:"type:uuid;not null"`
	ProductID uuid.UUID `json:"product_id"        gorm:"type:uuid;not null"`
	Product   *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Quantity  int       `json:"quantity"          gorm:"not null"`
	Price     float64   `json:"price"             gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewOrder(customerID uuid.UUID) *Order {
	return &Order{
		ID:         uuid.New(),
		CustomerID: customerID,
		Status:     OrderStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func NewOrderItem(
	orderID, productID uuid.UUID,
	quantity int,
	price float64,
) *OrderItem {
	return &OrderItem{
		ID:        uuid.New(),
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (o *Order) Validate() error {
	if o.CustomerID.String() == "" {
		return fmt.Errorf("customer ID is required")
	}
	if len(o.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	for _, item := range o.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (i *OrderItem) Validate() error {
	if i.ProductID.String() == "" {
		return fmt.Errorf("product ID is required")
	}
	if i.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	return nil
}

func ValidateOrderStatus(status OrderStatus) error {
	validStatuses := map[OrderStatus]bool{
		OrderStatusPending:   true,
		OrderStatusConfirmed: true,
		OrderStatusPreparing: true,
		OrderStatusReady:     true,
		OrderStatusShipped:   true,
		OrderStatusDelivered: true,
		OrderStatusCancelled: true,
		OrderStatusRefunded:  true,
		OrderStatusFailed:    true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid order status: %s", status)
	}
	return nil
}

func IsValidStatusTransition(from, to OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending: {
			OrderStatusConfirmed,
			OrderStatusCancelled,
			OrderStatusFailed,
		},
		OrderStatusConfirmed: {
			OrderStatusPreparing,
			OrderStatusCancelled,
			OrderStatusFailed,
		},
		OrderStatusPreparing: {
			OrderStatusReady,
			OrderStatusCancelled,
			OrderStatusFailed,
		},
		OrderStatusReady: {
			OrderStatusShipped,
			OrderStatusCancelled,
			OrderStatusFailed,
		},
		OrderStatusShipped: {
			OrderStatusDelivered,
			OrderStatusFailed,
		},
		OrderStatusDelivered: {
			OrderStatusRefunded,
		},
		OrderStatusCancelled: {
			OrderStatusRefunded,
		},
		OrderStatusRefunded: {}, // No further transitions
		OrderStatusFailed:   {}, // No further transitions
	}

	allowedTransitions, exists := transitions[from]
	if !exists {
		return false
	}

	for _, status := range allowedTransitions {
		if status == to {
			return true
		}
	}
	return false
}
