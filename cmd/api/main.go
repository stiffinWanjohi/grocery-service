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
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/repository/db"
	"github.com/grocery-service/internal/repository/postgres"
	"github.com/grocery-service/internal/service"
	"github.com/grocery-service/internal/service/notification"
)

// @title           Grocery Service API
// @version         1.0
// @description     API for managing grocery store operations
// @host      localhost:8080
// @BasePath  /api/v1
// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345".
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(database)
	customerRepo := postgres.NewCustomerRepository(database)
	productRepo := postgres.NewProductRepository(database)
	categoryRepo := postgres.NewCategoryRepository(database)
	orderRepo := postgres.NewOrderRepository(database)
	tokenRepo := postgres.NewTokenRepository(database)

	// Initialize services
	notificationService := initializeNotificationService(cfg)
	authService := service.NewAuthService(*cfg, userRepo, tokenRepo, []string{})
	customerService := service.NewCustomerService(customerRepo, userRepo)
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
		authService,
		customerService,
		productService,
		categoryService,
		orderService,
	)

	// Initialize router with middleware
	router := api.NewRouter(
		handlers.authHandler,
		handlers.customerHandler,
		handlers.productHandler,
		handlers.categoryHandler,
		handlers.orderHandler,
		authService,
	)

	startServer(router, cfg.Server.Port)
}

type handlers struct {
	authHandler     *handler.AuthHandler
	customerHandler *handler.CustomerHandler
	productHandler  *handler.ProductHandler
	categoryHandler *handler.CategoryHandler
	orderHandler    *handler.OrderHandler
}

func initializeNotificationService(
	cfg *config.Config,
) notification.NotificationService {
	smsService := notification.NewSMSService(cfg.SMS)
	emailService := notification.NewEmailService(cfg.SMTP)
	return notification.NewCompositeNotificationService(
		smsService,
		emailService,
	)
}

func initializeHandlers(
	authService service.AuthService,
	customerService service.CustomerService,
	productService service.ProductService,
	categoryService service.CategoryService,
	orderService service.OrderService,
) *handlers {
	return &handlers{
		authHandler:     handler.NewAuthHandler(authService),
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

	go func() {
		log.Printf("Server starting on port %d", port)
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
