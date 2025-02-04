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
	ProductRepository interface {
		Create(ctx context.Context, product *domain.Product) error
		GetByID(ctx context.Context, id string) (*domain.Product, error)
		List(ctx context.Context) ([]domain.Product, error)
		ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error)
		Update(ctx context.Context, product *domain.Product) error
		Delete(ctx context.Context, id string) error
		UpdateStock(ctx context.Context, id string, quantity int) error
	}

	ProductRepositoryImpl struct {
		*db.BaseRepository[domain.Product]
	}
)

func NewProductRepository(postgres *db.PostgresDB) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{
		BaseRepository: db.NewBaseRepository[domain.Product](postgres),
	}
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, product *domain.Product) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Product]) error {
		if err := txRepo.GetDB().WithContext(ctx).Create(product).Error; err != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidProductData, err)
		}
		return nil
	})
}

func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var product domain.Product
	err := r.BaseRepository.GetDB().WithContext(ctx).First(&product, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErrors.ErrProductNotFound, err)
	}
	return &product, nil
}

func (r *ProductRepositoryImpl) List(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	err := r.BaseRepository.GetDB().WithContext(ctx).Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return products, nil
}

func (r *ProductRepositoryImpl) ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error) {
	var products []domain.Product
	err := r.BaseRepository.GetDB().WithContext(ctx).
		Where("category_id = ?", categoryID).
		Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErrors.ErrDBQuery, err)
	}
	return products, nil
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, product *domain.Product) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Product]) error {
		result := txRepo.GetDB().WithContext(ctx).Save(product)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidProductData, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrProductNotFound
		}
		return nil
	})
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Product]) error {
		result := txRepo.GetDB().WithContext(ctx).Delete(&domain.Product{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrDBQuery, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrProductNotFound
		}
		return nil
	})
}

func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, id string, quantity int) error {
	return r.BaseRepository.WithTransaction(ctx, func(txRepo *db.BaseRepository[domain.Product]) error {
		result := txRepo.GetDB().WithContext(ctx).
			Model(&domain.Product{}).
			Where("id = ?", id).
			Update("stock", gorm.Expr("stock + ?", quantity))

		if result.Error != nil {
			return fmt.Errorf("%w: %v", customErrors.ErrInvalidProductData, result.Error)
		}
		if result.RowsAffected == 0 {
			return customErrors.ErrProductNotFound
		}
		return nil
	})
}
