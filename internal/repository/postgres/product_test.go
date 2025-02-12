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

func TestProductRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		setupProduct  func(categoryID uuid.UUID) *domain.Product
		expectedError error
	}{
		{
			name: "Success - Create Product",
			setupProduct: func(categoryID uuid.UUID) *domain.Product {
				return &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
			},
			expectedError: nil,
		},
		{
			name: "Error - Empty Name",
			setupProduct: func(categoryID uuid.UUID) *domain.Product {
				return &domain.Product{
					ID:          uuid.New(),
					Description: "Test Description",
					Price:       9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
			},
			expectedError: customErrors.ErrInvalidProductData,
		},
		{
			name: "Error - Invalid Price",
			setupProduct: func(categoryID uuid.UUID) *domain.Product {
				return &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       -9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
			},
			expectedError: customErrors.ErrInvalidProductData,
		},
		{
			name: "Error - Invalid Stock",
			setupProduct: func(categoryID uuid.UUID) *domain.Product {
				return &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       9.99,
					Stock:       -100,
					CategoryID:  categoryID,
				}
			},
			expectedError: customErrors.ErrInvalidProductData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			categoryID := createTestCategory(t, postgres.DB).ID
			product := tt.setupProduct(categoryID)
			err := repo.Create(ctx, product)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Product
				err = postgres.DB.First(&found, "id = ?", product.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, product.ID, found.ID)
				assert.Equal(t, product.Name, found.Name)
				assert.Equal(t, product.Description, found.Description)
				assert.Equal(t, product.Price, found.Price)
				assert.Equal(t, product.Stock, found.Stock)
				assert.Equal(t, product.CategoryID, found.CategoryID)
			}
		})
	}
}

func TestProductRepository_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, *domain.Product)
		expectedError error
	}{
		{
			name: "Success - Get Existing Product",
			setupTest: func(db *gorm.DB) (string, *domain.Product) {
				categoryID := createTestCategory(t, db).ID
				product := &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
				require.NoError(t, db.Create(product).Error)
				return product.ID.String(), product
			},
			expectedError: nil,
		},
		{
			name: "Error - Product Not Found",
			setupTest: func(_ *gorm.DB) (string, *domain.Product) {
				return uuid.New().String(), nil
			},
			expectedError: customErrors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			id, expected := tt.setupTest(postgres.DB)
			found, err := repo.GetByID(ctx, id)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, found)
				result := postgres.DB.Find(&domain.Product{}, "id = ?", id)
				assert.NoError(t, result.Error)
				assert.Equal(t, int64(0), result.RowsAffected)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, expected.ID, found.ID)
				assert.Equal(t, expected.Name, found.Name)
				assert.Equal(t, expected.Description, found.Description)
				assert.Equal(t, expected.Price, found.Price)
				assert.Equal(t, expected.Stock, found.Stock)
				assert.Equal(t, expected.CategoryID, found.CategoryID)
			}
		})
	}
}

func TestProductRepository_List(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) []domain.Product
		expectedCount int
	}{
		{
			name: "Success - List Multiple Products",
			setupTest: func(db *gorm.DB) []domain.Product {
				categoryID := createTestCategory(t, db).ID
				products := []domain.Product{
					{
						ID:         uuid.New(),
						Name:       "Product 1",
						Price:      9.99,
						Stock:      100,
						CategoryID: categoryID,
					},
					{
						ID:         uuid.New(),
						Name:       "Product 2",
						Price:      19.99,
						Stock:      200,
						CategoryID: categoryID,
					},
				}

				for _, p := range products {
					require.NoError(t, db.Create(&p).Error)
				}

				return products
			},
			expectedCount: 2,
		},
		{
			name: "Success - Empty List",
			setupTest: func(_ *gorm.DB) []domain.Product {
				return []domain.Product{}
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			expected := tt.setupTest(postgres.DB)
			found, err := repo.List(ctx)

			assert.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
			if tt.expectedCount > 0 {
				for i, p := range found {
					assert.Equal(t, expected[i].ID, p.ID)
					assert.Equal(t, expected[i].Name, p.Name)
					assert.Equal(t, expected[i].CategoryID, p.CategoryID)
				}
			}
		})
	}
}

