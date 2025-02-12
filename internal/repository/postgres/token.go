package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"gorm.io/gorm"
)

type (
	TokenRepository interface {
		Create(ctx context.Context, token *domain.Token) error
		GetByToken(ctx context.Context, token string) (*domain.Token, error)
		GetByUserAndType(
			ctx context.Context,
			userID string,
			tokenType domain.TokenType,
		) (*domain.Token, error)
		GetByProviderID(
			ctx context.Context,
			providerID string,
		) (*domain.Token, error)
		RevokeToken(ctx context.Context, token string) error
		DeleteExpiredTokens(ctx context.Context) error
		IsValid(ctx context.Context, token string) bool
	}

	TokenRepositoryImpl struct {
		*db.BaseRepository[domain.Token]
	}
)

func NewTokenRepository(
	postgres *db.PostgresDB,
) *TokenRepositoryImpl {
	return &TokenRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.Token](
			postgres,
		),
	}
}

func (r *TokenRepositoryImpl) Create(
	ctx context.Context,
	token *domain.Token,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Token]) error {
			if err := txRepo.GetDB().WithContext(ctx).Create(token).Error; err != nil {
				return fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
			}

			return nil
		},
	)
}

func (r *TokenRepositoryImpl) GetByToken(
	ctx context.Context,
	token string,
) (*domain.Token, error) {
	var t domain.Token
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Where("token = ? AND revoked_at IS NULL", token).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrTokenNotFound
		}

		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}

	return &t, nil
}

func (r *TokenRepositoryImpl) GetByUserAndType(
	ctx context.Context,
	userID string,
	tokenType domain.TokenType,
) (*domain.Token, error) {
	var t domain.Token
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Where("user_id = ? AND type = ? AND expires_at > ? AND revoked_at IS NULL",
			userID, tokenType, time.Now()).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrTokenNotFound
		}

		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}

	return &t, nil
}

func (r *TokenRepositoryImpl) GetByProviderID(
	ctx context.Context,
	providerID string,
) (*domain.Token, error) {
	var t domain.Token
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Where("provider_id = ? AND revoked_at IS NULL", providerID).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrTokenNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}

	return &t, nil
}

func (r *TokenRepositoryImpl) RevokeToken(
	ctx context.Context,
	token string,
) error {
	now := time.Now()
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Token]) error {
			result := txRepo.GetDB().WithContext(ctx).
				Model(&domain.Token{}).
				Where("token = ? AND revoked_at IS NULL", token).
				Update("revoked_at", now)

			if result.Error != nil {
				return fmt.Errorf(
					"%w: %v",
					customErrors.ErrDBQuery,
					result.Error,
				)
			}

			if result.RowsAffected == 0 {
				return customErrors.ErrTokenNotFound
			}

			return nil
		},
	)
}

func (r *TokenRepositoryImpl) DeleteExpiredTokens(
	ctx context.Context,
) error {
	return r.BaseRepository.WithTransaction(
		ctx,
		func(txRepo *db.BaseRepository[domain.Token]) error {
			if err := txRepo.GetDB().WithContext(ctx).
				Where("expires_at < ? OR revoked_at IS NOT NULL",
					time.Now()).
				Delete(&domain.Token{}).Error; err != nil {
				return fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
			}

			return nil
		},
	)
}

func (r *TokenRepositoryImpl) IsValid(
	ctx context.Context,
	token string,
) bool {
	var t domain.Token
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Where("token = ? AND expires_at > ? AND revoked_at IS NULL",
			token, time.Now()).
		First(&t).Error

	return err == nil
}
