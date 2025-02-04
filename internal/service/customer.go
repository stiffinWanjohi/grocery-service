package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository"
)

type CustomerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) Create(ctx context.Context, customer *domain.Customer) error {
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}
	return s.repo.Create(ctx, customer)
}

func (s *CustomerService) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CustomerService) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *CustomerService) List(ctx context.Context) ([]domain.Customer, error) {
	return s.repo.List(ctx)
}

func (s *CustomerService) Update(ctx context.Context, customer *domain.Customer) error {
	if err := s.validateCustomer(customer); err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}
	return s.repo.Update(ctx, customer)
}

func (s *CustomerService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *CustomerService) validateCustomer(customer *domain.Customer) error {
	if customer.Name == "" {
		return fmt.Errorf("customer name is required")
	}
	if !isValidEmail(customer.Email) {
		return fmt.Errorf("invalid email format")
	}
	if !isValidPhone(customer.Phone) {
		return fmt.Errorf("invalid phone format")
	}
	return nil
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func isValidPhone(phone string) bool {
	pattern := `^\+?[1-9]\d{1,14}$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}
