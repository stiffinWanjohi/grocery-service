package service

import (
	"context"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, product *domain.Product) error {
	if err := s.validateProduct(product); err != nil {
		return fmt.Errorf("invalid product: %w", err)
	}
	return s.repo.Create(ctx, product)
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductService) ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error) {
	return s.repo.ListByCategoryID(ctx, categoryID)
}

func (s *ProductService) Update(ctx context.Context, product *domain.Product) error {
	if err := s.validateProduct(product); err != nil {
		return fmt.Errorf("invalid product: %w", err)
	}
	return s.repo.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) UpdateStock(ctx context.Context, id string, quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	return s.repo.UpdateStock(ctx, id, quantity)
}

func (s *ProductService) validateProduct(product *domain.Product) error {
	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if product.Price <= 0 {
		return fmt.Errorf("product price must be greater than zero")
	}
	if product.Stock < 0 {
		return fmt.Errorf("product stock cannot be negative")
	}
	return nil
}