func TestProductRepository_ListByCategoryID(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, []domain.Product)
		expectedCount int
	}{
		{
			name: "Success - List Products By Category",
			setupTest: func(db *gorm.DB) (string, []domain.Product) {
				categoryID := createTestCategory(t, db).ID
				products := []domain.Product{
					{
						ID:         uuid.New(),
						Name:       "Product 1",
						Price:      9.99,
						Stock:      100,
						CategoryID: categoryID,
					},
					{
						ID:         uuid.New(),
						Name:       "Product 2",
						Price:      19.99,
						Stock:      200,
						CategoryID: categoryID,
					},
					{
						ID:         uuid.New(),
						Name:       "Product 3",
						Price:      29.99,
						Stock:      300,
						CategoryID: createTestCategory(t, db).ID,
					},
				}

				for _, p := range products {
					require.NoError(t, db.Create(&p).Error)
				}

				return categoryID.String(), products[:2]
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			categoryID, expectedProducts := tt.setupTest(postgres.DB)
			found, err := repo.ListByCategoryID(ctx, categoryID)

			assert.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
			for i, p := range found {
				assert.Equal(t, expectedProducts[i].ID, p.ID)
				assert.Equal(t, categoryID, p.CategoryID.String())
			}
		})
	}
}

func TestProductRepository_Update(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) *domain.Product
		updateFunc    func(*domain.Product)
		expectedError error
	}{
		{
			name: "Success - Update Product",
			setupTest: func(db *gorm.DB) *domain.Product {
				categoryID := createTestCategory(t, db).ID
				product := &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
				require.NoError(t, db.Create(product).Error)
				return product
			},
			updateFunc: func(p *domain.Product) {
				p.Name = "Updated Product"
				p.Price = 19.99
			},
			expectedError: nil,
		},
		{
			name: "Error - Product Not Found",
			setupTest: func(db *gorm.DB) *domain.Product {
				categoryID := createTestCategory(t, db).ID
				return &domain.Product{
					ID:         uuid.New(),
					Name:       "Non-existent Product",
					CategoryID: categoryID,
					Price:      9.99,
					Stock:      100,
				}
			},
			updateFunc:    func(_ *domain.Product) {},
			expectedError: customErrors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			product := tt.setupTest(postgres.DB)
			tt.updateFunc(product)
			err := repo.Update(ctx, product)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Product
				err = postgres.DB.First(&found, "id = ?", product.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, product.Name, found.Name)
				assert.Equal(t, product.Price, found.Price)
				assert.Equal(t, product.CategoryID, found.CategoryID)
			}
		})
	}
}

func TestProductRepository_Delete(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) string
		expectedError error
	}{
		{
			name: "Success - Delete Product",
			setupTest: func(db *gorm.DB) string {
				categoryID := createTestCategory(t, db).ID
				product := &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
				require.NoError(t, db.Create(product).Error)
				return product.ID.String()
			},
			expectedError: nil,
		},
		{
			name: "Error - Product Not Found",
			setupTest: func(_ *gorm.DB) string {
				return uuid.New().String()
			},
			expectedError: customErrors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			id := tt.setupTest(postgres.DB)
			err := repo.Delete(ctx, id)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				result := postgres.DB.Find(&domain.Product{}, "id = ?", id)
				assert.NoError(t, result.Error)
				assert.Equal(t, int64(0), result.RowsAffected)
			}
		})
	}
}

func TestProductRepository_UpdateStock(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, int)
		expectedError error
	}{
		{
			name: "Success - Update Stock",
			setupTest: func(db *gorm.DB) (string, int) {
				categoryID := createTestCategory(t, db).ID
				product := &domain.Product{
					ID:          uuid.New(),
					Name:        "Test Product",
					Description: "Test Description",
					Price:       9.99,
					Stock:       100,
					CategoryID:  categoryID,
				}
				require.NoError(t, db.Create(product).Error)
				return product.ID.String(), 50
			},
			expectedError: nil,
		},
		{
			name: "Error - Product Not Found",
			setupTest: func(_ *gorm.DB) (string, int) {
				return uuid.New().String(), 50
			},
			expectedError: customErrors.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Product{}, &domain.Category{})
			repo := NewProductRepository(postgres)
			ctx := context.Background()

			id, quantity := tt.setupTest(postgres.DB)
			err := repo.UpdateStock(ctx, id, quantity)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Product
				err = postgres.DB.First(&found, "id = ?", id).Error
				assert.NoError(t, err)
				assert.Equal(t, 150, found.Stock)
			}
		})
	}
}
