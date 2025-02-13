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
		ListByParentID(
			ctx context.Context,
			parentID string,
		) ([]domain.Category, error)
		ListRootCategories(ctx context.Context) ([]domain.Category, error)
		IsLeafCategory(ctx context.Context, id string) (bool, error)
		Update(ctx context.Context, category *domain.Category) error
		Delete(ctx context.Context, id string) error
	}

	CategoryRepositoryImpl struct {
		*db.BaseRepository[domain.Category]
	}
)

func NewCategoryRepository(
	postgres *db.PostgresDB,
) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.Category](
			postgres,
		),
	}
}

func (r *CategoryRepositoryImpl) Create(
	ctx context.Context,
	category *domain.Category,
) error {
	if err := category.Validate(); err != nil {
		return customErrors.ErrInvalidCategoryData
	}

	if category.ParentID == nil && category.Level != 0 {
		return fmt.Errorf(
			"%w: invalid level for root category",
			customErrors.ErrInvalidCategoryData,
		)
	}

	return r.BaseRepository.
		WithTransaction(
			ctx,
			func(txRepo *db.BaseRepository[domain.Category]) error {
				if category.Path != "" &&
					!category.ValidatePathFormat(category.Path) {
					return fmt.Errorf(
						"%w: invalid path format",
						customErrors.ErrInvalidCategoryData,
					)
				}

				if category.ParentID != nil {
					var parent domain.Category
					if err := txRepo.GetDB().
						WithContext(ctx).
						First(&parent, "id = ?", category.ParentID).Error; err != nil {
						return fmt.Errorf(
							"%w: parent category not found",
							customErrors.ErrInvalidCategoryData,
						)
					}
					category.Level = parent.Level + 1
					category.Path = fmt.Sprintf(
						"%s/%s",
						parent.Path,
						category.Name,
					)
				} else {
					category.Level = 0
					category.Path = category.Name
				}

				if !category.ValidatePathFormat(category.Path) {
					return fmt.Errorf(
						"%w: invalid path format",
						customErrors.ErrInvalidCategoryData,
					)
				}

				if err := txRepo.GetDB().
					WithContext(ctx).
					Create(category).Error; err != nil {
					return fmt.Errorf(
						"%w: %v",
						customErrors.ErrInvalidCategoryData,
						err,
					)
				}
				return nil
			})
}

func (r *CategoryRepositoryImpl) GetByID(
	ctx context.Context,
	id string,
) (*domain.Category, error) {
	if err := domain.ValidateID(id); err != nil {
		return nil, customErrors.ErrInvalidCategoryData
	}

	var category domain.Category
	err := r.BaseRepository.GetDB().
		WithContext(ctx).
		First(&category, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrCategoryNotFound
		}
		return nil, fmt.Errorf(
			"%w: %v",
			customErrors.ErrCategoryNotFound,
			err,
		)
	}
	return &category, nil
}

func (r *CategoryRepositoryImpl) List(
	ctx context.Context,
) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.BaseRepository.GetDB().
		WithContext(ctx).
		Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			customErrors.ErrDBQuery,
			err,
		)
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) ListByParentID(
	ctx context.Context,
	parentID string,
) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.BaseRepository.GetDB().
		WithContext(ctx).
		Where("parent_id = ?", parentID).
		Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			customErrors.ErrDBQuery,
			err,
		)
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) ListRootCategories(
	ctx context.Context,
) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.BaseRepository.GetDB().
		WithContext(ctx).
		Where("parent_id IS NULL").
		Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) IsLeafCategory(
	ctx context.Context,
	id string,
) (bool, error) {
	var count int64
	err := r.BaseRepository.GetDB().
		WithContext(ctx).
		Model(&domain.Category{}).
		Where("parent_id = ?", id).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return count == 0, nil
}

func (r *CategoryRepositoryImpl) Update(
	ctx context.Context,
	category *domain.Category,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Category]) error {
			var exists domain.Category
			if err := txRepo.GetDB().
				WithContext(ctx).
				First(&exists, "id = ?", category.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return customErrors.ErrCategoryNotFound
				}

				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrDBQuery,
					err,
				)
			}

			if err := txRepo.GetDB().
				WithContext(ctx).
				Save(category).Error; err != nil {
				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrInvalidCategoryData,
					err,
				)
			}
			return nil
		},
	)
}

func (r *CategoryRepositoryImpl) Delete(
	ctx context.Context,
	id string,
) error {
	if err := domain.ValidateID(id); err != nil {
		return customErrors.ErrInvalidCategoryData
	}

	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Category]) error {
			result := txRepo.GetDB().
				WithContext(ctx).
				Delete(&domain.Category{}, "id = ?", id)
			if result.Error != nil {
				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrDBQuery,
					result.Error,
				)
			}

			if result.RowsAffected == 0 {
				return customErrors.ErrCategoryNotFound
			}
			return nil
		},
	)
}
