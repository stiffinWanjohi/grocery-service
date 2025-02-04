package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	repoMocks "github.com/grocery-service/tests/mocks/repository"
	serviceMock "github.com/grocery-service/tests/mocks/service"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupOrderTest(t *testing.T) (OrderService, *repoMocks.OrderRepository, *repoMocks.ProductRepository, *repoMocks.CustomerRepository, *serviceMock.NotificationService) {
	orderRepo := repoMocks.NewOrderRepository(t)
	productRepo := repoMocks.NewProductRepository(t)
	customerRepo := repoMocks.NewCustomerRepository(t)
	notifier := serviceMock.NewNotificationService(t)

	service := NewOrderService(orderRepo, productRepo, customerRepo, notifier)

	return service, orderRepo, productRepo, customerRepo, notifier
}

func TestOrderService_Create(t *testing.T) {
	service, orderRepo, productRepo, customerRepo, notifier := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	productID := uuid.New()
	orderID := uuid.New()

	customer := &domain.Customer{ID: customerID, Name: "Test Customer"}
	product := &domain.Product{ID: productID, Name: "Test Product", Price: 10.0, Stock: 5}

	order := &domain.Order{
		ID:         orderID,
		CustomerID: customerID,
		Items: []domain.OrderItem{
			{
				ID:        uuid.New(),
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	customerRepo.On("GetByID", ctx, customerID.String()).Return(customer, nil)
	productRepo.On("GetByID", ctx, productID.String()).Return(product, nil)
	productRepo.On("UpdateStock", ctx, productID.String(), 3).Return(nil)
	orderRepo.On("Create", ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID && o.Status == domain.OrderStatusPending && o.TotalPrice == 20.0
	}), mock.Anything).Return(nil)
	notifier.On("SendOrderConfirmation", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID
	})).Return(nil)

	err := service.Create(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, 20.0, order.TotalPrice) // 2 items * $10.00
}

func TestOrderService_CreateValidationError(t *testing.T) {
	service, _, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	invalidOrder := &domain.Order{} // deliberately invalid order

	err := service.Create(ctx, invalidOrder)
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrInvalidOrderData)
}

func TestOrderService_CreateInsufficientStock(t *testing.T) {
	service, _, productRepo, customerRepo, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	productID := uuid.New()

	customer := &domain.Customer{ID: customerID, Name: "Test Customer"}
	product := &domain.Product{ID: productID, Name: "Test Product", Price: 10.0, Stock: 1}

	order := &domain.Order{
		CustomerID: customerID,
		Items: []domain.OrderItem{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	customerRepo.On("GetByID", ctx, customerID.String()).Return(customer, nil)
	productRepo.On("GetByID", ctx, productID.String()).Return(product, nil)

	err := service.Create(ctx, order)
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrInsufficientStock)
}

func TestOrderService_GetByID(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		Status:     domain.OrderStatusPending,
	}

	orderRepo.On("GetByID", ctx, order.ID.String()).Return(order, nil)
	orderRepo.On("GetByID", ctx, "non-existent").Return(nil, customErrors.ErrOrderNotFound)

	found, err := service.GetByID(ctx, order.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)

	_, err = service.GetByID(ctx, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrInvalidOrderData)

	_, err = service.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, customErrors.ErrOrderNotFound)
}

func TestOrderService_List(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	expectedOrders := []domain.Order{
		{ID: uuid.New(), Status: domain.OrderStatusPending},
		{ID: uuid.New(), Status: domain.OrderStatusConfirmed},
	}

	orderRepo.On("List", ctx).Return(expectedOrders, nil)

	orders, err := service.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
}

func TestOrderService_ListByCustomerID(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	customerID := uuid.New()
	expectedOrders := []domain.Order{
		{ID: uuid.New(), CustomerID: customerID, Status: domain.OrderStatusPending},
		{ID: uuid.New(), CustomerID: customerID, Status: domain.OrderStatusConfirmed},
	}

	orderRepo.On("ListByCustomerID", ctx, customerID.String()).Return(expectedOrders, nil)

	orders, err := service.ListByCustomerID(ctx, customerID.String())
	assert.NoError(t, err)
	assert.Len(t, orders, 2)

	_, err = service.ListByCustomerID(ctx, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrInvalidOrderData)
}

func TestOrderService_Update(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		Status:     domain.OrderStatusPending,
		Items:      []domain.OrderItem{{ProductID: uuid.New(), Quantity: 1}},
	}

	existingOrder := &domain.Order{
		ID:     order.ID,
		Status: domain.OrderStatusPending,
	}

	orderRepo.On("GetByID", ctx, order.ID.String()).Return(existingOrder, nil)
	orderRepo.On("Update", ctx, order).Return(nil)

	err := service.Update(ctx, order)
	assert.NoError(t, err)
}

