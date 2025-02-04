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
	notificationService := initializeNotificationService(cfg)

	// Initialize services
	customerService := service.NewCustomerService(customerRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	orderService := service.NewOrderService(
		orderRepo,
		productRepo,
		customerRepo,
		notificationService,
	)

	// Initialize API handlers
	handlers := initializeHandlers(
		customerService,
		productService,
		categoryService,
		orderService,
	)

	// Initialize router with middleware
	router := api.NewRouter(
		handlers.customerHandler,
		handlers.productHandler,
		handlers.categoryHandler,
		handlers.orderHandler,
		middleware.AuthConfig{
			JWTSecret: cfg.JWT.Secret,
			Issuer:    cfg.JWT.Issuer,
		},
	)

	// Start server
	startServer(router, cfg.Server.Port)
}

type handlers struct {
	customerHandler *handler.CustomerHandler
	productHandler  *handler.ProductHandler
	categoryHandler *handler.CategoryHandler
	orderHandler    *handler.OrderHandler
}

func initializeNotificationService(cfg *config.Config) notification.NotificationService {
	smsService := notification.NewSMSService(cfg.SMS)
	emailService := notification.NewEmailService(cfg.SMTP)
	return notification.NewCompositeNotificationService(
		smsService,
		emailService,
	)
}

func initializeHandlers(
	customerService service.CustomerService,
	productService service.ProductService,
	categoryService service.CategoryService,
	orderService service.OrderService,
) *handlers {
	return &handlers{
		customerHandler: handler.NewCustomerHandler(customerService),
		productHandler:  handler.NewProductHandler(productService),
		categoryHandler: handler.NewCategoryHandler(categoryService),
		orderHandler:    handler.NewOrderHandler(orderService),
	}
}

func startServer(handler http.Handler, port int) {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", port)
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
