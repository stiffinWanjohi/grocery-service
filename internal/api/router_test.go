package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	handler "github.com/grocery-service/internal/api/handlers"
	"github.com/grocery-service/internal/api/middleware"
	serviceMock "github.com/grocery-service/tests/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	// Setup mock services
	customerService := serviceMock.NewCustomerService(t)
	productService := serviceMock.NewProductService(t)
	categoryService := serviceMock.NewCategoryService(t)
	orderService := serviceMock.NewOrderService(t)

	// Setup handlers with mock services
	customerHandler := handler.NewCustomerHandler(customerService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	orderHandler := handler.NewOrderHandler(orderService)

	authConfig := middleware.AuthConfig{
		JWTSecret: "test-secret",
		Issuer:    "test-issuer",
	}

	// Initialize router
	router := NewRouter(
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
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Public - Get Categories",
			method:         http.MethodGet,
			path:           "/categories",
			authenticated:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Public - Get Products",
			method:         http.MethodGet,
			path:           "/products",
			authenticated:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Protected - Get Customers without Auth",
			method:         http.MethodGet,
			path:           "/customers",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Protected - Get Orders without Auth",
			method:         http.MethodGet,
			path:           "/orders",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authenticated {
				// Add authentication token if needed
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
	customerService := serviceMock.NewCustomerService(t)
	productService := serviceMock.NewProductService(t)
	categoryService := serviceMock.NewCategoryService(t)
	orderService := serviceMock.NewOrderService(t)

	// Initialize router with proper handlers
	router := NewRouter(
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

// Helper function to get middleware names from the router
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
