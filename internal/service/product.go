package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	repository "github.com/grocery-service/internal/repository/postgres"
	customErrors "github.com/grocery-service/utils/errors"
)

type (
	ProductService interface {
		Create(ctx context.Context, product *domain.Product) error
		GetByID(ctx context.Context, id string) (*domain.Product, error)
		List(ctx context.Context) ([]domain.Product, error)
		ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error)
		Update(ctx context.Context, product *domain.Product) error
		Delete(ctx context.Context, id string) error
		UpdateStock(ctx context.Context, id string, quantity int) error
	}

	ProductServiceImpl struct {
		repo         repository.ProductRepository
		categoryRepo repository.CategoryRepository
	}
)

func NewProductService(
	repo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) ProductService {
	return &ProductServiceImpl{
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

func (s *ProductServiceImpl) Create(ctx context.Context, product *domain.Product) error {
	if err := product.Validate(); err != nil {
		return fmt.Errorf("%w: %v", customErrors.ErrInvalidProductData, err)
	}

	if product.CategoryID != uuid.Nil {
		if _, err := s.categoryRepo.GetByID(ctx, product.CategoryID.String()); err != nil {
			return fmt.Errorf("invalid category: %w", err)
		}
	}

	return s.repo.Create(ctx, product)
}

func (s *ProductServiceImpl) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: product ID is required", customErrors.ErrInvalidProductData)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ProductServiceImpl) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductServiceImpl) ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error) {
	if categoryID == "" {
		return nil, fmt.Errorf("%w: category ID is required", customErrors.ErrInvalidProductData)
	}
	return s.repo.ListByCategoryID(ctx, categoryID)
}

func (s *ProductServiceImpl) Update(ctx context.Context, product *domain.Product) error {
	if err := product.Validate(); err != nil {
		return fmt.Errorf("%w: %v", customErrors.ErrInvalidProductData, err)
	}

	existingProduct, err := s.repo.GetByID(ctx, product.ID.String())
	if err != nil {
		return fmt.Errorf("failed to get existing product: %w", err)
	}

	if product.CategoryID != existingProduct.CategoryID && product.CategoryID != uuid.Nil {
		if _, err := s.categoryRepo.GetByID(ctx, product.CategoryID.String()); err != nil {
			return fmt.Errorf("invalid category: %w", err)
		}
	}

	return s.repo.Update(ctx, product)
}

func (s *ProductServiceImpl) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: product ID is required", customErrors.ErrInvalidProductData)
	}
	return s.repo.Delete(ctx, id)
}

func (s *ProductServiceImpl) UpdateStock(ctx context.Context, id string, quantity int) error {
	if id == "" {
		return fmt.Errorf("%w: product ID is required", customErrors.ErrInvalidProductData)
	}

	if quantity < 0 {
		return fmt.Errorf("%w: stock quantity cannot be negative", customErrors.ErrInvalidProductData)
	}

	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	return s.repo.UpdateStock(ctx, id, quantity)
}
