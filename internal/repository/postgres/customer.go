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
	CustomerRepository interface {
		Create(ctx context.Context, customer *domain.Customer) error
		GetByID(ctx context.Context, id string) (*domain.Customer, error)
		GetByUserID(ctx context.Context, userID string) (*domain.Customer, error)
		List(ctx context.Context) ([]domain.Customer, error)
		Update(ctx context.Context, customer *domain.Customer) error
		Delete(ctx context.Context, id string) error
	}

	CustomerRepositoryImpl struct {
		*db.BaseRepository[domain.Customer]
	}
)

func NewCustomerRepository(postgres *db.PostgresDB) *CustomerRepositoryImpl {
	return &CustomerRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.Customer](postgres),
	}
}

func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *domain.Customer) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Customer]) error {
		if err := txRepo.GetDB().WithContext(ctx).Create(customer).Error; err != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidCustomerData, err)
		}
		return nil
	})
}

func (r *CustomerRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Preload("User").
		First(&customer, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrCustomerNotFound, err)
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrCustomerNotFound, err)
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) List(ctx context.Context) ([]domain.Customer, error) {
	var customers []domain.Customer
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Preload("User").
		Find(&customers).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return customers, nil
}

func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *domain.Customer) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Customer]) error {
		result := txRepo.GetDB().WithContext(ctx).Save(customer)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidCustomerData, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrCustomerNotFound
		}
		return nil
	})
}

func (r *CustomerRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Customer]) error {
		result := txRepo.GetDB().WithContext(ctx).Delete(&domain.Customer{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrDBQuery, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrCustomerNotFound
		}
		return nil
	})
}
