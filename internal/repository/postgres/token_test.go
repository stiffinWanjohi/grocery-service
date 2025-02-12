package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
)

func TestTokenRepositoryImpl(t *testing.T) {
	postgres := setupTestDB(t, &domain.Token{})
	repo := NewTokenRepository(postgres)
	ctx := context.Background()
	testUser := createTestUser(t, postgres.DB)

	t.Run("Create and GetByToken", func(t *testing.T) {
		token := &domain.Token{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "test-token-" + uuid.NewString(),
			Type:      domain.TokenTypeAccess,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		retrieved, err := repo.GetByToken(ctx, token.Token)
		assert.NoError(t, err)

		assert.Equal(t, token.ID, retrieved.ID)
		assert.Equal(t, token.UserID, retrieved.UserID)
		assert.Equal(t, token.Token, retrieved.Token)
	})

	t.Run("GetByUserAndType", func(t *testing.T) {
		err := postgres.DB.
			Where(
				"user_id = ? AND type = ?",
				testUser.ID,
				domain.TokenTypeAccess,
			).
			Delete(&domain.Token{}).Error
		assert.NoError(t, err)

		token := &domain.Token{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "test-token-" + uuid.NewString(),
			Type:      domain.TokenTypeAccess,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err = repo.Create(ctx, token)
		assert.NoError(t, err)

		retrieved, err := repo.GetByUserAndType(
			ctx,
			testUser.ID.String(),
			domain.TokenTypeAccess,
		)
		assert.NoError(t, err)
		assert.Equal(t, token.ID, retrieved.ID)
		assert.Equal(t, token.Token, retrieved.Token)
		assert.Equal(t, token.Type, retrieved.Type)
		assert.Equal(t, token.UserID, retrieved.UserID)
	})

	t.Run("GetByProviderID", func(t *testing.T) {
		providerID := "test-provider-" + uuid.NewString()
		token := &domain.Token{
			ID:         uuid.New(),
			UserID:     testUser.ID,
			Token:      "test-token-" + uuid.NewString(),
			Type:       domain.TokenTypeAccess,
			ExpiresAt:  time.Now().Add(time.Hour),
			ProviderID: providerID,
		}

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		retrieved, err := repo.GetByProviderID(ctx, providerID)
		assert.NoError(t, err)
		assert.Equal(t, token.ID, retrieved.ID)
	})

	t.Run("RevokeToken", func(t *testing.T) {
		token := &domain.Token{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "test-token-" + uuid.NewString(),
			Type:      domain.TokenTypeAccess,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err := repo.Create(ctx, token)
		assert.NoError(t, err)

		err = repo.RevokeToken(ctx, token.Token)
		assert.NoError(t, err)

		// Try to get revoked token
		_, err = repo.GetByToken(ctx, token.Token)
		assert.ErrorIs(t, err, customErrors.ErrTokenNotFound)
	})

	t.Run("DeleteExpiredTokens", func(t *testing.T) {
		expiredToken := &domain.Token{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "expired-token-" + uuid.NewString(),
			Type:      domain.TokenTypeAccess,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}

		err := repo.Create(ctx, expiredToken)
		assert.NoError(t, err)

		err = repo.DeleteExpiredTokens(ctx)
		assert.NoError(t, err)

		// Try to get deleted token
		_, err = repo.GetByToken(ctx, expiredToken.Token)
		assert.ErrorIs(t, err, customErrors.ErrTokenNotFound)
	})

	t.Run("IsValid", func(t *testing.T) {
		validToken := &domain.Token{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "valid-token-" + uuid.NewString(),
			Type:      domain.TokenTypeAccess,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err := repo.Create(ctx, validToken)
		assert.NoError(t, err)

		isValid := repo.IsValid(ctx, validToken.Token)
		assert.True(t, isValid)

		// Test invalid token
		isValid = repo.IsValid(ctx, "non-existent-token")
		assert.False(t, isValid)
	})
}
