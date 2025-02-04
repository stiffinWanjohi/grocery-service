package service

import (
	"context"
	"fmt"

	"github.com/grocery-service/internal/domain"
	repository "github.com/grocery-service/internal/repository/postgres"
	customErrors "github.com/grocery-service/utils/errors"
)

type (
	CustomerService interface {
		Create(ctx context.Context, customer *domain.Customer) error
		GetByID(ctx context.Context, id string) (*domain.Customer, error)
		GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
		List(ctx context.Context) ([]domain.Customer, error)
		Update(ctx context.Context, customer *domain.Customer) error
		Delete(ctx context.Context, id string) error
	}

	CustomerServiceImpl struct {
		repo repository.CustomerRepository
	}
)

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &CustomerServiceImpl{repo: repo}
}

func (s *CustomerServiceImpl) Create(ctx context.Context, customer *domain.Customer) error {
	if err := customer.Validate(); err != nil {
		return fmt.Errorf("%w: %v", customErrors.ErrInvalidCustomerData, err)
	}

	existing, err := s.repo.GetByEmail(ctx, customer.Email)
	if err == nil && existing != nil {
		return fmt.Errorf("%w: email already registered", customErrors.ErrInvalidCustomerData)
	}

	return s.repo.Create(ctx, customer)
}

func (s *CustomerServiceImpl) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: customer ID is required", customErrors.ErrInvalidCustomerData)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *CustomerServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", customErrors.ErrInvalidCustomerData)
	}
	return s.repo.GetByEmail(ctx, email)
}

func (s *CustomerServiceImpl) List(ctx context.Context) ([]domain.Customer, error) {
	return s.repo.List(ctx)
}

func (s *CustomerServiceImpl) Update(ctx context.Context, customer *domain.Customer) error {
	if err := customer.Validate(); err != nil {
		return fmt.Errorf("%w: %v", customErrors.ErrInvalidCustomerData, err)
	}

	existing, err := s.repo.GetByEmail(ctx, customer.Email)
	if err == nil && existing != nil && existing.ID != customer.ID {
		return fmt.Errorf("%w: email already registered", customErrors.ErrInvalidCustomerData)
	}

	return s.repo.Update(ctx, customer)
}

func (s *CustomerServiceImpl) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: customer ID is required", customErrors.ErrInvalidCustomerData)
	}

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find customer: %w", err)
	}

	return s.repo.Delete(ctx, id)
}
