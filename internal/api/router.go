package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	handler "github.com/grocery-service/internal/api/handlers"
	customMiddleware "github.com/grocery-service/internal/api/middleware"
)

func NewRouter(
	customerHandler *handler.CustomerHandler,
	productHandler *handler.ProductHandler,
	categoryHandler *handler.CategoryHandler,
	orderHandler *handler.OrderHandler,
	authConfig customMiddleware.AuthConfig,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.Logging)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Mount("/categories", categoryHandler.Routes())
		r.Mount("/products", productHandler.Routes())
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.Authentication(authConfig))
		r.Mount("/customers", customerHandler.Routes())
		r.Mount("/orders", orderHandler.Routes())
	})

	return r
}
