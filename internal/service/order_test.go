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

func createTestCustomer() (*domain.Customer, *domain.User) {
	userID := uuid.New()
	customerID := uuid.New()

	user := &domain.User{
		ID:      userID,
		Name:    "Test Customer",
		Email:   "test@example.com",
		Role:    domain.CustomerRole,
		Picture: "https://example.com/picture.jpg",
	}

	customer := &domain.Customer{
		ID:     customerID,
		UserID: userID,
		User:   user,
	}

	return customer, user
}

func createTestProduct() *domain.Product {
	return &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       10.0,
		Stock:       5,
	}
}

func createTestOrder(customerID uuid.UUID) *domain.Order {
	return &domain.Order{
		ID:         uuid.New(),
		CustomerID: customerID,
		Status:     domain.OrderStatusPending,
		Items: []domain.OrderItem{
			{
				ID:        uuid.New(),
				ProductID: uuid.New(),
				Quantity:  2,
				Price:     10.0,
			},
		},
	}
}

func TestOrderService_Create(t *testing.T) {
	tests := []struct {
		name          string
		setupOrder    func() *domain.Order
		setupMocks    func(*repoMocks.OrderRepository, *repoMocks.ProductRepository, *repoMocks.CustomerRepository, *serviceMock.NotificationService, *domain.Order)
		expectedError error
	}{
		{
			name: "Success - Create Order",
			setupOrder: func() *domain.Order {
				customer, _ := createTestCustomer()
				product := createTestProduct()
				order := createTestOrder(customer.ID)
				order.Items[0].ProductID = product.ID
				return order
			},
			setupMocks: func(or *repoMocks.OrderRepository, pr *repoMocks.ProductRepository, cr *repoMocks.CustomerRepository, ns *serviceMock.NotificationService, order *domain.Order) {
				customer := &domain.Customer{ID: order.CustomerID}
				product := &domain.Product{
					ID:    order.Items[0].ProductID,
					Price: 10.0,
					Stock: 5,
				}

				cr.On("GetByID", mock.Anything, order.CustomerID.String()).Return(customer, nil)
				pr.On("GetByID", mock.Anything, order.Items[0].ProductID.String()).Return(product, nil)
				pr.On("UpdateStock", mock.Anything, order.Items[0].ProductID.String(), 3).Return(nil)
				or.On("Create", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
					return o.ID == order.ID && o.Status == domain.OrderStatusPending
				})).Return(nil)
				ns.On("SendOrderConfirmation", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
					return o.ID == order.ID
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Error - Invalid Order Data",
			setupOrder: func() *domain.Order {
				return &domain.Order{}
			},
			setupMocks: func(or *repoMocks.OrderRepository, pr *repoMocks.ProductRepository, cr *repoMocks.CustomerRepository, ns *serviceMock.NotificationService, order *domain.Order) {
			},
			expectedError: customErrors.ErrInvalidOrderData,
		},
		{
			name: "Error - Insufficient Stock",
			setupOrder: func() *domain.Order {
				customer, _ := createTestCustomer()
				product := createTestProduct()
				product.Stock = 1
				order := createTestOrder(customer.ID)
				order.Items[0].ProductID = product.ID
				order.Items[0].Quantity = 2
				return order
			},
			setupMocks: func(or *repoMocks.OrderRepository, pr *repoMocks.ProductRepository, cr *repoMocks.CustomerRepository, ns *serviceMock.NotificationService, order *domain.Order) {
				customer := &domain.Customer{ID: order.CustomerID}
				product := &domain.Product{
					ID:    order.Items[0].ProductID,
					Price: 10.0,
					Stock: 1,
				}

				cr.On("GetByID", mock.Anything, order.CustomerID.String()).Return(customer, nil)
				pr.On("GetByID", mock.Anything, order.Items[0].ProductID.String()).Return(product, nil)
			},
			expectedError: customErrors.ErrInsufficientStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, productRepo, customerRepo, notifier := setupOrderTest(t)
			order := tt.setupOrder()
			tt.setupMocks(orderRepo, productRepo, customerRepo, notifier, order)

			err := service.Create(context.Background(), order)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, domain.OrderStatusPending, order.Status)
				assert.Greater(t, order.TotalPrice, 0.0)
			}
		})
	}
}

