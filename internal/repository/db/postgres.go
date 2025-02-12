package db

import (
	"fmt"
	"os"

	"github.com/grocery-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgresDB(
	config *config.DatabaseConfig,
) (*PostgresDB, error) {
	dsn := config.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return &PostgresDB{DB: db}, nil
}

// NewTestDB creates a new PostgresDB instance for testing
func NewTestDB(
	config *config.TestDatabaseConfig,
) (*PostgresDB, error) {
	testDsn := config.TestDSN()
	db, err := gorm.Open(
		postgres.Open(testDsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get test database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	return &PostgresDB{DB: db}, nil
}

func (db *PostgresDB) Close() error {
	if db == nil || db.DB == nil {
		return fmt.Errorf("invalid database connection")
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
