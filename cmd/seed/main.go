package main

import (
	"context"
	"log"

	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"github.com/grocery-service/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Seed categories
	categoryRepo := postgres.NewCategoryRepository(database)
	if err := seedCategories(ctx, categoryRepo); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}

	// Seed products
	productRepo := postgres.NewProductRepository(database)
	if err := seedProducts(ctx, productRepo); err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	log.Println("Seeding completed successfully")
}

func seedCategories(ctx context.Context, repo *postgres.CategoryRepository) error {
	categories := []string{
		"Fruits & Vegetables",
		"Dairy & Eggs",
		"Meat & Seafood",
		"Bakery",
		"Beverages",
	}

	for _, name := range categories {
		if err := repo.Create(ctx, &domain.Category{Name: name}); err != nil {
			return err
		}
	}
	return nil
}

func seedProducts(ctx context.Context, repo *postgres.ProductRepository) error {
	// Add sample products
	return nil
}
