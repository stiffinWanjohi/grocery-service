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

func setupProductTestDB(t *testing.T) *db.PostgresDB {
	postgres, err := db.NewTestDB()
	require.NoError(t, err)

	err = postgres.DB.Migrator().DropTable(&domain.Product{})
	require.NoError(t, err)
	err = postgres.DB.AutoMigrate(&domain.Product{})
	require.NoError(t, err)

	return postgres
}

func TestProductRepository_Create(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       9.99,
		Stock:       100,
		CategoryID:  uuid.New(),
	}

	err := repo.Create(ctx, product)
	assert.NoError(t, err)

	var found domain.Product
	err = postgres.DB.First(&found, "id = ?", product.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, product.Name, found.Name)
	assert.Equal(t, product.Price, found.Price)
	assert.Equal(t, product.Stock, found.Stock)
}

func TestProductRepository_GetByID(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       9.99,
		Stock:       100,
		CategoryID:  uuid.New(),
	}

	err := postgres.DB.Create(product).Error
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, product.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, product.ID, found.ID)
	assert.Equal(t, product.Name, found.Name)

	_, err = repo.GetByID(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrProductNotFound)
}

func TestProductRepository_List(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	products := []domain.Product{
		{ID: uuid.New(), Name: "Product 1", Price: 9.99, Stock: 100, CategoryID: uuid.New()},
		{ID: uuid.New(), Name: "Product 2", Price: 19.99, Stock: 200, CategoryID: uuid.New()},
	}

	for _, p := range products {
		err := postgres.DB.Create(&p).Error
		require.NoError(t, err)
	}

	found, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, len(products))
}

func TestProductRepository_ListByCategoryID(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	categoryID := uuid.New()
	products := []domain.Product{
		{ID: uuid.New(), Name: "Product 1", Price: 9.99, Stock: 100, CategoryID: categoryID},
		{ID: uuid.New(), Name: "Product 2", Price: 19.99, Stock: 200, CategoryID: categoryID},
		{ID: uuid.New(), Name: "Product 3", Price: 29.99, Stock: 300, CategoryID: uuid.New()},
	}

	for _, p := range products {
		err := postgres.DB.Create(&p).Error
		require.NoError(t, err)
	}

	found, err := repo.ListByCategoryID(ctx, categoryID.String())
	assert.NoError(t, err)
	assert.Len(t, found, 2)
	for _, p := range found {
		assert.Equal(t, categoryID, p.CategoryID)
	}
}

func TestProductRepository_Update(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       9.99,
		Stock:       100,
		CategoryID:  uuid.New(),
	}

	err := postgres.DB.Create(product).Error
	require.NoError(t, err)

	product.Name = "Updated Product"
	product.Price = 19.99
	err = repo.Update(ctx, product)
	assert.NoError(t, err)

	var found domain.Product
	err = postgres.DB.First(&found, "id = ?", product.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Product", found.Name)
	assert.Equal(t, 19.99, found.Price)

	nonExistent := &domain.Product{ID: uuid.New()}
	err = repo.Update(ctx, nonExistent)
	assert.ErrorIs(t, err, customErrors.ErrProductNotFound)
}

func TestProductRepository_Delete(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       9.99,
		Stock:       100,
		CategoryID:  uuid.New(),
	}

	err := postgres.DB.Create(product).Error
	require.NoError(t, err)

	err = repo.Delete(ctx, product.ID.String())
	assert.NoError(t, err)

	err = postgres.DB.First(&domain.Product{}, "id = ?", product.ID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = repo.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrProductNotFound)
}

func TestProductRepository_UpdateStock(t *testing.T) {
	postgres := setupProductTestDB(t)
	repo := NewProductRepository(postgres)
	ctx := context.Background()

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       9.99,
		Stock:       100,
		CategoryID:  uuid.New(),
	}

	err := postgres.DB.Create(product).Error
	require.NoError(t, err)

	newStock := 50
	err = repo.UpdateStock(ctx, product.ID.String(), newStock)
	assert.NoError(t, err)

	var found domain.Product
	err = postgres.DB.First(&found, "id = ?", product.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, newStock, found.Stock)

	err = repo.UpdateStock(ctx, uuid.New().String(), newStock)
	assert.ErrorIs(t, err, customErrors.ErrProductNotFound)
}
