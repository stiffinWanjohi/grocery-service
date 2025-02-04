package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCustomerTestDB(t *testing.T) *db.PostgresDB {
	postgres, err := db.NewTestDB()
	require.NoError(t, err)

	err = postgres.DB.Migrator().DropTable(&domain.Customer{})
	require.NoError(t, err)
	err = postgres.DB.AutoMigrate(&domain.Customer{})
	require.NoError(t, err)

	return postgres
}

func TestCustomerRepository_Create(t *testing.T) {
	postgres := setupCustomerTestDB(t)
	repo := NewCustomerRepository(postgres)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := repo.Create(ctx, customer)
	assert.NoError(t, err)

	var found domain.Customer
	err = postgres.DB.First(&found, "id = ?", customer.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, customer.Name, found.Name)
	assert.Equal(t, customer.Email, found.Email)
}

func TestCustomerRepository_GetByID(t *testing.T) {
	postgres := setupCustomerTestDB(t)
	repo := NewCustomerRepository(postgres)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := postgres.DB.Create(customer).Error
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, customer.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)
	assert.Equal(t, customer.Name, found.Name)

	_, err = repo.GetByID(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrCustomerNotFound)
}

func TestCustomerRepository_GetByEmail(t *testing.T) {
	postgres := setupCustomerTestDB(t)
	repo := NewCustomerRepository(postgres)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := postgres.DB.Create(customer).Error
	require.NoError(t, err)

	found, err := repo.GetByEmail(ctx, customer.Email)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)
	assert.Equal(t, customer.Email, found.Email)

	_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.ErrorIs(t, err, customErrors.ErrCustomerNotFound)
}

func TestCustomerRepository_List(t *testing.T) {
	postgres := setupCustomerTestDB(t)
	repo := NewCustomerRepository(postgres)
	ctx := context.Background()

	customers := []domain.Customer{
		{ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
		{ID: uuid.New(), Name: "Jane Doe", Email: "jane@example.com"},
	}

	for _, c := range customers {
		err := postgres.DB.Create(&c).Error
		require.NoError(t, err)
	}

	found, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, len(customers))
}

func TestCustomerRepository_Update(t *testing.T) {
	postgres := setupCustomerTestDB(t)
	repo := NewCustomerRepository(postgres)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := postgres.DB.Create(customer).Error
	require.NoError(t, err)

	customer.Name = "John Updated"
	err = repo.Update(ctx, customer)
	assert.NoError(t, err)

	var found domain.Customer
	err = postgres.DB.First(&found, "id = ?", customer.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "John Updated", found.Name)

	nonExistent := &domain.Customer{ID: uuid.New()}
	err = repo.Update(ctx, nonExistent)
	assert.ErrorIs(t, err, customErrors.ErrCustomerNotFound)
}

func TestCustomerRepository_Delete(t *testing.T) {
	postgres := setupCustomerTestDB(t)
	repo := NewCustomerRepository(postgres)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := postgres.DB.Create(customer).Error
	require.NoError(t, err)

	err = repo.Delete(ctx, customer.ID.String())
	assert.NoError(t, err)

	err = postgres.DB.First(&domain.Customer{}, "id = ?", customer.ID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = repo.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrCustomerNotFound)
}
