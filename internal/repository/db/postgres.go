package db

import (
	"fmt"

	"github.com/grocery-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgresDB(config *config.DatabaseConfig) (*PostgresDB, error) {
	dsn := config.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return &PostgresDB{DB: db}, nil
}

// NewTestDB creates a new PostgresDB instance for testing
func NewTestDB() (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		"localhost",
		"postgres",
		"postgres",
		"grocery_test",
		"5432",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get test database instance: %w", err)
	}

	// Set connection pool settings for test DB
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	return &PostgresDB{DB: db}, nil
}

func (db *PostgresDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}
