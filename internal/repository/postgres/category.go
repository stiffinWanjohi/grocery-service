package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"gorm.io/gorm"
)

type (
	CategoryRepository interface {
		Create(ctx context.Context, category *domain.Category) error
		GetByID(ctx context.Context, id string) (*domain.Category, error)
		List(ctx context.Context) ([]domain.Category, error)
		ListByParentID(ctx context.Context, parentID string) ([]domain.Category, error)
		Update(ctx context.Context, category *domain.Category) error
		Delete(ctx context.Context, id string) error
	}

	CategoryRepositoryImpl struct {
		*db.BaseRepository[domain.Category]
	}
)

func NewCategoryRepository(postgres *db.PostgresDB) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.Category](postgres),
	}
}

func (r *CategoryRepositoryImpl) Create(ctx context.Context, category *domain.Category) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Category]) error {
		if err := txRepo.GetDB().WithContext(ctx).Create(category).Error; err != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidCategoryData, err)
		}
		return nil
	})
}

func (r *CategoryRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	var category domain.Category
	err := r.BaseRepository.GetDB().WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrCategoryNotFound, err)
	}
	return &category, nil
}

func (r *CategoryRepositoryImpl) List(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.BaseRepository.GetDB().WithContext(ctx).Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) ListByParentID(ctx context.Context, parentID string) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Where("parent_id = ?", parentID).
		Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) Update(ctx context.Context, category *domain.Category) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Category]) error {
		result := txRepo.GetDB().WithContext(ctx).Save(category)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidCategoryData, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrCategoryNotFound
		}
		return nil
	})
}

func (r *CategoryRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Category]) error {
		result := txRepo.GetDB().WithContext(ctx).Delete(&domain.Category{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrDBQuery, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrCategoryNotFound
		}
		return nil
	})
}