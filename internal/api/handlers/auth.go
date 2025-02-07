package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/service"
	"github.com/grocery-service/utils/api"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
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
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	url := h.service.GetAuthURL()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		api.ErrorResponse(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	authResponse, err := h.service.HandleCallback(r.Context(), code)
	if err != nil {
		api.ErrorResponse(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	api.SuccessResponse(w, authResponse, http.StatusOK)
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
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refresh domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&refresh); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authResponse, err := h.service.RefreshToken(r.Context(), refresh.RefreshToken)
	if err != nil {
		api.ErrorResponse(w, "Token refresh failed", http.StatusUnauthorized)
		return
	}

	api.SuccessResponse(w, authResponse, http.StatusOK)
}

// @Summary Revoke token
// @Description Invalidate the current session token
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} api.Response
// @Failure 401 {object} api.Response
// @Router /auth/revoke [post]
func (h *AuthHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		api.ErrorResponse(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.service.RevokeToken(r.Context(), token); err != nil {
		api.ErrorResponse(w, "Token revocation failed", http.StatusInternalServerError)
		return
	}

	api.SuccessResponse(w, nil, http.StatusOK)
}
