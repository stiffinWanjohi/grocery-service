package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryImpl(t *testing.T) {
	postgres := setupTestDB(t, &domain.User{}, &domain.Token{})
	repo := NewUserRepository(postgres)
	ctx := context.Background()

	t.Run("Create and GetByID", func(t *testing.T) {
		user := &domain.User{
			ID:      uuid.New(),
			Email:   "test-" + uuid.NewString() + "@example.com",
			Name:    "Test User",
			Role:    domain.CustomerRole,
			Picture: "https://example.com/picture.jpg",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
		assert.Equal(t, user.Name, retrieved.Name)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		user := &domain.User{
			ID:      uuid.New(),
			Email:   "test-" + uuid.NewString() + "@example.com",
			Name:    "Test User",
			Role:    domain.CustomerRole,
			Picture: "https://example.com/picture.jpg",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		retrieved, err := repo.GetByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)

		// Test non-existent email
		_, err = repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.ErrorIs(t, err, customErrors.ErrUserNotFound)
	})

	t.Run("GetByProviderID", func(t *testing.T) {
		user := &domain.User{
			ID:      uuid.New(),
			Email:   "test-" + uuid.NewString() + "@example.com",
			Name:    "Test User",
			Role:    domain.CustomerRole,
			Picture: "https://example.com/picture.jpg",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		// Create a token with provider ID for the user
		providerID := "provider-" + uuid.NewString()
		token := &domain.Token{
			ID:         uuid.New(),
			UserID:     user.ID,
			Token:      "test-token-" + uuid.NewString(),
			Type:       domain.TokenTypeAccess,
			ProviderID: providerID,
		}

		tokenRepo := NewTokenRepository(postgres)

		err = tokenRepo.Create(ctx, token)
		require.NoError(t, err)

		retrieved, err := repo.GetByProviderID(ctx, providerID)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)

		_, err = repo.GetByProviderID(ctx, "nonexistent-provider")
		assert.ErrorIs(t, err, customErrors.ErrUserNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		user := &domain.User{
			ID:      uuid.New(),
			Email:   "test-" + uuid.NewString() + "@example.com",
			Name:    "Test User",
			Role:    domain.CustomerRole,
			Picture: "https://example.com/picture.jpg",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		user.Name = "Updated Name"
		err = repo.Update(ctx, user)
		assert.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)

		nonExistentUser := &domain.User{
			ID:    uuid.New(),
			Email: "nonexistent@example.com",
			Name:  "Non Existent",
		}

		err = repo.Update(ctx, nonExistentUser)
		assert.Error(t, err)
		assert.ErrorIs(t, err, customErrors.ErrUserNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		user := &domain.User{
			ID:      uuid.New(),
			Email:   "test-" + uuid.NewString() + "@example.com",
			Name:    "Test User",
			Role:    domain.CustomerRole,
			Picture: "https://example.com/picture.jpg",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		err = repo.Delete(ctx, user.ID.String())
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, user.ID.String())
		assert.ErrorIs(t, err, customErrors.ErrUserNotFound)

		err = repo.Delete(ctx, uuid.New().String())
		assert.ErrorIs(t, err, customErrors.ErrUserNotFound)
	})
}
