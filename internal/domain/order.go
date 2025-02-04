package domain

import (
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
	ID         uuid.UUID   `json:"id" gorm:"type:uuid;primary_key"`
	CustomerID uuid.UUID   `json:"customer_id" gorm:"type:uuid;not null"`
	Customer   Customer    `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Items      []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	Status     OrderStatus `json:"status" gorm:"not null"`
	TotalPrice float64     `json:"total_price" gorm:"not null"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	Product   Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"`
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

func NewOrderItem(orderID, productID uuid.UUID, quantity int, price float64) *OrderItem {
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
