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
		Create(ctx context.Context, userID string) (*domain.Customer, error)
		GetByID(ctx context.Context, id string) (*domain.Customer, error)
		GetByUserID(
			ctx context.Context,
			userID string,
		) (*domain.Customer, error)
		List(ctx context.Context) ([]domain.Customer, error)
		Delete(ctx context.Context, id string) error
	}

	CustomerServiceImpl struct {
		customerRepo repository.CustomerRepository
		userRepo     repository.UserRepository
	}
)

func NewCustomerService(
	customerRepo repository.CustomerRepository,
	userRepo repository.UserRepository,
) CustomerService {
	return &CustomerServiceImpl{
		customerRepo: customerRepo,
		userRepo:     userRepo,
	}
}

func (s *CustomerServiceImpl) Create(
	ctx context.Context,
	userID string,
) (*domain.Customer, error) {
	if userID == "" {
		return nil, fmt.Errorf(
			"%w: user ID is required",
			customErrors.ErrInvalidCustomerData,
		)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find user: %w",
			err,
		)
	}

	existing, err := s.customerRepo.GetByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf(
			"%w: customer already exists for this user",
			customErrors.ErrInvalidCustomerData,
		)
	}

	customer := &domain.Customer{
		UserID: user.ID,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *CustomerServiceImpl) GetByID(
	ctx context.Context,
	id string,
) (*domain.Customer, error) {
	if id == "" {
		return nil, fmt.Errorf(
			"%w: customer ID is required",
			customErrors.ErrInvalidCustomerData,
		)
	}
	return s.customerRepo.GetByID(ctx, id)
}

func (s *CustomerServiceImpl) GetByUserID(
	ctx context.Context,
	userID string,
) (*domain.Customer, error) {
	if userID == "" {
		return nil, fmt.Errorf(
			"%w: user ID is required",
			customErrors.ErrInvalidCustomerData,
		)
	}
	return s.customerRepo.GetByUserID(ctx, userID)
}

func (s *CustomerServiceImpl) List(
	ctx context.Context,
) ([]domain.Customer, error) {
	return s.customerRepo.List(ctx)
}

func (s *CustomerServiceImpl) Delete(
	ctx context.Context,
	id string,
) error {
	if id == "" {
		return fmt.Errorf(
			"%w: customer ID is required",
			customErrors.ErrInvalidCustomerData,
		)
	}

	_, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf(
			"failed to find customer: %w",
			err,
		)
	}

	return s.customerRepo.Delete(ctx, id)
}
