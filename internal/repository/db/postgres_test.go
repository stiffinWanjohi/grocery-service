package db

import (
	"testing"

	"github.com/grocery-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresDB(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name: "successful connection",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Name:     "grocery_test",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "failed connection - wrong port",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     1234,
				User:     "postgres",
				Password: "postgres",
				Name:     "grocery_test",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewPostgresDB(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, db)

			// Test connection
			sqlDB, err := db.DB.DB()
			require.NoError(t, err)
			assert.NoError(t, sqlDB.Ping())

			// Clean up
			assert.NoError(t, db.Close())
		})
	}
}

func TestNewTestDB(t *testing.T) {
	// Test creating test database
	db, err := NewTestDB()
	require.NoError(t, err)
	assert.NotNil(t, db)

	// Verify connection works
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())

	// Test closing connection
	assert.NoError(t, db.Close())
}

func TestPostgresDB_Close(t *testing.T) {
	// Create test database
	db, err := NewTestDB()
	require.NoError(t, err)

	// Test closing
	err = db.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	err = sqlDB.Ping()
	assert.Error(t, err)
}
