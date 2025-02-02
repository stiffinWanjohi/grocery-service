package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *db.PostgresDB
}

func NewOrderRepository(db *db.PostgresDB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	if err := r.db.DB.WithContext(ctx).Create(order).Error; err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.DB.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		First(&order, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.DB.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Find(&orders).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) ListByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.DB.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Where("customer_id = ?", customerID).
		Find(&orders).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list orders by customer ID: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	result := r.db.DB.WithContext(ctx).Save(order)
	if result.Error != nil {
		return fmt.Errorf("failed to update order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	result := r.db.DB.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update order status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) AddOrderItem(ctx context.Context, orderItem *domain.OrderItem) error {
	if err := r.db.DB.WithContext(ctx).Create(orderItem).Error; err != nil {
		return fmt.Errorf("failed to add order item: %w", err)
	}
	return nil
}

func (r *OrderRepository) RemoveOrderItem(ctx context.Context, orderID, orderItemID string) error {
	result := r.db.DB.WithContext(ctx).
		Where("order_id = ? AND id = ?", orderID, orderItemID).
		Delete(&domain.OrderItem{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove order item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrOrderItemNotFound
	}
	return nil
}
