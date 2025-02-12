package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"gorm.io/gorm"
)

type (
	OrderRepository interface {
		Create(
			ctx context.Context,
			order *domain.Order,
			updateStockFunc func(ctx context.Context, productID string, newStock int) error,
		) error
		GetByID(ctx context.Context, id string) (*domain.Order, error)
		List(ctx context.Context) ([]domain.Order, error)
		ListByCustomerID(
			ctx context.Context,
			customerID string,
		) ([]domain.Order, error)
		Update(ctx context.Context, order *domain.Order) error
		UpdateStatus(
			ctx context.Context,
			id string,
			status domain.OrderStatus,
		) error
		AddOrderItem(
			ctx context.Context,
			orderID string, item *domain.OrderItem,
			updateStockFunc func(ctx context.Context, productID string, newStock int) error,
			updateOrderTotalFunc func(ctx context.Context, order *domain.Order, price float64) error,
		) error
		RemoveOrderItem(
			ctx context.Context,
			orderID, itemID string,
			restoreStockFunc func(ctx context.Context, productID string, quantity int) error,
			updateOrderTotalFunc func(ctx context.Context, order *domain.Order, price float64) error,
		) error
	}

	OrderRepositoryImpl struct {
		*db.BaseRepository[domain.Order]
	}
)

func NewOrderRepository(
	postgres *db.PostgresDB,
) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.Order](
			postgres,
		),
	}
}

func (r *OrderRepositoryImpl) Create(
	ctx context.Context,
	order *domain.Order,
	updateStockFunc func(ctx context.Context, productID string, newStock int) error,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Order]) error {
			for _, item := range order.Items {
				if err := updateStockFunc(ctx, item.ProductID.String(), item.Quantity); err != nil {
					return fmt.Errorf(
						"failed to update product stock: %w",
						err,
					)
				}
			}

			if err := txRepo.GetDB().WithContext(ctx).Create(order).Error; err != nil {
				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrInvalidOrderData,
					err,
				)
			}
			return nil
		},
	)
}

func (r *OrderRepositoryImpl) GetByID(
	ctx context.Context,
	id string,
) (*domain.Order, error) {
	var order domain.Order
	result := r.BaseRepository.GetDB().WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		First(&order, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrOrderNotFound
		}

		return nil, fmt.Errorf(
			"%w: %v",
			customErrors.ErrOrderNotFound,
			result.Error,
		)
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) List(
	ctx context.Context,
) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			customErrors.ErrDBQuery,
			err,
		)
	}
	return orders, nil
}

func (r *OrderRepositoryImpl) ListByCustomerID(
	ctx context.Context,
	customerID string,
) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Where("customer_id = ?", customerID).
		Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			customErrors.ErrDBQuery,
			err,
		)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) Update(
	ctx context.Context,
	order *domain.Order,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Order]) error {
			result := txRepo.GetDB().WithContext(ctx).
				Model(&domain.Order{}).
				Where("id = ?", order.ID).
				Updates(order)
			if result.Error != nil {
				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrInvalidOrderData,
					result.Error,
				)
			}

			if result.RowsAffected == 0 {
				return customErrors.ErrOrderNotFound
			}
			return nil
		},
	)
}

func (r *OrderRepositoryImpl) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.OrderStatus,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Order]) error {
			result := txRepo.GetDB().WithContext(ctx).
				Model(&domain.Order{}).
				Where("id = ?", id).
				Update("status", status)
			if result.Error != nil {
				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrOrderStatusInvalid,
					result.Error,
				)
			}

			if result.RowsAffected == 0 {
				return customErrors.ErrOrderNotFound
			}
			return nil
		},
	)
}

func (r *OrderRepositoryImpl) AddOrderItem(
	ctx context.Context,
	orderID string,
	item *domain.OrderItem,
	updateStockFunc func(ctx context.Context, productID string, newStock int) error,
	updateOrderTotalFunc func(ctx context.Context, order *domain.Order, price float64) error,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Order]) error {
			if err := updateStockFunc(ctx, item.ProductID.String(), item.Quantity); err != nil {
				return fmt.Errorf(
					"failed to update product stock: %w",
					err,
				)
			}

			if err := txRepo.GetDB().WithContext(ctx).Create(item).Error; err != nil {
				return fmt.Errorf(
					"failed to add order item: %w",
					err,
				)
			}

			var order domain.Order
			if err := txRepo.GetDB().WithContext(ctx).
				Preload("Items").
				First(&order, "id = ?", orderID).Error; err != nil {
				return fmt.Errorf(
					"failed to get order: %w",
					err,
				)
			}

			if err := updateOrderTotalFunc(ctx, &order, item.Price*float64(item.Quantity)); err != nil {
				return fmt.Errorf(
					"failed to update order total: %w",
					err,
				)
			}
			return nil
		},
	)
}

func (r *OrderRepositoryImpl) RemoveOrderItem(
	ctx context.Context,
	orderID, itemID string,
	restoreStockFunc func(ctx context.Context, productID string, quantity int) error,
	updateOrderTotalFunc func(ctx context.Context, order *domain.Order, price float64) error,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Order]) error {
			var order domain.Order
			if err := txRepo.GetDB().WithContext(ctx).
				Preload("Items").
				First(&order, "id = ?", orderID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return customErrors.ErrOrderNotFound
				}

				return fmt.Errorf(
					"failed to get order: %w",
					err,
				)
			}

			if order.Status != domain.OrderStatusPending {
				return customErrors.ErrOrderStatusInvalid
			}

			var itemToRemove *domain.OrderItem
			for _, item := range order.Items {
				if item.ID.String() == itemID {
					itemToRemove = &item
					break
				}
			}

			if itemToRemove == nil {
				return customErrors.ErrOrderItemNotFound
			}

			if err := restoreStockFunc(ctx, itemToRemove.ProductID.String(), itemToRemove.Quantity); err != nil {
				return fmt.Errorf(
					"failed to restore product stock: %w",
					err,
				)
			}

			result := txRepo.GetDB().WithContext(ctx).
				Where("id = ?", itemID).
				Delete(&domain.OrderItem{})
			if result.Error != nil {
				return fmt.Errorf(
					"failed to remove order item: %w",
					result.Error,
				)
			}

			if err := updateOrderTotalFunc(
				ctx,
				&order,
				-(itemToRemove.Price * float64(itemToRemove.Quantity)),
			); err != nil {
				return fmt.Errorf(
					"failed to update order total: %w",
					err,
				)
			}
			return nil
		},
	)
}
