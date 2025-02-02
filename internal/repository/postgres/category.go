package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *db.PostgresDB
}

func NewCategoryRepository(db *db.PostgresDB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	if err := r.db.DB.WithContext(ctx).Create(category).Error; err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.DB.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &category, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.DB.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

func (r *CategoryRepository) ListByParentID(ctx context.Context, parentID string) ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.DB.WithContext(ctx).Where("parent_id = ?", parentID).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to list categories by parent ID: %w", err)
	}
	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	result := r.db.DB.WithContext(ctx).Save(category)
	if result.Error != nil {
		return fmt.Errorf("failed to update category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	result := r.db.DB.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
