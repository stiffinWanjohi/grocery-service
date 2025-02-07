package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthConfig struct {
	JWTSecret     string
	Issuer        string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	AllowedUsers  []string
	TokenDuration time.Duration
}

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
	AdminRole    string     = "admin"
	CustomerRole string     = "customer"
)

type OpenIDConfig struct {
	oauth2Config *oauth2.Config
	userInfoURL  string
}

func newOpenIDConfig(config AuthConfig) *OpenIDConfig {
	return &OpenIDConfig{
		oauth2Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes: []string{
				"openid",
				"profile",
				"email",
			},
			Endpoint: google.Endpoint,
		},
		userInfoURL: "https://openidconnect.googleapis.com/v1/userinfo",
	}
}

func Authentication(config AuthConfig) func(http.Handler) http.Handler {
	oidConfig := newOpenIDConfig(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for OAuth/OpenID endpoints
			if r.URL.Path == "/auth/callback" || r.URL.Path == "/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			// Try JWT authentication first
			if userCtx, ok := validateJWT(r, config); ok {
				next.ServeHTTP(w, r.WithContext(userCtx))
				return
			}

			// Try OpenID token
			if userCtx, ok := validateOpenIDToken(r, oidConfig); ok {
				next.ServeHTTP(w, r.WithContext(userCtx))
				return
			}

			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}
}

func validateJWT(r *http.Request, config AuthConfig) (context.Context, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["iss"] != config.Issuer {
		return nil, false
	}

	email, _ := claims["email"].(string)
	if len(config.AllowedUsers) > 0 {
		isAllowed := false
		for _, allowedEmail := range config.AllowedUsers {
			if email == allowedEmail {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return nil, false
		}
	}

	ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
	ctx = context.WithValue(ctx, UserEmailKey, email)
	if role, ok := claims["role"].(string); ok {
		ctx = context.WithValue(ctx, UserRoleKey, role)
	}

	return ctx, true
}

func validateOpenIDToken(r *http.Request, config *OpenIDConfig) (context.Context, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "OpenID ") {
		return nil, false
	}

	tokenString := strings.TrimPrefix(authHeader, "OpenID ")
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.userInfoURL, nil)
	if err != nil {
		return nil, false
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, false
	}
	defer resp.Body.Close()

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, false
	}

	ctx := context.WithValue(r.Context(), UserIDKey, userInfo.Sub)
	ctx = context.WithValue(ctx, UserEmailKey, userInfo.Email)

	return ctx, true
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(UserIDKey) == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(string)
		if !ok || role != AdminRole {
			http.Error(w, "forbidden: admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireCustomer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(string)
		if !ok || (role != CustomerRole && role != AdminRole) {
			http.Error(w, "forbidden: customer access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
