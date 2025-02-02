package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *db.PostgresDB
}

func NewCustomerRepository(db *db.PostgresDB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	if err := r.db.DB.WithContext(ctx).Create(customer).Error; err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.DB.WithContext(ctx).First(&customer, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return &customer, nil
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.DB.WithContext(ctx).Where("email = ?", email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}
	return &customer, nil
}

func (r *CustomerRepository) List(ctx context.Context) ([]domain.Customer, error) {
	var customers []domain.Customer
	if err := r.db.DB.WithContext(ctx).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	return customers, nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	result := r.db.DB.WithContext(ctx).Save(customer)
	if result.Error != nil {
		return fmt.Errorf("failed to update customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	result := r.db.DB.WithContext(ctx).Delete(&domain.Customer{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}
