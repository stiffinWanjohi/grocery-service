package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/grocery-service/internal/service"
)

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Issuer       string // Okta domain, e.g., "https://{yourOktaDomain}/oauth2/default"
}

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
	AdminRole    string     = "admin"
	CustomerRole string     = "customer"
)

func Authentication(
	authService service.AuthService,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			if r.URL.Path == "/api/v1/auth/callback" ||
				r.URL.Path == "/api/v1/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			token, err := extractBearerToken(r)
			if err != nil {
				http.Error(
					w,
					"unauthorized: missing or malformed token",
					http.StatusUnauthorized,
				)
				return
			}

			userInfo, err := authService.GetUserInfo(r.Context(), token)
			if err != nil {
				http.Error(
					w,
					"unauthorized: invalid token",
					http.StatusUnauthorized,
				)
				return
			}

			// Add user information to the request context
			ctx := context.WithValue(r.Context(), UserIDKey, userInfo.ID)
			ctx = context.WithValue(ctx, UserEmailKey, userInfo.Email)
			ctx = context.WithValue(ctx, UserRoleKey, userInfo.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("missing or malformed authorization header")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

type UserInfo struct {
	Sub   string      `json:"sub"`
	Email string      `json:"email"`
	Role  interface{} `json:"role"` // Customize this based on your Okta claims
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
			http.Error(
				w,
				"forbidden: admin access required",
				http.StatusForbidden,
			)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// func RequireCustomer(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		role, ok := r.Context().Value(UserRoleKey).(string)
// 		if !ok || (role != CustomerRole && role != AdminRole) {
// 			http.Error(
// 				w,
// 				"forbidden: customer access required",
// 				http.StatusForbidden,
// 			)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
