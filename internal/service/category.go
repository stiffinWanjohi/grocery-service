package service

import (
	"context"
	"fmt"

	"github.com/grocery-service/internal/domain"
	repository "github.com/grocery-service/internal/repository/postgres"
	customErrors "github.com/grocery-service/utils/errors"
)

type (
	CategoryService interface {
		Create(ctx context.Context, category *domain.Category) error
		GetByID(ctx context.Context, id string) (*domain.Category, error)
		List(ctx context.Context) ([]domain.Category, error)
		ListByParentID(
			ctx context.Context,
			parentID string,
		) ([]domain.Category, error)
		Update(ctx context.Context, category *domain.Category) error
		Delete(ctx context.Context, id string) error
	}

	CategoryServiceImpl struct {
		repo repository.CategoryRepository
	}
)

func NewCategoryService(
	repo repository.CategoryRepository,
) CategoryService {
	return &CategoryServiceImpl{repo: repo}
}

func (s *CategoryServiceImpl) Create(
	ctx context.Context,
	category *domain.Category,
) error {
	if err := category.Validate(); err != nil {
		return fmt.Errorf("%w: %v", customErrors.ErrInvalidCategoryData, err)
	}

	if category.ParentID != nil {
		_, err := s.repo.GetByID(ctx, category.ParentID.String())
		if err != nil {
			return fmt.Errorf("invalid parent category: %w", err)
		}
	}

	return s.repo.Create(ctx, category)
}

func (s *CategoryServiceImpl) GetByID(
	ctx context.Context,
	id string,
) (*domain.Category, error) {
	if id == "" {
		return nil, fmt.Errorf(
			"%w: category ID is required",
			customErrors.ErrInvalidCategoryData,
		)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *CategoryServiceImpl) List(
	ctx context.Context,
) ([]domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryServiceImpl) ListByParentID(
	ctx context.Context,
	parentID string,
) ([]domain.Category, error) {
	if parentID == "" {
		return nil, fmt.Errorf(
			"%w: parent ID is required",
			customErrors.ErrInvalidCategoryData,
		)
	}

	return s.repo.ListByParentID(ctx, parentID)
}

func (s *CategoryServiceImpl) Update(
	ctx context.Context,
	category *domain.Category,
) error {
	if err := category.Validate(); err != nil {
		return fmt.Errorf("%w: %v", customErrors.ErrInvalidCategoryData, err)
	}

	if category.ParentID != nil {
		if category.ID == *category.ParentID {
			return fmt.Errorf(
				"%w: category cannot be its own parent",
				customErrors.ErrInvalidCategoryData,
			)
		}

		_, err := s.repo.GetByID(ctx, category.ParentID.String())
		if err != nil {
			return fmt.Errorf("invalid parent category: %w", err)
		}
	}

	return s.repo.Update(ctx, category)
}

func (s *CategoryServiceImpl) Delete(
	ctx context.Context,
	id string,
) error {
	if id == "" {
		return fmt.Errorf(
			"%w: category ID is required",
			customErrors.ErrInvalidCategoryData,
		)
	}

	subcategories, err := s.repo.ListByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf(
			"failed to check subcategories: %w",
			err,
		)
	}

	if len(subcategories) > 0 {
		return fmt.Errorf(
			"%w: cannot delete category with subcategories",
			customErrors.ErrInvalidCategoryData,
		)
	}

	return s.repo.Delete(ctx, id)
}
