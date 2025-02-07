package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"github.com/grocery-service/internal/repository/postgres"
)

type categoryProducts struct {
	category string
	products []productData
}

type productData struct {
	name        string
	description string
	price       float64
	stock       int
}

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

	categoryRepo := postgres.NewCategoryRepository(database)
	productRepo := postgres.NewProductRepository(database)

	// Get or create categories and seed products
	categoryMap, err := seedCategories(ctx, categoryRepo)
	if err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}

	if err := seedProducts(ctx, productRepo, categoryMap); err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	log.Println("Seeding completed successfully")
}

func seedCategories(ctx context.Context, repo postgres.CategoryRepository) (map[string]uuid.UUID, error) {
	categoryMap := make(map[string]uuid.UUID)
	categories := []string{
		"Fruits & Vegetables",
		"Dairy & Eggs",
		"Meat & Seafood",
		"Bakery",
		"Beverages",
	}

	for _, name := range categories {
		category := &domain.Category{Name: name}
		if err := repo.Create(ctx, category); err != nil {
			return nil, err
		}
		categoryMap[name] = category.ID
	}

	return categoryMap, nil
}

func seedProducts(ctx context.Context, repo postgres.ProductRepository, categoryMap map[string]uuid.UUID) error {
	productsData := []categoryProducts{
		{
			category: "Fruits & Vegetables",
			products: []productData{
				{"Banana", "Fresh bananas from Ecuador", 0.99, 100},
				{"Apple", "Red delicious apples", 0.75, 150},
				{"Carrot", "Organic carrots", 1.99, 80},
				{"Tomato", "Vine-ripened tomatoes", 2.49, 90},
				{"Spinach", "Fresh baby spinach", 3.99, 50},
			},
		},
		{
			category: "Dairy & Eggs",
			products: []productData{
				{"Milk", "Whole milk, 1 gallon", 3.99, 40},
				{"Eggs", "Large brown eggs, dozen", 4.99, 60},
				{"Cheese", "Cheddar cheese block", 5.99, 30},
				{"Yogurt", "Greek yogurt, plain", 1.99, 45},
				{"Butter", "Unsalted butter", 4.49, 35},
			},
		},
		{
			category: "Meat & Seafood",
			products: []productData{
				{"Chicken Breast", "Boneless skinless, per lb", 5.99, 40},
				{"Ground Beef", "80/20 ground beef, per lb", 6.99, 35},
				{"Salmon", "Fresh Atlantic salmon, per lb", 12.99, 25},
				{"Pork Chops", "Center cut, per lb", 7.99, 30},
				{"Shrimp", "Large frozen shrimp, 1lb bag", 14.99, 20},
			},
		},
		{
			category: "Bakery",
			products: []productData{
				{"Bread", "Whole wheat bread", 3.49, 40},
				{"Bagels", "Plain bagels, 6 pack", 4.99, 30},
				{"Muffins", "Blueberry muffins, 4 pack", 5.99, 25},
				{"Croissants", "Butter croissants, 4 pack", 6.99, 20},
				{"Cookies", "Chocolate chip cookies, dozen", 4.99, 35},
			},
		},
		{
			category: "Beverages",
			products: []productData{
				{"Water", "Spring water, 24 pack", 4.99, 60},
				{"Soda", "Cola, 12 pack", 5.99, 45},
				{"Coffee", "Ground coffee, 12 oz", 8.99, 30},
				{"Tea", "Black tea, 100 bags", 6.99, 40},
				{"Juice", "Orange juice, 1 gallon", 7.99, 35},
			},
		},
	}

	for _, cp := range productsData {
		categoryID, exists := categoryMap[cp.category]
		if !exists {
			continue
		}

		for _, p := range cp.products {
			product := &domain.Product{
				Name:        p.name,
				Description: p.description,
				Price:       p.price,
				Stock:       p.stock,
				CategoryID:  categoryID,
			}

			if err := repo.Create(ctx, product); err != nil {
				return err
			}
		}
	}

	return nil
}
