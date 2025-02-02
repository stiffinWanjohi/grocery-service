package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *db.PostgresDB
}

func NewProductRepository(db *db.PostgresDB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	if err := r.db.DB.WithContext(ctx).Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.DB.WithContext(ctx).First(&product, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.DB.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.DB.WithContext(ctx).Where("category_id = ?", categoryID).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to list products by category ID: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	result := r.db.DB.WithContext(ctx).Save(product)
	if result.Error != nil {
		return fmt.Errorf("failed to update product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	result := r.db.DB.WithContext(ctx).Delete(&domain.Product{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id string, quantity int) error {
	result := r.db.DB.WithContext(ctx).
		Model(&domain.Product{}).
		Where("id = ?", id).
		Update("stock", quantity)
	if result.Error != nil {
		return fmt.Errorf("failed to update product stock: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}
