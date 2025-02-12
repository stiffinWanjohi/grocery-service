package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	repoMocks "github.com/grocery-service/tests/mocks/repository"
	serviceMocks "github.com/grocery-service/tests/mocks/service"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAuthService(t *testing.T) {
	mockUserRepo := repoMocks.NewUserRepository(t)
	mockTokenRepo := repoMocks.NewTokenRepository(t)
	cfg := config.Config{
		OAuth: config.OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-secret",
			RedirectURL:  "http://localhost:8080/callback",
			Scopes:       []string{"email", "profile"},
			AllowedUsers: []string{"test@example.com"},
		},
	}
	service := NewAuthService(
		cfg,
		mockUserRepo,
		mockTokenRepo,
	)
	assert.NotNil(t, service)
}

func TestGetAuthURL(t *testing.T) {
	mockUserRepo := repoMocks.NewUserRepository(t)
	mockTokenRepo := repoMocks.NewTokenRepository(t)
	cfg := config.Config{
		OAuth: config.OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-secret",
			RedirectURL:  "http://localhost:8080/callback",
			Scopes:       []string{"email", "profile"},
		},
	}
	service := NewAuthService(
		cfg,
		mockUserRepo,
		mockTokenRepo,
	)
	url := service.GetAuthURL()
	assert.Contains(t, url, "client_id=test-client-id")
	assert.Contains(
		t,
		url,
		"redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback",
	)
}

func TestHandleCallback(t *testing.T) {
	mockAuthService := serviceMocks.NewAuthService(t)
	ctx := context.Background()

	testUser := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
	}
	tests := []struct {
		name          string
		code          string
		setupMocks    func()
		expectedError error
	}{
		{
			name: "successful new user authentication",
			code: "valid-code",
			setupMocks: func() {
				mockAuthService.On("HandleCallback", mock.Anything, "valid-code").
					Return(&domain.AuthResponse{
						AccessToken:  "new-access-token",
						TokenType:    "Bearer",
						ExpiresIn:    3600,
						RefreshToken: "new-refresh-token",
						User:         testUser,
					}, nil)
			},
			expectedError: nil,
		},
		{
			name: "existing user authentication",
			code: "valid-code",
			setupMocks: func() {
				mockAuthService.On("HandleCallback", mock.Anything, "valid-code").
					Return(&domain.AuthResponse{
						AccessToken:  "new-access-token",
						TokenType:    "Bearer",
						ExpiresIn:    3600,
						RefreshToken: "new-refresh-token",
						User:         testUser,
					}, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			service := mockAuthService

			resp, err := service.HandleCallback(
				ctx,
				tt.code,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "new-access-token", resp.AccessToken)
				assert.Equal(t, "Bearer", resp.TokenType)
				assert.Equal(t, testUser, resp.User)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	mockAuthService := serviceMocks.NewAuthService(t)
	ctx := context.Background()
	userID := uuid.New()

	testUser := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	tests := []struct {
		name          string
		refreshToken  string
		setupMocks    func()
		expectedError error
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid-refresh-token",
			setupMocks: func() {
				mockAuthService.On("RefreshToken", mock.Anything, "valid-refresh-token").
					Return(&domain.AuthResponse{
						AccessToken:  "new-access-token",
						TokenType:    "Bearer",
						ExpiresIn:    3600,
						RefreshToken: "new-refresh-token",
						User:         testUser,
					}, nil)
			},
			expectedError: nil,
		},
		{
			name:         "invalid token",
			refreshToken: "invalid-token",
			setupMocks: func() {
				mockAuthService.On("RefreshToken", mock.Anything, "invalid-token").
					Return(nil, customErrors.ErrInvalidToken)
			},
			expectedError: customErrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			service := mockAuthService

			resp, err := service.RefreshToken(
				ctx,
				tt.refreshToken,
			)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "new-access-token", resp.AccessToken)
				assert.Equal(t, "Bearer", resp.TokenType)
				assert.Equal(t, testUser, resp.User)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockUserRepo := repoMocks.NewUserRepository(t)
	mockTokenRepo := repoMocks.NewTokenRepository(t)
	ctx := context.Background()
	userID := uuid.New()

	testUser := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	tests := []struct {
		name          string
		token         string
		setupMocks    func()
		expectedError error
	}{
		{
			name:  "valid token",
			token: "valid-token",
			setupMocks: func() {
				storedToken := &domain.Token{
					UserID: userID,
					Token:  "valid-token",
				}
				mockTokenRepo.On("GetByToken", mock.Anything, "valid-token").
					Return(storedToken, nil)
				mockTokenRepo.On("IsValid", mock.Anything, "valid-token").
					Return(true)
				mockUserRepo.On("GetByID", mock.Anything, userID.String()).
					Return(testUser, nil)
			},
			expectedError: nil,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			setupMocks: func() {
				mockTokenRepo.On("GetByToken", mock.Anything, "invalid-token").
					Return(nil, customErrors.ErrInvalidToken)
			},
			expectedError: customErrors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			cfg := config.Config{
				OAuth: config.OAuthConfig{},
			}

			service := NewAuthService(
				cfg,
				mockUserRepo,
				mockTokenRepo,
			)

			user, err := service.ValidateToken(
				ctx,
				tt.token,
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, testUser.ID, user.ID)
			}
		})
	}
}

func TestRevokeToken(t *testing.T) {
	mockUserRepo := repoMocks.NewUserRepository(t)
	mockTokenRepo := repoMocks.NewTokenRepository(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		token         string
		setupMocks    func()
		expectedError error
	}{
		{
			name:  "successful token revocation",
			token: "valid-token",
			setupMocks: func() {
				mockTokenRepo.On("RevokeToken", mock.Anything, "valid-token").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "revocation error",
			token: "invalid-token",
			setupMocks: func() {
				mockTokenRepo.On("RevokeToken", mock.Anything, "invalid-token").
					Return(fmt.Errorf("revocation failed"))
			},
			expectedError: fmt.Errorf("revocation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			cfg := config.Config{
				OAuth: config.OAuthConfig{},
			}

			service := NewAuthService(
				cfg,
				mockUserRepo,
				mockTokenRepo,
			)

			err := service.RevokeToken(ctx, tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
