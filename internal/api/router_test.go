package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	handler "github.com/grocery-service/internal/api/handlers"
	"github.com/grocery-service/internal/api/middleware"
	"github.com/grocery-service/internal/domain"
	serviceMock "github.com/grocery-service/tests/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRouter(t *testing.T) {
	// Setup mock services
	authService := serviceMock.NewAuthService(t)
	customerService := serviceMock.NewCustomerService(t)
	productService := serviceMock.NewProductService(t)
	categoryService := serviceMock.NewCategoryService(t)
	orderService := serviceMock.NewOrderService(t)

	// Setup handlers with mock services
	authHandler := handler.NewAuthHandler(authService)
	customerHandler := handler.NewCustomerHandler(customerService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	orderHandler := handler.NewOrderHandler(orderService)

	authConfig := middleware.AuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
	}

	// Initialize router
	router := NewRouter(
		authHandler,
		customerHandler,
		productHandler,
		categoryHandler,
		orderHandler,
		authConfig,
	)

	// Test cases for routes
	tests := []struct {
		name           string
		method         string
		path           string
		setupAuth      func(t *testing.T, service *serviceMock.AuthService)
		expectedStatus int
	}{
		{
			name:   "Auth - Login Redirect",
			method: http.MethodGet,
			path:   "/auth/login",
			setupAuth: func(t *testing.T, service *serviceMock.AuthService) {
				service.On("GetAuthURL").Return("https://accounts.google.com/o/oauth2/auth")
			},
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:   "Auth - Callback Success",
			method: http.MethodGet,
			path:   "/auth/callback?code=test-code",
			setupAuth: func(t *testing.T, service *serviceMock.AuthService) {
				service.On("HandleCallback", mock.Anything, "test-code").Return(&domain.AuthResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Public - Get Categories",
			method:         http.MethodGet,
			path:           "/api/v1/categories",
			setupAuth:      func(t *testing.T, service *serviceMock.AuthService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Public - List Products",
			method:         http.MethodGet,
			path:           "/api/v1/products",
			setupAuth:      func(t *testing.T, service *serviceMock.AuthService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Protected - Create Product",
			method: http.MethodPost,
			path:   "/api/v1/products",
			setupAuth: func(t *testing.T, service *serviceMock.AuthService) {
				service.On("ValidateToken", mock.Anything, mock.Anything).Return(&domain.User{Role: domain.AdminRole}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Protected - Customer Orders",
			method: http.MethodGet,
			path:   "/api/v1/orders/me",
			setupAuth: func(t *testing.T, service *serviceMock.AuthService) {
				service.On("ValidateToken", mock.Anything, mock.Anything).Return(&domain.User{Role: domain.CustomerRole}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup auth service expectations
			tt.setupAuth(t, authService)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.method != http.MethodGet || tt.path != "/auth/login" {
				req.Header.Set("Authorization", "Bearer test-token")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRouterMiddleware(t *testing.T) {
	// Setup mock services
	authService := serviceMock.NewAuthService(t)
	customerService := serviceMock.NewCustomerService(t)
	productService := serviceMock.NewProductService(t)
	categoryService := serviceMock.NewCategoryService(t)
	orderService := serviceMock.NewOrderService(t)

	// Initialize router with proper handlers
	router := NewRouter(
		handler.NewAuthHandler(authService),
		handler.NewCustomerHandler(customerService),
		handler.NewProductHandler(productService),
		handler.NewCategoryHandler(categoryService),
		handler.NewOrderHandler(orderService),
		middleware.AuthConfig{},
	)

	// Get the middleware stack
	middlewares := getMiddlewareStack(router)

	// Verify essential middleware is present
	assert.Contains(t, middlewares, "RequestID")
	assert.Contains(t, middlewares, "RealIP")
	assert.Contains(t, middlewares, "Logger")
	assert.Contains(t, middlewares, "Recoverer")
	assert.Contains(t, middlewares, "Logging")
}

func getMiddlewareStack(router *chi.Mux) []string {
	var middlewareList []string
	walkFn := func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		for range middlewares {
			middlewareList = append(middlewareList, "middleware")
		}
		return nil
	}

	if err := chi.Walk(router, walkFn); err != nil {
		return []string{}
	}

	return middlewareList
}
