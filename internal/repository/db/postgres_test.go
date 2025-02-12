package db

import (
	"fmt"
	"testing"

	"github.com/grocery-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewPostgresDB(t *testing.T) {
	mockConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:     getEnvOrDefault("TEST_DB_NAME", "grocery_test"),
		SSLMode:  "disable",
	}

	if mockConfig.Password == "postgres" {
		t.Skip("Skipping test: no database credentials provided")
	}

	tests := []struct {
		name          string
		config        *config.DatabaseConfig
		expectedError bool
	}{
		{
			name:          "Success - Valid Database Config",
			config:        mockConfig,
			expectedError: false,
		},
		{
			name: "Error - Invalid Database Config",
			config: &config.DatabaseConfig{
				Host:     "invalid-host",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Name:     "grocery",
				SSLMode:  "disable",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewPostgresDB(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, db)

				// Test connection pool settings
				sqlDB, err := db.DB.DB()
				require.NoError(t, err)
				maxIdle := sqlDB.Stats().Idle
				maxOpen := sqlDB.Stats().OpenConnections
				assert.LessOrEqual(t, maxIdle, 10)
				assert.LessOrEqual(t, maxOpen, 100)

				// Clean up
				err = sqlDB.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewTestDB(t *testing.T) {
	testCfg, err := config.LoadTestConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	tests := []struct {
		name          string
		config        *config.TestDatabaseConfig
		expectedError bool
	}{
		{
			name:          "Success - Valid Test Database Config",
			config:        testCfg,
			expectedError: false,
		},
		{
			name: "Error - Invalid Database Connection",
			config: &config.TestDatabaseConfig{
				Host:     "invalid-host",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Name:     "grocery_test",
				SSLMode:  "disable",
			},
			expectedError: true,
		},
		{
			name: "Error - Invalid Port",
			config: &config.TestDatabaseConfig{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Password: "postgres",
				Name:     "grocery_test",
				SSLMode:  "disable",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewTestDB(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)

				// Test connection pool settings
				sqlDB, err := db.DB.DB()
				require.NoError(t, err)
				maxIdle := sqlDB.Stats().Idle
				maxOpen := sqlDB.Stats().OpenConnections
				assert.LessOrEqual(t, maxIdle, 10)
				assert.LessOrEqual(t, maxOpen, 100)

				// Clean up
				err = sqlDB.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostgresDB_Close(t *testing.T) {
	testCfg, err := config.LoadTestConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	tests := []struct {
		name    string
		db      *PostgresDB
		wantErr bool
	}{
		{
			name: "Success - Close DB Connection",
			db: func() *PostgresDB {
				dsn := fmt.Sprintf(
					"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
					testCfg.Host,
					testCfg.Port,
					testCfg.User,
					testCfg.Password,
					testCfg.Name,
					testCfg.SSLMode,
				)

				db, err := gorm.Open(
					postgres.Open(dsn),
					&gorm.Config{},
				)
				require.NoError(t, err)
				return &PostgresDB{DB: db}
			}(),
			wantErr: false,
		},

		{
			name:    "Error - Invalid DB Connection",
			db:      &PostgresDB{DB: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.db.Close()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(
					t,
					err.Error(),
					"invalid database connection",
				)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
