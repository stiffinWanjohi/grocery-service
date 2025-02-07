package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	handler "github.com/grocery-service/internal/api/handlers"
	customMiddleware "github.com/grocery-service/internal/api/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	authHandler *handler.AuthHandler,
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Swagger endpoint
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Auth routes (OpenID Connect endpoints)
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
		r.Post("/refresh", authHandler.RefreshToken)
		r.With(customMiddleware.Authentication(authConfig)).Post("/revoke", authHandler.RevokeToken)
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Mount("/categories", categoryHandler.Routes())
			r.Get("/products", productHandler.List)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Authentication(authConfig))

			// Customer routes
			r.Mount("/customers", customerHandler.Routes())

			// Product routes
			r.Route("/products", func(r chi.Router) {
				r.Get("/{id}", productHandler.GetByID)

				// Admin only routes
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.RequireAdmin)
					r.Post("/", productHandler.Create)
					r.Put("/{id}", productHandler.Update)
					r.Delete("/{id}", productHandler.Delete)
				})
			})

			// Order routes
			r.Route("/orders", func(r chi.Router) {
				// Customer routes
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.RequireCustomer)
					r.Post("/", orderHandler.Create)
					r.Get("/me", orderHandler.ListByCustomerID)
				})

				// Admin routes
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.RequireAdmin)
					r.Get("/", orderHandler.List)
					r.Get("/{id}", orderHandler.GetByID)
					r.Put("/{id}/status", orderHandler.UpdateStatus)
				})
			})
		})
	})

	return r
}
