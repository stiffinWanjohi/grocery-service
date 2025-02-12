package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (*domain.Order, func(context.Context, string, int) error)
		expectedError error
	}{
		{
			name: "Success - Create Order with Items",
			setupTest: func(db *gorm.DB) (*domain.Order, func(context.Context, string, int) error) {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
					Items: []domain.OrderItem{
						{
							ID:        uuid.New(),
							ProductID: product.ID,
							Quantity:  2,
							Price:     product.Price,
						},
					},
				}
				updateStockFunc := func(_ context.Context, _ string, _ int) error {
					return nil
				}
				return order, updateStockFunc
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			order, updateStockFunc := tt.setupTest(postgres.DB)
			err := repo.Create(ctx, order, updateStockFunc)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Order
				err = postgres.DB.Preload("Items").First(&found, "id = ?", order.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, order.CustomerID, found.CustomerID)
				assert.Equal(t, order.Status, found.Status)
				assert.Len(t, found.Items, len(order.Items))
			}
		})
	}
}

func TestOrderRepository_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, *domain.Order)
		expectedError error
	}{
		{
			name: "Success - Get Existing Order",
			setupTest: func(db *gorm.DB) (string, *domain.Order) {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
					Items: []domain.OrderItem{
						{
							ID:        uuid.New(),
							ProductID: product.ID,
							Quantity:  2,
							Price:     product.Price,
						},
					},
				}
				require.NoError(t, db.Create(order).Error)
				return order.ID.String(), order
			},
			expectedError: nil,
		},
		{
			name: "Error - Order Not Found",
			setupTest: func(_ *gorm.DB) (string, *domain.Order) {
				return uuid.New().String(), nil
			},
			expectedError: customErrors.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			id, expected := tt.setupTest(postgres.DB)
			found, err := repo.GetByID(ctx, id)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, expected.ID, found.ID)
				assert.Equal(t, expected.CustomerID, found.CustomerID)
				assert.Len(t, found.Items, len(expected.Items))
			}
		})
	}
}

func TestOrderRepository_List(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) []*domain.Order
		expectedCount int
	}{
		{
			name: "Success - List Multiple Orders",
			setupTest: func(db *gorm.DB) []*domain.Order {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				orders := []*domain.Order{
					{
						ID:         uuid.New(),
						CustomerID: customer.ID,
						Status:     domain.OrderStatusPending,
						Items: []domain.OrderItem{
							{
								ID:        uuid.New(),
								ProductID: product.ID,
								Quantity:  2,
								Price:     product.Price,
							},
						},
					},
					{
						ID:         uuid.New(),
						CustomerID: customer.ID,
						Status:     domain.OrderStatusDelivered,
						Items: []domain.OrderItem{
							{
								ID:        uuid.New(),
								ProductID: product.ID,
								Quantity:  1,
								Price:     product.Price,
							},
						},
					},
				}
				for _, order := range orders {
					require.NoError(t, db.Create(order).Error)
				}
				return orders
			},
			expectedCount: 2,
		},
		{
			name: "Success - Empty List",
			setupTest: func(_ *gorm.DB) []*domain.Order {
				return []*domain.Order{}
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			_ = tt.setupTest(postgres.DB)
			found, err := repo.List(ctx)

			assert.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
		})
	}
}

func TestOrderRepository_ListByCustomerID(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, []*domain.Order)
		expectedCount int
		expectedError error
	}{
		{
			name: "Success - List Customer Orders",
			setupTest: func(db *gorm.DB) (string, []*domain.Order) {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				orders := []*domain.Order{
					{
						ID:         uuid.New(),
						CustomerID: customer.ID,
						Status:     domain.OrderStatusPending,
						Items: []domain.OrderItem{
							{
								ID:        uuid.New(),
								ProductID: product.ID,
								Quantity:  2,
								Price:     product.Price,
							},
						},
					},
					{
						ID:         uuid.New(),
						CustomerID: customer.ID,
						Status:     domain.OrderStatusDelivered,
						Items: []domain.OrderItem{
							{
								ID:        uuid.New(),
								ProductID: product.ID,
								Quantity:  1,
								Price:     product.Price,
							},
						},
					},
				}
				for _, order := range orders {
					require.NoError(t, db.Create(order).Error)
				}
				return customer.ID.String(), orders
			},
			expectedCount: 2,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
				&domain.Category{},
				&domain.Customer{},
				&domain.User{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			customerID, _ := tt.setupTest(postgres.DB)
			found, err := repo.ListByCustomerID(ctx, customerID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, found, tt.expectedCount)
				for _, order := range found {
					assert.Equal(t, customerID, order.CustomerID.String())
				}
			}
		})
	}
}

