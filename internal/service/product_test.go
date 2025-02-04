package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	mocks "github.com/grocery-service/tests/mocks/repository"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
)

func TestProductService_Create(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &domain.Product{
		ID:         uuid.New(),
		Name:       "Test Product",
		Price:      10.99,
		Stock:      100,
		CategoryID: uuid.New(),
	}

	mockProductRepo.On("Create", ctx, product).Return(nil)

	err := service.Create(ctx, product)
	assert.NoError(t, err)

	// Test validation errors
	testCases := []struct {
		name    string
		product *domain.Product
		errMsg  string
	}{
		{
			name:    "Empty name",
			product: &domain.Product{ID: uuid.New(), Price: 10.99, Stock: 100},
			errMsg:  "product name is required",
		},
		{
			name:    "Zero price",
			product: &domain.Product{ID: uuid.New(), Name: "Test", Price: 0, Stock: 100},
			errMsg:  "product price must be greater than zero",
		},
		{
			name:    "Negative price",
			product: &domain.Product{ID: uuid.New(), Name: "Test", Price: -10, Stock: 100},
			errMsg:  "product price must be greater than zero",
		},
		{
			name:    "Negative stock",
			product: &domain.Product{ID: uuid.New(), Name: "Test", Price: 10.99, Stock: -1},
			errMsg:  "product stock cannot be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.Create(ctx, tc.product)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestProductService_GetByID(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &domain.Product{
		ID:         uuid.New(),
		Name:       "Test Product",
		Price:      10.99,
		Stock:      100,
		CategoryID: uuid.New(),
	}

	mockProductRepo.On("GetByID", ctx, product.ID.String()).Return(product, nil)
	mockProductRepo.On("GetByID", ctx, "non-existent").Return(nil, customErrors.ErrProductNotFound)

	found, err := service.GetByID(ctx, product.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, product.ID, found.ID)
	assert.Equal(t, product.Name, found.Name)

	_, err = service.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, customErrors.ErrProductNotFound)
}

func TestProductService_List(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	products := []domain.Product{
		{ID: uuid.New(), Name: "Product 1", Price: 10.99, Stock: 100},
		{ID: uuid.New(), Name: "Product 2", Price: 20.99, Stock: 200},
	}

	mockProductRepo.On("List", ctx).Return(products, nil)

	found, err := service.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, len(products))
}

func TestProductService_ListByCategoryID(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	categoryID := uuid.New()
	products := []domain.Product{
		{ID: uuid.New(), Name: "Product 1", Price: 10.99, Stock: 100, CategoryID: categoryID},
		{ID: uuid.New(), Name: "Product 2", Price: 20.99, Stock: 200, CategoryID: categoryID},
	}

	mockProductRepo.On("ListByCategoryID", ctx, categoryID.String()).Return(products, nil)

	found, err := service.ListByCategoryID(ctx, categoryID.String())
	assert.NoError(t, err)
	assert.Len(t, found, len(products))
}

func TestProductService_Update(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &domain.Product{
		ID:         uuid.New(),
		Name:       "Test Product",
		Price:      10.99,
		Stock:      100,
		CategoryID: uuid.New(),
	}

	mockProductRepo.On("Update", ctx, product).Return(nil)

	err := service.Update(ctx, product)
	assert.NoError(t, err)

	invalidProduct := &domain.Product{
		ID: uuid.New(),
	}
	err = service.Update(ctx, invalidProduct)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product name is required")
}

func TestProductService_UpdateStock(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	productID := uuid.New().String()

	mockProductRepo.On("UpdateStock", ctx, productID, 50).Return(nil)
	err := service.UpdateStock(ctx, productID, 50)
	assert.NoError(t, err)

	err = service.UpdateStock(ctx, productID, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity cannot be negative")
}

func TestProductService_Delete(t *testing.T) {
	mockProductRepo := mocks.NewProductRepository(t)
	mockCategoryRepo := mocks.NewCategoryRepository(t)
	service := NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.Background()

	id := uuid.New().String()

	mockProductRepo.On("Delete", ctx, id).Return(nil)
	mockProductRepo.On("Delete", ctx, "non-existent").Return(customErrors.ErrProductNotFound)

	err := service.Delete(ctx, id)
	assert.NoError(t, err)

	err = service.Delete(ctx, "non-existent")
	assert.ErrorIs(t, err, customErrors.ErrProductNotFound)
}
