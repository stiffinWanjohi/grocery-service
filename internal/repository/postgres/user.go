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
	UserRepository interface {
		Create(ctx context.Context, user *domain.User) error
		GetByID(ctx context.Context, id string) (*domain.User, error)
		GetByEmail(ctx context.Context, email string) (*domain.User, error)
		GetByProviderID(ctx context.Context, providerID string) (*domain.User, error)
		Update(ctx context.Context, user *domain.User) error
		Delete(ctx context.Context, id string) error
	}

	UserRepositoryImpl struct {
		*db.BaseRepository[domain.User]
	}
)

func NewUserRepository(postgres *db.PostgresDB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.User](postgres),
	}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.User]) error {
		if err := txRepo.GetDB().WithContext(ctx).Create(user).Error; err != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidUserData, err)
		}
		return nil
	})
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.BaseRepository.GetDB().WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.BaseRepository.GetDB().WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByProviderID(ctx context.Context, providerID string) (*domain.User, error) {
	var user domain.User
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Joins("JOIN tokens ON tokens.user_id = users.id").
		Where("tokens.provider_id = ?", providerID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.User]) error {
		result := txRepo.GetDB().WithContext(ctx).Save(user)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidUserData, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrUserNotFound
		}
		return nil
	})
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.User]) error {
		result := txRepo.GetDB().WithContext(ctx).Delete(&domain.User{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrDBQuery, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrUserNotFound
		}
		return nil
	})
}
