package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
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
	customerHandler := handler.NewCustomerHandler(
		customerService,
	)
	productHandler := handler.NewProductHandler(
		productService,
	)
	categoryHandler := handler.NewCategoryHandler(
		categoryService,
	)
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

	// customerID := uuid.New()

	// Test cases for routes
	tests := []struct {
		name           string
		method         string
		path           string
		setupAuth      func(t *testing.T, service *serviceMock.AuthService)
		setupCategory  func(t *testing.T, service *serviceMock.CategoryService)
		setupProduct   func(t *testing.T, service *serviceMock.ProductService)
		expectedStatus int
	}{
		{
			name:   "Auth - Login Redirect",
			method: http.MethodGet,
			path:   "/auth/login",
			setupAuth: func(_ *testing.T, service *serviceMock.AuthService) {
				service.On("GetAuthURL").
					Return("https://accounts.google.com/o/oauth2/auth")
			},
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:   "Auth - Callback Success",
			method: http.MethodGet,
			path:   "/auth/callback?code=test-code",
			setupAuth: func(_ *testing.T, service *serviceMock.AuthService) {
				service.On("HandleCallback", mock.Anything, "test-code").
					Return(&domain.AuthResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Public - Get Categories",
			method:    http.MethodGet,
			path:      "/api/v1/categories",
			setupAuth: func(_ *testing.T, _ *serviceMock.AuthService) {},
			setupCategory: func(_ *testing.T, service *serviceMock.CategoryService) {
				service.On("List", mock.Anything).
					Return([]domain.Category{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Public - List Products",
			method:    http.MethodGet,
			path:      "/api/v1/products",
			setupAuth: func(_ *testing.T, _ *serviceMock.AuthService) {},
			setupProduct: func(_ *testing.T, service *serviceMock.ProductService) {
				service.On("List", mock.Anything).
					Return([]domain.Product{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		// {
		// 	name:   "Protected - Create Product",
		// 	method: http.MethodPost,
		// 	path:   "/api/v1/products",
		// 	setupAuth: func(_ *testing.T, service *serviceMock.AuthService) {
		// 		service.On("ValidateToken", mock.Anything, "test-token").
		// 			Return(&domain.User{Role: domain.AdminRole}, nil).
		// 			Once()
		// 	},
		// 	expectedStatus: http.StatusOK,
		// },
		// {
		// 	name:   "Protected - Customer Orders",
		// 	method: http.MethodGet,
		// 	path:   "/api/v1/orders/customer/" + customerID.String(),
		// 	setupAuth: func(_ *testing.T, service *serviceMock.AuthService) {
		// 		service.On("ValidateToken", mock.Anything, mock.MatchedBy(func(token string) bool {
		// 			return strings.TrimPrefix(
		// 				token,
		// 				"Bearer ",
		// 			) == "test-token"
		// 		})).
		// 			Return(&domain.User{
		// 				ID:   customerID,
		// 				Role: domain.CustomerRole,
		// 			}, nil).
		// 			Once()
		// 	},
		// 	expectedStatus: http.StatusOK,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService.ExpectedCalls = nil
			customerService.ExpectedCalls = nil
			productService.ExpectedCalls = nil
			categoryService.ExpectedCalls = nil
			orderService.ExpectedCalls = nil

			tt.setupAuth(t, authService)

			if tt.setupCategory != nil {
				tt.setupCategory(t, categoryService)
			}

			if tt.setupProduct != nil {
				tt.setupProduct(t, productService)
			}

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			if tt.method == http.MethodPost ||
				strings.Contains(tt.path, "/orders/") {
				req.Header.Set(
					"Authorization",
					"test-token",
				)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.method == http.MethodPost ||
				strings.Contains(tt.path, "/orders/") {
				authService.AssertExpectations(t)
			}
		})
	}
}

func TestRouterMiddleware(t *testing.T) {
	authService := serviceMock.NewAuthService(t)
	customerService := serviceMock.NewCustomerService(t)
	productService := serviceMock.NewProductService(t)
	categoryService := serviceMock.NewCategoryService(t)
	orderService := serviceMock.NewOrderService(t)

	router := NewRouter(
		handler.NewAuthHandler(authService),
		handler.NewCustomerHandler(customerService),
		handler.NewProductHandler(productService),
		handler.NewCategoryHandler(categoryService),
		handler.NewOrderHandler(orderService),
		middleware.AuthConfig{},
	)

	middlewares := getMiddlewareStack(router)

	assert.Contains(t, middlewares, "RequestID")
	assert.Contains(t, middlewares, "RealIP")
	assert.Contains(t, middlewares, "Logger")
	assert.Contains(t, middlewares, "Recoverer")
	assert.Contains(t, middlewares, "Logging")
}

func getMiddlewareStack(router *chi.Mux) []string {
	var middlewareList []string
	walkFn := func(_, _ string, _ http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		for _, mw := range middlewares {
			middlewareName := runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).
				Name()
			parts := strings.Split(middlewareName, ".")
			name := parts[len(parts)-1]
			name = strings.TrimSuffix(name, "-fm")
			middlewareList = append(middlewareList, name)
		}
		return nil
	}

	if err := chi.Walk(router, walkFn); err != nil {
		return []string{}
	}

	seen := make(map[string]bool)
	unique := []string{}
	for _, name := range middlewareList {
		if !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}

	return unique
}