func TestOrderService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		setupMocks    func(*repoMocks.OrderRepository, string)
		expectedError error
	}{
		{
			name:    "Success - Get Existing Order",
			orderID: uuid.New().String(),
			setupMocks: func(or *repoMocks.OrderRepository, orderID string) {
				customer, _ := createTestCustomer()
				order := createTestOrder(customer.ID)
				order.ID = uuid.MustParse(orderID)
				or.On("GetByID", mock.Anything, orderID).Return(order, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Error - Empty OrderID",
			orderID:       "",
			setupMocks:    func(or *repoMocks.OrderRepository, orderID string) {},
			expectedError: customErrors.ErrInvalidOrderData,
		},
		{
			name:    "Error - Order Not Found",
			orderID: uuid.New().String(),
			setupMocks: func(or *repoMocks.OrderRepository, orderID string) {
				or.On("GetByID", mock.Anything, orderID).Return(nil, customErrors.ErrOrderNotFound)
			},
			expectedError: customErrors.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, _, _, _ := setupOrderTest(t)
			tt.setupMocks(orderRepo, tt.orderID)

			order, err := service.GetByID(context.Background(), tt.orderID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.orderID, order.ID.String())
			}
		})
	}
}

func TestOrderService_List(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*repoMocks.OrderRepository)
		expectedCount int
		expectedError error
	}{
		{
			name: "Success - List Multiple Orders",
			setupMocks: func(or *repoMocks.OrderRepository) {
				customer, _ := createTestCustomer()
				orders := []domain.Order{
					*createTestOrder(customer.ID),
					*createTestOrder(customer.ID),
				}
				or.On("List", mock.Anything).Return(orders, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "Success - Empty List",
			setupMocks: func(or *repoMocks.OrderRepository) {
				or.On("List", mock.Anything).Return([]domain.Order{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, _, _, _ := setupOrderTest(t)
			tt.setupMocks(orderRepo)

			orders, err := service.List(context.Background())

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, orders)
			} else {
				assert.NoError(t, err)
				assert.Len(t, orders, tt.expectedCount)
			}
		})
	}
}

func TestOrderService_ListByCustomerID(t *testing.T) {
	tests := []struct {
		name          string
		customerID    string
		setupMocks    func(*repoMocks.OrderRepository, *repoMocks.CustomerRepository, string)
		expectedCount int
		expectedError error
	}{
		{
			name:       "Success - List Customer Orders",
			customerID: uuid.New().String(),
			setupMocks: func(or *repoMocks.OrderRepository, cr *repoMocks.CustomerRepository, customerID string) {
				customer, _ := createTestCustomer()
				customer.ID = uuid.MustParse(customerID)
				orders := []domain.Order{
					*createTestOrder(customer.ID),
					*createTestOrder(customer.ID),
				}
				cr.On("GetByID", mock.Anything, customerID).Return(customer, nil)
				or.On("ListByCustomerID", mock.Anything, customerID).Return(orders, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:          "Error - Empty CustomerID",
			customerID:    "",
			setupMocks:    func(or *repoMocks.OrderRepository, cr *repoMocks.CustomerRepository, customerID string) {},
			expectedCount: 0,
			expectedError: customErrors.ErrInvalidOrderData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, _, customerRepo, _ := setupOrderTest(t)
			tt.setupMocks(orderRepo, customerRepo, tt.customerID)

			orders, err := service.ListByCustomerID(context.Background(), tt.customerID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, orders)
			} else {
				assert.NoError(t, err)
				assert.Len(t, orders, tt.expectedCount)
			}
		})
	}
}

func TestOrderService_Update(t *testing.T) {
	tests := []struct {
		name          string
		setupOrder    func() *domain.Order
		setupMocks    func(*repoMocks.OrderRepository, *domain.Order)
		expectedError error
	}{
		{
			name: "Success - Update Pending Order",
			setupOrder: func() *domain.Order {
				customer, _ := createTestCustomer()
				return createTestOrder(customer.ID)
			},
			setupMocks: func(or *repoMocks.OrderRepository, order *domain.Order) {
				existingOrder := &domain.Order{
					ID:     order.ID,
					Status: domain.OrderStatusPending,
				}
				or.On("GetByID", mock.Anything, order.ID.String()).Return(existingOrder, nil)
				or.On("Update", mock.Anything, order).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Error - Update Delivered Order",
			setupOrder: func() *domain.Order {
				customer, _ := createTestCustomer()
				return createTestOrder(customer.ID)
			},
			setupMocks: func(or *repoMocks.OrderRepository, order *domain.Order) {
				existingOrder := &domain.Order{
					ID:     order.ID,
					Status: domain.OrderStatusDelivered,
				}
				or.On("GetByID", mock.Anything, order.ID.String()).Return(existingOrder, nil)
			},
			expectedError: customErrors.ErrOrderStatusInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, _, _, _ := setupOrderTest(t)
			order := tt.setupOrder()
			tt.setupMocks(orderRepo, order)

			err := service.Update(context.Background(), order)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		fromStatus    domain.OrderStatus
		toStatus      domain.OrderStatus
		setupMocks    func(*repoMocks.OrderRepository, *serviceMock.NotificationService, string, domain.OrderStatus)
		expectedError error
	}{
		{
			name:       "Success - Pending to Confirmed",
			orderID:    uuid.New().String(),
			fromStatus: domain.OrderStatusPending,
			toStatus:   domain.OrderStatusConfirmed,
			setupMocks: func(or *repoMocks.OrderRepository, ns *serviceMock.NotificationService, orderID string, toStatus domain.OrderStatus) {
				order := &domain.Order{
					ID:     uuid.MustParse(orderID),
					Status: domain.OrderStatusPending,
				}
				or.On("GetByID", mock.Anything, orderID).Return(order, nil)
				or.On("UpdateStatus", mock.Anything, orderID, toStatus).Return(nil)
				ns.On("SendOrderStatusUpdate", mock.Anything, mock.MatchedBy(func(o *domain.Order) bool {
					return o.ID.String() == orderID && o.Status == toStatus
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "Error - Invalid Status Transition",
			orderID:    uuid.New().String(),
			fromStatus: domain.OrderStatusPending,
			toStatus:   domain.OrderStatusDelivered,
			setupMocks: func(or *repoMocks.OrderRepository, ns *serviceMock.NotificationService, orderID string, _ domain.OrderStatus) {
				order := &domain.Order{
					ID:     uuid.MustParse(orderID),
					Status: domain.OrderStatusPending,
				}
				or.On("GetByID", mock.Anything, orderID).Return(order, nil)
			},
			expectedError: customErrors.ErrOrderStatusInvalid,
		},
		{
			name:       "Error - Empty OrderID",
			orderID:    "",
			fromStatus: domain.OrderStatusPending,
			toStatus:   domain.OrderStatusConfirmed,
			setupMocks: func(or *repoMocks.OrderRepository, ns *serviceMock.NotificationService, orderID string, _ domain.OrderStatus) {
			},
			expectedError: customErrors.ErrInvalidOrderData,
		},
		{
			name:       "Error - Order Not Found",
			orderID:    uuid.New().String(),
			fromStatus: domain.OrderStatusPending,
			toStatus:   domain.OrderStatusConfirmed,
			setupMocks: func(or *repoMocks.OrderRepository, ns *serviceMock.NotificationService, orderID string, _ domain.OrderStatus) {
				or.On("GetByID", mock.Anything, orderID).Return(nil, customErrors.ErrOrderNotFound)
			},
			expectedError: customErrors.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, _, _, notifier := setupOrderTest(t)
			tt.setupMocks(orderRepo, notifier, tt.orderID, tt.toStatus)

			err := service.UpdateStatus(context.Background(), tt.orderID, tt.toStatus)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderService_RemoveOrderItemFromNonPendingOrder(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	customer, _ := createTestCustomer()
	orderID := uuid.New()
	itemID := uuid.New()

	order := &domain.Order{
		ID:         orderID,
		CustomerID: customer.ID,
		Status:     domain.OrderStatusConfirmed,
	}

	orderRepo.On("GetByID", ctx, orderID.String()).Return(order, nil)

	err := service.RemoveOrderItem(ctx, orderID.String(), itemID.String())
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrOrderStatusInvalid)
}

func TestOrderService_RemoveOrderItemNotFound(t *testing.T) {
	service, orderRepo, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	customer, _ := createTestCustomer()
	orderID := uuid.New()
	itemID := uuid.New()

	order := &domain.Order{
		ID:         orderID,
		CustomerID: customer.ID,
		Status:     domain.OrderStatusPending,
		Items:      []domain.OrderItem{},
	}

	orderRepo.On("GetByID", ctx, orderID.String()).Return(order, nil)

	err := service.RemoveOrderItem(ctx, orderID.String(), itemID.String())
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrOrderItemNotFound)
}

func TestOrderService_RemoveOrderItemInvalidInput(t *testing.T) {
	service, _, _, _, _ := setupOrderTest(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		orderID  string
		itemID   string
		expected error
	}{
		{
			name:     "Empty OrderID",
			orderID:  "",
			itemID:   uuid.New().String(),
			expected: customErrors.ErrInvalidOrderData,
		},
		{
			name:     "Empty ItemID",
			orderID:  uuid.New().String(),
			itemID:   "",
			expected: customErrors.ErrInvalidOrderData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RemoveOrderItem(ctx, tt.orderID, tt.itemID)
			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expected)
		})
	}
}
