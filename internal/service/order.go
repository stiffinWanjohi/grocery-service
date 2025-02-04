package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository"
	"github.com/grocery-service/internal/service/notification"
	"gorm.io/gorm"
)

type OrderService struct {
	repo         repository.OrderRepository
	productRepo  repository.ProductRepository
	customerRepo repository.CustomerRepository
	notifier     notification.NotificationService
	db           *gorm.DB
}

func NewOrderService(
	repo repository.OrderRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
	notifier notification.NotificationService,
	db *gorm.DB,
) *OrderService {
	return &OrderService{
		repo:         repo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
		notifier:     notifier,
		db:           db,
	}
}

func (s *OrderService) Create(ctx context.Context, order *domain.Order) error {
	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	// Verify customer exists
	if _, err := s.customerRepo.GetByID(ctx, order.CustomerID.String()); err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate total price and update product stock
	var totalPrice float64
	for _, item := range order.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID.String())
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get product: %w", err)
		}

		if product.Stock < item.Quantity {
			tx.Rollback()
			return fmt.Errorf("insufficient stock for product %s: requested %d, available %d",
				product.Name, item.Quantity, product.Stock)
		}

		// Update product stock
		newStock := product.Stock - item.Quantity
		if err := s.productRepo.UpdateStock(ctx, product.ID.String(), newStock); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update product stock: %w", err)
		}

		item.Price = product.Price
		totalPrice += product.Price * float64(item.Quantity)
	}

	order.TotalPrice = totalPrice
	order.Status = domain.OrderStatusPending

	if err := s.repo.Create(ctx, order); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create order: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Send notification asynchronously
	go func() {
		if err := s.notifier.SendOrderConfirmation(context.Background(), order); err != nil {
			// TODO: Implement proper error logging
			fmt.Printf("failed to send order confirmation: %v\n", err)
		}
	}()

	return nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (s *OrderService) List(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) ListByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error) {
	orders, err := s.repo.ListByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list customer orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) Update(ctx context.Context, order *domain.Order) error {
	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	existingOrder, err := s.repo.GetByID(ctx, order.ID.String())
	if err != nil {
		return fmt.Errorf("failed to get existing order: %w", err)
	}

	if existingOrder.Status == domain.OrderStatusDelivered {
		return fmt.Errorf("cannot update delivered order")
	}

	return s.repo.Update(ctx, order)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	if err := s.validateOrderStatus(status); err != nil {
		return fmt.Errorf("invalid status: %w", err)
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, status) {
		tx.Rollback()
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, status)
	}

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update order status for notification
	order.Status = status

	// Send notification asynchronously
	go func() {
		if err := s.notifier.SendOrderStatusUpdate(context.Background(), order); err != nil {
			// TODO: Implement proper error logging
			fmt.Printf("failed to send order status update: %v\n", err)
		}
	}()

	return nil
}

func (s *OrderService) AddOrderItem(ctx context.Context, orderID string, item *domain.OrderItem) error {
	if err := s.validateOrderItem(item); err != nil {
		return fmt.Errorf("invalid order item: %w", err)
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.OrderStatusPending {
		tx.Rollback()
		return fmt.Errorf("can only add items to pending orders")
	}

	product, err := s.productRepo.GetByID(ctx, item.ProductID.String())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.Stock < item.Quantity {
		tx.Rollback()
		return fmt.Errorf("insufficient stock for product %s: requested %d, available %d",
			product.Name, item.Quantity, product.Stock)
	}

	item.Price = product.Price
	item.OrderID = uuid.MustParse(orderID)

	if err := s.repo.AddOrderItem(ctx, item); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to add order item: %w", err)
	}

	// Update product stock
	newStock := product.Stock - item.Quantity
	if err := s.productRepo.UpdateStock(ctx, product.ID.String(), newStock); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	// Update order total
	order.TotalPrice += product.Price * float64(item.Quantity)
	if err := s.repo.Update(ctx, order); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order total: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *OrderService) RemoveOrderItem(ctx context.Context, orderID, itemID string) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != domain.OrderStatusPending {
		tx.Rollback()
		return fmt.Errorf("can only remove items from pending orders")
	}

	// Find the item to be removed
	var itemToRemove *domain.OrderItem
	for _, item := range order.Items {
		if item.ID.String() == itemID {
			itemToRemove = &item
			break
		}
	}

	if itemToRemove == nil {
		tx.Rollback()
		return fmt.Errorf("item not found in order")
	}

	// Restore product stock
	product, err := s.productRepo.GetByID(ctx, itemToRemove.ProductID.String())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get product: %w", err)
	}

	newStock := product.Stock + itemToRemove.Quantity
	if err := s.productRepo.UpdateStock(ctx, product.ID.String(), newStock); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to restore product stock: %w", err)
	}

	if err := s.repo.RemoveOrderItem(ctx, orderID, itemID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove order item: %w", err)
	}

	// Update order total
	order.TotalPrice -= itemToRemove.Price * float64(itemToRemove.Quantity)
	if err := s.repo.Update(ctx, order); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order total: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *OrderService) validateOrder(order *domain.Order) error {
	if order.CustomerID.String() == "" {
		return fmt.Errorf("customer ID is required")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	for _, item := range order.Items {
		if err := s.validateOrderItem(&item); err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderService) validateOrderItem(item *domain.OrderItem) error {
	if item.ProductID.String() == "" {
		return fmt.Errorf("product ID is required")
	}
	if item.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	return nil
}

func (s *OrderService) validateOrderStatus(status domain.OrderStatus) error {
	validStatuses := map[domain.OrderStatus]bool{
		domain.OrderStatusPending:   true,
		domain.OrderStatusConfirmed: true,
		domain.OrderStatusPreparing: true,
		domain.OrderStatusReady:     true,
		domain.OrderStatusShipped:   true,
		domain.OrderStatusDelivered: true,
		domain.OrderStatusCancelled: true,
		domain.OrderStatusRefunded:  true,
		domain.OrderStatusFailed:    true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid order status: %s", status)
	}
	return nil
}

func (s *OrderService) isValidStatusTransition(from, to domain.OrderStatus) bool {
	transitions := map[domain.OrderStatus][]domain.OrderStatus{
		domain.OrderStatusPending: {
			domain.OrderStatusConfirmed,
			domain.OrderStatusCancelled,
			domain.OrderStatusFailed,
		},
		domain.OrderStatusConfirmed: {
			domain.OrderStatusPreparing,
			domain.OrderStatusCancelled,
			domain.OrderStatusFailed,
		},
		domain.OrderStatusPreparing: {
			domain.OrderStatusReady,
			domain.OrderStatusCancelled,
			domain.OrderStatusFailed,
		},
		domain.OrderStatusReady: {
			domain.OrderStatusShipped,
			domain.OrderStatusCancelled,
			domain.OrderStatusFailed,
		},
		domain.OrderStatusShipped: {
			domain.OrderStatusDelivered,
			domain.OrderStatusFailed,
		},
		domain.OrderStatusDelivered: {
			domain.OrderStatusRefunded,
		},
		domain.OrderStatusCancelled: {
			domain.OrderStatusRefunded,
		},
		domain.OrderStatusRefunded: {}, // No further transitions
		domain.OrderStatusFailed:   {}, // No further transitions
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
