package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(
	t *testing.T,
	models ...interface{},
) *db.PostgresDB {
	testConfig, err := config.LoadTestConfig()
	require.NoError(t, err)
	postgres, err := db.NewTestDB(testConfig)
	require.NoError(t, err)

	// Drop and recreate tables for all models
	for _, model := range models {
		err = postgres.DB.Migrator().DropTable(model)
		require.NoError(t, err)
		err = postgres.DB.AutoMigrate(model)
		require.NoError(t, err)
	}

	return postgres
}

func createTestUser(
	t *testing.T,
	db *gorm.DB,
) *domain.User {
	uniqueEmail := fmt.Sprintf(
		"test-%s@example.com",
		uuid.New().String(),
	)

	user := &domain.User{
		ID:      uuid.New(),
		Email:   uniqueEmail,
		Name:    "Test User",
		Role:    domain.CustomerRole,
		Picture: "https://example.com/picture.jpg",
	}

	require.NoError(t, db.Create(user).Error)
	return user
}

func createTestCustomer(
	t *testing.T,
	db *gorm.DB,
) *domain.Customer {
	user := createTestUser(t, db)

	customer := &domain.Customer{
		ID:        uuid.New(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, db.Create(customer).Error)
	return customer
}

func createTestCategory(
	t *testing.T,
	db *gorm.DB,
) *domain.Category {
	category := &domain.Category{
		ID:   uuid.New(),
		Name: "Test Category",
	}

	require.NoError(t, db.Create(category).Error)
	return category
}

func createTestProduct(
	t *testing.T,
	db *gorm.DB,
) *domain.Product {
	category := createTestCategory(t, db)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       10.00,
		Stock:       100,
		CategoryID:  category.ID,
	}

	require.NoError(t, db.Create(product).Error)
	return product
}