func TestOrderRepository_Update(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) *domain.Order
		updateFunc    func(*domain.Order)
		expectedError error
	}{
		{
			name: "Success - Update Order Status",
			setupTest: func(db *gorm.DB) *domain.Order {
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
				}
				require.NoError(t, db.Create(order).Error)
				return order
			},
			updateFunc: func(order *domain.Order) {
				order.Status = domain.OrderStatusConfirmed
			},
			expectedError: nil,
		},
		{
			name: "Error - Order Not Found",
			setupTest: func(_ *gorm.DB) *domain.Order {
				return &domain.Order{ID: uuid.New()}
			},
			updateFunc:    func(_ *domain.Order) {},
			expectedError: customErrors.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
				&domain.Category{},
				&domain.Customer{},
				&domain.User{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			order := tt.setupTest(postgres.DB)
			tt.updateFunc(order)

			err := repo.Update(ctx, order)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Order
				err = postgres.DB.First(&found, "id = ?", order.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, order.Status, found.Status)
			}
		})
	}
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) *domain.Order
		newStatus     domain.OrderStatus
		expectedError error
	}{
		{
			name: "Success - Update Status",
			setupTest: func(db *gorm.DB) *domain.Order {
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
				}
				require.NoError(t, db.Create(order).Error)
				return order
			},
			newStatus:     domain.OrderStatusConfirmed,
			expectedError: nil,
		},
		{
			name: "Error - Order Not Found",
			setupTest: func(_ *gorm.DB) *domain.Order {
				return &domain.Order{ID: uuid.New()}
			},
			newStatus:     domain.OrderStatusConfirmed,
			expectedError: customErrors.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
				&domain.Category{},
				&domain.Customer{},
				&domain.User{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			order := tt.setupTest(postgres.DB)
			err := repo.UpdateStatus(ctx, order.ID.String(), tt.newStatus)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Order
				err = postgres.DB.First(&found, "id = ?", order.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.newStatus, found.Status)
			}
		})
	}
}

func TestOrderRepository_AddOrderItem(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (*domain.Order, *domain.OrderItem)
		expectedError error
	}{
		{
			name: "Success - Add Order Item",
			setupTest: func(db *gorm.DB) (*domain.Order, *domain.OrderItem) {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
				}
				require.NoError(t, db.Create(order).Error)

				item := &domain.OrderItem{
					ID:        uuid.New(),
					OrderID:   order.ID,
					ProductID: product.ID,
					Quantity:  2,
					Price:     product.Price,
				}
				return order, item
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
				&domain.Category{},
				&domain.Customer{},
				&domain.User{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			order, item := tt.setupTest(postgres.DB)

			updateStockFunc := func(_ context.Context, _ string, _ int) error {
				return nil
			}

			updateOrderTotalFunc := func(_ context.Context, _ *domain.Order, _ float64) error {
				return nil
			}

			err := repo.AddOrderItem(
				ctx,
				order.ID.String(),
				item,
				updateStockFunc,
				updateOrderTotalFunc,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Order
				err = postgres.DB.Preload("Items").
					First(&found, "id = ?", order.ID).Error
				assert.NoError(t, err)
				assert.Len(t, found.Items, 1)
			}
		})
	}
}

func TestOrderRepository_RemoveOrderItem(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (*domain.Order, string)
		expectedError error
	}{
		{
			name: "Success - Remove Order Item",
			setupTest: func(db *gorm.DB) (*domain.Order, string) {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
					Items: []domain.OrderItem{
						{
							ID:        uuid.New(),
							ProductID: product.ID,
							Quantity:  2,
							Price:     product.Price,
						},
					},
				}
				require.NoError(t, db.Create(order).Error)
				return order, order.Items[0].ID.String()
			},
			expectedError: nil,
		},
		{
			name: "Error - Order Item Not Found",
			setupTest: func(db *gorm.DB) (*domain.Order, string) {
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusPending,
				}
				require.NoError(t, db.Create(order).Error)
				return order, uuid.New().String()
			},
			expectedError: customErrors.ErrOrderItemNotFound,
		},
		{
			name: "Error - Order Not Found",
			setupTest: func(db *gorm.DB) (*domain.Order, string) {
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: uuid.New(),
				}
				return order, uuid.New().String()
			},
			expectedError: customErrors.ErrOrderNotFound,
		},
		{
			name: "Error - Invalid Order Status",
			setupTest: func(db *gorm.DB) (*domain.Order, string) {
				product := createTestProduct(t, db)
				customer := createTestCustomer(t, db)
				order := &domain.Order{
					ID:         uuid.New(),
					CustomerID: customer.ID,
					Status:     domain.OrderStatusDelivered,
					Items: []domain.OrderItem{
						{
							ID:        uuid.New(),
							ProductID: product.ID,
							Quantity:  2,
							Price:     product.Price,
						},
					},
				}
				require.NoError(t, db.Create(order).Error)
				return order, order.Items[0].ID.String()
			},
			expectedError: customErrors.ErrOrderStatusInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(
				t,
				&domain.OrderItem{},
				&domain.Order{},
				&domain.Product{},
				&domain.Category{},
				&domain.Customer{},
				&domain.User{},
			)
			repo := NewOrderRepository(postgres)
			ctx := context.Background()

			order, itemID := tt.setupTest(postgres.DB)

			restoreStockFunc := func(_ context.Context, _ string, _ int) error {
				return nil
			}

			updateOrderTotalFunc := func(_ context.Context, _ *domain.Order, _ float64) error {
				return nil
			}

			err := repo.RemoveOrderItem(
				ctx,
				order.ID.String(),
				itemID,
				restoreStockFunc,
				updateOrderTotalFunc,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Order
				err = postgres.DB.Preload("Items").First(&found, "id = ?", order.ID).Error
				assert.NoError(t, err)
				assert.Empty(t, found.Items)
			}
		})
	}
}
