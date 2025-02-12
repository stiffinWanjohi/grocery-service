package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	repository "github.com/grocery-service/internal/repository/postgres"
	"github.com/grocery-service/internal/service/notification"
	customErrors "github.com/grocery-service/utils/errors"
)

type (
	OrderService interface {
		Create(ctx context.Context, order *domain.Order) error
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
			orderID string,
			item *domain.OrderItem,
		) error
		RemoveOrderItem(ctx context.Context, orderID, itemID string) error
	}

	OrderServiceImpl struct {
		repo         repository.OrderRepository
		productRepo  repository.ProductRepository
		customerRepo repository.CustomerRepository
		notifier     notification.NotificationService
	}
)

func NewOrderService(
	repo repository.OrderRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
	notifier notification.NotificationService,
) OrderService {
	return &OrderServiceImpl{
		repo:         repo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
		notifier:     notifier,
	}
}

func (s *OrderServiceImpl) Create(
	ctx context.Context,
	order *domain.Order,
) error {
	if err := order.Validate(); err != nil {
		return fmt.Errorf(
			"%w: %v",
			customErrors.ErrInvalidOrderData,
			err,
		)
	}

	if _, err := s.customerRepo.GetByID(ctx, order.CustomerID.String()); err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}

	var totalPrice float64
	for _, item := range order.Items {
		product, err := s.productRepo.GetByID(
			ctx,
			item.ProductID.String(),
		)
		if err != nil {
			return fmt.Errorf("failed to get product: %w", err)
		}

		if product.Stock < item.Quantity {
			return fmt.Errorf(
				"%w: product %s - requested %d, available %d",
				customErrors.ErrInsufficientStock,
				product.Name,
				item.Quantity,
				product.Stock,
			)
		}

		totalPrice += product.Price * float64(item.Quantity)
	}

	order.TotalPrice = totalPrice
	order.Status = domain.OrderStatusPending

	err := s.repo.Create(
		ctx,
		order,
		func(ctx context.Context, productID string, quantity int) error {
			product, err := s.productRepo.GetByID(ctx, productID)
			if err != nil {
				return err
			}

			newStock := product.Stock - quantity
			return s.productRepo.UpdateStock(ctx, productID, newStock)
		},
	)
	if err != nil {
		return err
	}

	go func() {
		if err := s.notifier.SendOrderConfirmation(context.Background(), order); err != nil {
			fmt.Printf(
				"failed to send order confirmation: %v\n",
				err,
			)
		}
	}()

	return nil
}

func (s *OrderServiceImpl) GetByID(
	ctx context.Context,
	id string,
) (*domain.Order, error) {
	if id == "" {
		return nil, fmt.Errorf(
			"%w: order ID is required",
			customErrors.ErrInvalidOrderData,
		)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *OrderServiceImpl) List(
	ctx context.Context,
) ([]domain.Order, error) {
	return s.repo.List(ctx)
}

func (s *OrderServiceImpl) ListByCustomerID(
	ctx context.Context,
	customerID string,
) ([]domain.Order, error) {
	if customerID == "" {
		return nil, fmt.Errorf(
			"%w: customer ID is required",
			customErrors.ErrInvalidOrderData,
		)
	}
	return s.repo.ListByCustomerID(ctx, customerID)
}

func (s *OrderServiceImpl) Update(
	ctx context.Context,
	order *domain.Order,
) error {
	if err := order.Validate(); err != nil {
		return fmt.Errorf(
			"%w: %v",
			customErrors.ErrInvalidOrderData,
			err,
		)
	}

	existingOrder, err := s.repo.GetByID(
		ctx,
		order.ID.String(),
	)
	if err != nil {
		return fmt.Errorf(
			"failed to get existing order: %w",
			err,
		)
	}

	if existingOrder.Status == domain.OrderStatusDelivered {
		return fmt.Errorf(
			"%w: cannot update delivered order",
			customErrors.ErrOrderStatusInvalid,
		)
	}

	return s.repo.Update(ctx, order)
}

func (s *OrderServiceImpl) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.OrderStatus,
) error {
	if id == "" {
		return fmt.Errorf(
			"%w: order ID is required",
			customErrors.ErrInvalidOrderData,
		)
	}

	if err := domain.ValidateOrderStatus(status); err != nil {
		return fmt.Errorf(
			"%w: %v",
			customErrors.ErrOrderStatusInvalid,
			err,
		)
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !domain.IsValidStatusTransition(
		order.Status,
		status,
	) {
		return fmt.Errorf(
			"%w: invalid transition from %s to %s",
			customErrors.ErrOrderStatusInvalid,
			order.Status,
			status,
		)
	}

	err = s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	order.Status = status

	go func() {
		if err := s.notifier.SendOrderStatusUpdate(context.Background(), order); err != nil {
			fmt.Printf(
				"failed to send order status update: %v\n",
				err,
			)
		}
	}()

	return nil
}

func (s *OrderServiceImpl) AddOrderItem(
	ctx context.Context,
	orderID string,
	item *domain.OrderItem,
) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf(
			"%w: %v",
			customErrors.ErrInvalidOrderItemData,
			err,
		)
	}

	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.OrderStatusPending {
		return fmt.Errorf(
			"%w: can only add items to pending orders",
			customErrors.ErrOrderStatusInvalid,
		)
	}

	product, err := s.productRepo.GetByID(
		ctx,
		item.ProductID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.Stock < item.Quantity {
		return fmt.Errorf(
			"%w: product %s - requested %d, available %d",
			customErrors.ErrInsufficientStock,
			product.Name,
			item.Quantity,
			product.Stock,
		)
	}

	item.Price = product.Price
	item.OrderID = uuid.MustParse(orderID)

	return s.repo.AddOrderItem(
		ctx,
		orderID,
		item,
		func(ctx context.Context, productID string, newStock int) error {
			return s.productRepo.UpdateStock(
				ctx,
				productID,
				newStock,
			)
		},
		func(ctx context.Context, order *domain.Order, price float64) error {
			order.TotalPrice += price
			return s.repo.Update(ctx, order)
		},
	)
}

func (s *OrderServiceImpl) RemoveOrderItem(
	ctx context.Context,
	orderID, itemID string,
) error {
	if orderID == "" || itemID == "" {
		return fmt.Errorf(
			"%w: order ID and item ID are required",
			customErrors.ErrInvalidOrderData,
		)
	}
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.OrderStatusPending {
		return fmt.Errorf(
			"%w: can only remove items from pending orders",
			customErrors.ErrOrderStatusInvalid,
		)
	}

	return s.repo.RemoveOrderItem(
		ctx,
		orderID,
		itemID,
		func(ctx context.Context, productID string, quantity int) error {
			product, err := s.productRepo.GetByID(ctx, productID)
			if err != nil {
				return err
			}

			newStock := product.Stock + quantity
			return s.productRepo.UpdateStock(ctx, productID, newStock)
		},
		func(ctx context.Context, order *domain.Order, price float64) error {
			order.TotalPrice += price
			return s.repo.Update(ctx, order)
		},
	)
}