func TestOrderService_UpdateDeliveredOrderFails(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		Status:     domain.OrderStatusPending,
	}

	existingOrder := &domain.Order{
		ID:     order.ID,
		Status: domain.OrderStatusDelivered,
	}

	orderRepo.On("GetByID", ctx, order.ID.String()).Return(existingOrder, nil)

	err := service.Update(ctx, order)
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrOrderStatusInvalid)
}

func TestOrderService_UpdateStatus(t *testing.T) {
	service, orderRepo, _, _, notifier := setupOrderTest(t)
	ctx := context.Background()

	order := &domain.Order{
		ID:     uuid.New(),
		Status: domain.OrderStatusPending,
	}

	testCases := []struct {
		name          string
		fromStatus    domain.OrderStatus
		toStatus      domain.OrderStatus
		expectSuccess bool
	}{
		{
			name:          "Pending to Confirmed",
			fromStatus:    domain.OrderStatusPending,
			toStatus:      domain.OrderStatusConfirmed,
			expectSuccess: true,
		},
		{
			name:          "Pending to Delivered",
			fromStatus:    domain.OrderStatusPending,
			toStatus:      domain.OrderStatusDelivered,
			expectSuccess: false,
		},
		{
			name:          "Confirmed to Preparing",
			fromStatus:    domain.OrderStatusConfirmed,
			toStatus:      domain.OrderStatusPreparing,
			expectSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order.Status = tc.fromStatus
			orderRepo.On("GetByID", ctx, order.ID.String()).Return(order, nil).Once()

			if tc.expectSuccess {
				orderRepo.On("UpdateStatus", ctx, order.ID.String(), tc.toStatus).Return(nil).Once()
				notifier.On("SendOrderStatusUpdate", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
					return o.ID == order.ID && o.Status == tc.toStatus
				})).Return(nil).Once()
			}

			err := service.UpdateStatus(ctx, order.ID.String(), tc.toStatus)
			if tc.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestOrderService_AddOrderItem(t *testing.T) {
	service, orderRepo, productRepo, _, _ := setupOrderTest(t)
	ctx := context.Background()

	orderID := uuid.New()
	productID := uuid.New()

	order := &domain.Order{
		ID:         orderID,
		Status:     domain.OrderStatusPending,
		TotalPrice: 100.0,
	}

	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 10.0,
		Stock: 5,
	}

	item := &domain.OrderItem{
		ID:        uuid.New(),
		ProductID: productID,
		Quantity:  2,
	}

	orderRepo.On("GetByID", ctx, orderID.String()).Return(order, nil)
	productRepo.On("GetByID", ctx, productID.String()).Return(product, nil)
	productRepo.On("UpdateStock", ctx, productID.String(), 3).Return(nil)
	orderRepo.On("AddOrderItem", ctx, orderID.String(), mock.MatchedBy(func(i *domain.OrderItem) bool {
		return i.OrderID == orderID && i.Price == product.Price
	}), mock.Anything, mock.Anything).Return(nil)

	err := service.AddOrderItem(ctx, orderID.String(), item)
	assert.NoError(t, err)
}

func TestOrderService_AddOrderItemToNonPendingOrder(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	orderID := uuid.New()
	productID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusConfirmed,
	}

	item := &domain.OrderItem{
		ProductID: productID,
		Quantity:  2,
	}

	orderRepo.On("GetByID", ctx, orderID.String()).Return(order, nil)

	err := service.AddOrderItem(ctx, orderID.String(), item)
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrOrderStatusInvalid)
}

func TestOrderService_RemoveOrderItem(t *testing.T) {
	service, orderRepo, productRepo, _, _ := setupOrderTest(t)
	ctx := context.Background()

	orderID := uuid.New()
	productID := uuid.New()
	itemID := uuid.New()

	order := &domain.Order{
		ID:         orderID,
		Status:     domain.OrderStatusPending,
		TotalPrice: 100.0,
		Items: []domain.OrderItem{
			{
				ID:        itemID,
				ProductID: productID,
				Quantity:  2,
				Price:     10.0,
			},
		},
	}

	product := &domain.Product{
		ID:    productID,
		Stock: 5,
	}

	orderRepo.On("GetByID", ctx, orderID.String()).Return(order, nil)
	productRepo.On("GetByID", ctx, productID.String()).Return(product, nil)
	productRepo.On("UpdateStock", ctx, productID.String(), 7).Return(nil)
	orderRepo.On("RemoveOrderItem", ctx, orderID.String(), itemID.String(), mock.Anything, mock.Anything).Return(nil)

	err := service.RemoveOrderItem(ctx, orderID.String(), itemID.String())
	assert.NoError(t, err)
}

func TestOrderService_RemoveOrderItemFromNonPendingOrder(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	orderID := uuid.New()
	itemID := uuid.New()

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusConfirmed,
	}

	orderRepo.On("GetByID", ctx, orderID.String()).Return(order, nil)

	err := service.RemoveOrderItem(ctx, orderID.String(), itemID.String())
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrOrderStatusInvalid)
}
