package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	repository "github.com/grocery-service/internal/repository/postgres"
	customErrors "github.com/grocery-service/utils/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService interface {
	GetAuthURL() string
	HandleCallback(ctx context.Context, code string) (*domain.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
	RevokeToken(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
}

type authService struct {
	oauth2Config *oauth2.Config
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	allowedUsers []string
}

func NewAuthService(
	cfg config.Config,
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
) AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.ClientID,
		ClientSecret: cfg.OAuth.ClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes:       cfg.OAuth.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &authService{
		oauth2Config: oauthConfig,
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		allowedUsers: cfg.OAuth.AllowedUsers,
	}
}

func (s *authService) GetAuthURL() string {
	return s.oauth2Config.AuthCodeURL("state")
}

func (s *authService) createUserAndTokens(ctx context.Context, userInfo *domain.UserInfo, oauth2Token *oauth2.Token) (*domain.User, error) {
	user := &domain.User{
		ID:        uuid.New(),
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		Picture:   userInfo.Picture,
		Role:      domain.CustomerRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.storeTokens(ctx, user.ID, userInfo.ID, oauth2Token); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) storeTokens(ctx context.Context, userID uuid.UUID, providerID string, oauth2Token *oauth2.Token) error {
	accessToken := &domain.Token{
		UserID:     userID,
		Token:      oauth2Token.AccessToken,
		Type:       domain.TokenTypeAccess,
		ExpiresAt:  oauth2Token.Expiry,
		Provider:   "google",
		ProviderID: providerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, accessToken); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	if oauth2Token.RefreshToken != "" {
		refreshToken := &domain.Token{
			UserID:     userID,
			Token:      oauth2Token.RefreshToken,
			Type:       domain.TokenTypeRefresh,
			ExpiresAt:  time.Now().AddDate(0, 1, 0),
			Provider:   "google",
			ProviderID: providerID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
			return fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	return nil
}

func (s *authService) HandleCallback(ctx context.Context, code string) (*domain.AuthResponse, error) {
	oauth2Token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	userInfo, err := s.getUserInfo(ctx, oauth2Token)
	if err != nil {
		return nil, err
	}

	if len(s.allowedUsers) > 0 {
		isAllowed := false
		for _, allowedEmail := range s.allowedUsers {
			if userInfo.Email == allowedEmail {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return nil, customErrors.ErrUnauthorized
		}
	}

	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		user, err = s.createUserAndTokens(ctx, userInfo, oauth2Token)
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.storeTokens(ctx, user.ID, userInfo.ID, oauth2Token); err != nil {
			return nil, err
		}
	}

	return &domain.AuthResponse{
		AccessToken:  oauth2Token.AccessToken,
		TokenType:    oauth2Token.TokenType,
		ExpiresIn:    int(time.Until(oauth2Token.Expiry).Seconds()),
		RefreshToken: oauth2Token.RefreshToken,
		User:         user,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	token, err := s.tokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, customErrors.ErrInvalidToken
	}

	if !s.tokenRepo.IsValid(ctx, refreshToken) {
		return nil, customErrors.ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID.String())
	if err != nil {
		return nil, customErrors.ErrUserNotFound
	}

	oauth2Token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := s.oauth2Config.TokenSource(ctx, oauth2Token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if err := s.storeTokens(ctx, user.ID, token.ProviderID, newToken); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  newToken.AccessToken,
		TokenType:    newToken.TokenType,
		ExpiresIn:    int(time.Until(newToken.Expiry).Seconds()),
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string) error {
	return s.tokenRepo.RevokeToken(ctx, token)
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	storedToken, err := s.tokenRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, customErrors.ErrInvalidToken
	}

	if !s.tokenRepo.IsValid(ctx, token) {
		return nil, customErrors.ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID.String())
	if err != nil {
		return nil, customErrors.ErrUserNotFound
	}

	return user, nil
}

func (s *authService) getUserInfo(ctx context.Context, token *oauth2.Token) (*domain.UserInfo, error) {
	client := s.oauth2Config.Client(ctx, token)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo domain.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}
