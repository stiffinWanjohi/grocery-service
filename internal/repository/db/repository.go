package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository is a generic interface for database operations with transaction support
type Repository[T any] interface {
	// WithTx creates a new repository instance using the provided transaction
	WithTx(tx *gorm.DB) Repository[T]
}

// BaseRepository provides a base implementation for repository operations
type BaseRepository[T any] struct {
	db *PostgresDB
}

// NewBaseRepository creates a new BaseRepository
func NewBaseRepository[T any](db *PostgresDB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// GetDB returns the database instance
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db.DB
}

// WithTx creates a new repository instance using the provided transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: &PostgresDB{DB: tx}}
}

// BeginTransaction starts a new database transaction
func (r *BaseRepository[T]) BeginTransaction() (*gorm.DB, error) {
	tx := r.db.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	return tx, nil
}

// WithTransaction is a utility method to run a function within a transaction
func (r *BaseRepository[T]) WithTransaction(
	ctx context.Context,
	fn func(txRepo *BaseRepository[T]) error,
) error {
	tx, err := r.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	txRepo := r.WithTx(tx)

	if err := fn(txRepo); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
