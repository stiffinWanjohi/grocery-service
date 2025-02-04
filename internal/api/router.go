package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(customMiddleware.Logging)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
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
	})

	return r
}
