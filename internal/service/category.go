package service

import (
	"context"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, category *domain.Category) error {
	if err := s.validateCategory(category); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}
	return s.repo.Create(ctx, category)
}

func (s *CategoryService) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) ListByParentID(ctx context.Context, parentID string) ([]domain.Category, error) {
	return s.repo.ListByParentID(ctx, parentID)
}

func (s *CategoryService) Update(ctx context.Context, category *domain.Category) error {
	if err := s.validateCategory(category); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}
	return s.repo.Update(ctx, category)
}

func (s *CategoryService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) validateCategory(category *domain.Category) error {
	if category.Name == "" {
		return fmt.Errorf("category name is required")
	}
	return nil
}
