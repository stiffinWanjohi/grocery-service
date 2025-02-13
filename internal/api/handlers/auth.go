package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/service"
	"github.com/grocery-service/utils/api"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(
	service service.AuthService,
) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/login", h.Login)
	r.Get("/callback", h.Callback)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/revoke", h.RevokeToken)

	return r
}

// @Summary OpenID Connect login
// @Description Redirect to OpenID provider login page
// @Tags auth
// @Produce json
// @Success 302
// @Router /auth/login [get]
func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Set content negotiation headers
	w.Header().Set("Accept", "text/html,application/json")

	// Get the auth URL
	authURL := h.service.GetAuthURL()

	// Check if it's an API request
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"auth_url": authURL,
		}); err != nil {
			if err := api.ErrorResponse(
				w,
				"Failed to encode response",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
			return
		}
		return
	}

	// Otherwise redirect to Auth0
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// @Summary OpenID Connect callback
// @Description Handle OpenID callback and create session
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code"
// @Success 200 {object} api.Response{data=domain.AuthResponse}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.Response
// @Router /auth/callback [get]
func (h *AuthHandler) Callback(
	w http.ResponseWriter,
	r *http.Request,
) {
	code := r.URL.Query().Get("code")
	if code == "" {
		if err := api.ErrorResponse(
			w,
			"Missing authorization code",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	// Log the incoming code
	fmt.Printf("Received authorization code: %s\n", code)

	authResponse, err := h.service.HandleCallback(r.Context(), code)
	if err != nil {
		// Log the specific error
		fmt.Printf("Auth callback error: %v\n", err)

		statusCode := http.StatusUnauthorized
		message := "Authentication failed"

		// Provide more specific error messages based on the error type
		if strings.Contains(err.Error(), "failed to exchange token") {
			message = "Failed to exchange authorization code for token"
		} else if strings.Contains(err.Error(), "failed to fetch user info") {
			message = "Failed to fetch user information"
		}

		if err := api.ErrorResponse(w, message, statusCode); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	// Log successful authentication
	fmt.Printf(
		"Authentication successful for user: %s\n",
		authResponse.User.Email,
	)

	if err := api.SuccessResponse(w, authResponse, http.StatusOK); err != nil {
		if err := api.ErrorResponse(w, "Failed to send response", http.StatusInternalServerError); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}
}

// @Summary Refresh token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body domain.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} api.Response{data=domain.AuthResponse}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(
	w http.ResponseWriter,
	r *http.Request,
) {
	var refresh domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&refresh); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	authResponse, err := h.service.RefreshToken(
		r.Context(),
		refresh.RefreshToken,
	)
	if err != nil {
		if err := api.ErrorResponse(
			w,
			"Token refresh failed",
			http.StatusUnauthorized,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	if err := api.SuccessResponse(w, authResponse, http.StatusOK); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}
}

// @Summary Revoke token
// @Description Invalidate the current session token
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} api.Response
// @Failure 401 {object} api.Response
// @Router /auth/revoke [post]
func (h *AuthHandler) RevokeToken(
	w http.ResponseWriter,
	r *http.Request,
) {
	token := r.Header.Get("Authorization")
	if token == "" {
		if err := api.ErrorResponse(
			w,
			"Missing authorization token",
			http.StatusUnauthorized,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.service.RevokeToken(r.Context(), token); err != nil {
		if err := api.ErrorResponse(
			w,
			"Token revocation failed",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	if err := api.SuccessResponse(w, nil, http.StatusOK); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}
}
