package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOrderTestDB(t *testing.T) *db.PostgresDB {
	postgres, err := db.NewTestDB()
	require.NoError(t, err)

	err = postgres.DB.Migrator().DropTable(&domain.OrderItem{}, &domain.Order{}, &domain.Product{})
	require.NoError(t, err)
	err = postgres.DB.AutoMigrate(&domain.Product{}, &domain.Order{}, &domain.OrderItem{})
	require.NoError(t, err)

	return postgres
}

func createTestProduct(t *testing.T, db *gorm.DB) *domain.Product {
	product := &domain.Product{
		ID:    uuid.New(),
		Name:  "Test Product",
		Price: 10.00,
		Stock: 100,
	}
	require.NoError(t, db.Create(product).Error)
	return product
}

func TestOrderRepository_Create(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	product := createTestProduct(t, postgres.DB)
	customerID := uuid.New()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: customerID,
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

	updateStockFunc := func(ctx context.Context, productID string, newStock int) error {
		return nil
	}

	err := repo.Create(ctx, order, updateStockFunc)
	assert.NoError(t, err)

	var found domain.Order
	err = postgres.DB.Preload("Items").First(&found, "id = ?", order.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, order.CustomerID, found.CustomerID)
	assert.Equal(t, order.Status, found.Status)
	assert.Len(t, found.Items, 1)
}

func TestOrderRepository_GetByID(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	product := createTestProduct(t, postgres.DB)
	customerID := uuid.New()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: customerID,
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

	err := postgres.DB.Create(order).Error
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, order.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)
	assert.Equal(t, order.CustomerID, found.CustomerID)
	assert.Len(t, found.Items, 1)

	_, err = repo.GetByID(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrOrderNotFound)
}

func TestOrderRepository_List(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	product := createTestProduct(t, postgres.DB)
	customerID := uuid.New()

	orders := []*domain.Order{
		{
			ID:         uuid.New(),
			CustomerID: customerID,
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
			CustomerID: customerID,
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
		err := postgres.DB.Create(order).Error
		require.NoError(t, err)
	}

	found, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestOrderRepository_ListByCustomerID(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	product := createTestProduct(t, postgres.DB)
	customerID := uuid.New()

	orders := []*domain.Order{
		{
			ID:         uuid.New(),
			CustomerID: customerID,
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
			CustomerID: customerID,
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
		err := postgres.DB.Create(order).Error
		require.NoError(t, err)
	}

	found, err := repo.ListByCustomerID(ctx, customerID.String())
	assert.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestOrderRepository_Update(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		Status:     domain.OrderStatusPending,
	}

	err := postgres.DB.Create(order).Error
	require.NoError(t, err)

	order.Status = domain.OrderStatusConfirmed
	err = repo.Update(ctx, order)
	assert.NoError(t, err)

	var found domain.Order
	err = postgres.DB.First(&found, "id = ?", order.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.OrderStatusConfirmed, found.Status)

	nonExistent := &domain.Order{ID: uuid.New()}
	err = repo.Update(ctx, nonExistent)
	assert.ErrorIs(t, err, customErrors.ErrOrderNotFound)
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		Status:     domain.OrderStatusPending,
	}

	err := postgres.DB.Create(order).Error
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, order.ID.String(), domain.OrderStatusConfirmed)
	assert.NoError(t, err)

	var found domain.Order
	err = postgres.DB.First(&found, "id = ?", order.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.OrderStatusConfirmed, found.Status)

	err = repo.UpdateStatus(ctx, uuid.New().String(), domain.OrderStatusConfirmed)
	assert.ErrorIs(t, err, customErrors.ErrOrderNotFound)
}

func TestOrderRepository_AddOrderItem(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	product := createTestProduct(t, postgres.DB)
	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
		Status:     domain.OrderStatusPending,
	}

	err := postgres.DB.Create(order).Error
	require.NoError(t, err)

	item := &domain.OrderItem{
		ID:        uuid.New(),
		OrderID:   order.ID,
		ProductID: product.ID,
		Quantity:  2,
		Price:     product.Price,
	}

	updateStockFunc := func(ctx context.Context, productID string, newStock int) error {
		return nil
	}

	updateOrderTotalFunc := func(ctx context.Context, order *domain.Order, price float64) error {
		return nil
	}

	err = repo.AddOrderItem(ctx, order.ID.String(), item, updateStockFunc, updateOrderTotalFunc)
	assert.NoError(t, err)

	var found domain.Order
	err = postgres.DB.Preload("Items").First(&found, "id = ?", order.ID).Error
	assert.NoError(t, err)
	assert.Len(t, found.Items, 1)
}

func TestOrderRepository_RemoveOrderItem(t *testing.T) {
	postgres := setupOrderTestDB(t)
	repo := NewOrderRepository(postgres)
	ctx := context.Background()

	product := createTestProduct(t, postgres.DB)
	order := &domain.Order{
		ID:         uuid.New(),
		CustomerID: uuid.New(),
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

	err := postgres.DB.Create(order).Error
	require.NoError(t, err)

	restoreStockFunc := func(ctx context.Context, productID string, quantity int) error {
		return nil
	}

	updateOrderTotalFunc := func(ctx context.Context, order *domain.Order, price float64) error {
		return nil
	}

	err = repo.RemoveOrderItem(ctx, order.ID.String(), order.Items[0].ID.String(), restoreStockFunc, updateOrderTotalFunc)
	assert.NoError(t, err)

	var found domain.Order
	err = postgres.DB.Preload("Items").First(&found, "id = ?", order.ID).Error
	assert.NoError(t, err)
	assert.Len(t, found.Items, 0)

	err = repo.RemoveOrderItem(ctx, order.ID.String(), uuid.New().String(), restoreStockFunc, updateOrderTotalFunc)
	assert.ErrorIs(t, err, customErrors.ErrOrderItemNotFound)
}
