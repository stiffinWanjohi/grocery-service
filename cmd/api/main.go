package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grocery-service/internal/api"
	handler "github.com/grocery-service/internal/api/handlers"
	"github.com/grocery-service/internal/api/middleware"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/repository/db"
	"github.com/grocery-service/internal/repository/postgres"
	"github.com/grocery-service/internal/service"
	"github.com/grocery-service/internal/service/notification"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	customerRepo := postgres.NewCustomerRepository(database)
	productRepo := postgres.NewProductRepository(database)
	categoryRepo := postgres.NewCategoryRepository(database)
	orderRepo := postgres.NewOrderRepository(database)

	// Initialize notification service
	smsService, err := notification.NewSMSService(cfg.SMS)
	if err != nil {
		log.Fatalf("Failed to initialize SMS service: %v", err)
	}

	// Initialize services
	customerService := service.NewCustomerService(customerRepo)
	productService := service.NewProductService(productRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	orderService := service.NewOrderService(orderRepo, productRepo, customerRepo, smsService, database.DB)

	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(customerService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	orderHandler := handler.NewOrderHandler(orderService)

	// Initialize router
	router := api.NewRouter(
		customerHandler,
		productHandler,
		categoryHandler,
		orderHandler,
		middleware.AuthConfig{
			JWTSecret: cfg.JWT.Secret,
			Issuer:    cfg.JWT.Issuer,
		},
	)

	// Configure server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
