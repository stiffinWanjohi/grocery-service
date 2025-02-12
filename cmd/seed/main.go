package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	"github.com/grocery-service/internal/repository/postgres"
)

var (
	// Super Categories (Level 0)
	produceID     = uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	dairyID       = uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	bakeryID      = uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	meatSeafoodID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
	pantryID      = uuid.MustParse("550e8400-e29b-41d4-a716-446655440005")

	// Subcategories (Level 1)
	// Produce subcategories
	fruitsID     = uuid.MustParse("550e8400-e29b-41d4-a716-446655440006")
	vegetablesID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440007")
	herbsID      = uuid.MustParse("550e8400-e29b-41d4-a716-446655440008")

	// Dairy subcategories
	milkID   = uuid.MustParse("550e8400-e29b-41d4-a716-446655440009")
	cheeseID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440010")
	yogurtID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440011")

	// Bakery subcategories
	breadID  = uuid.MustParse("550e8400-e29b-41d4-a716-446655440012")
	pastryID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440013")
	cakesID  = uuid.MustParse("550e8400-e29b-41d4-a716-446655440014")

	// Meat & Seafood subcategories
	poultryID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440015")
	beefID    = uuid.MustParse("550e8400-e29b-41d4-a716-446655440016")
	seafoodID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440017")

	// Pantry subcategories
	spicesID      = uuid.MustParse("550e8400-e29b-41d4-a716-446655440018")
	cannedGoodsID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440019")
	condimentsID  = uuid.MustParse("550e8400-e29b-41d4-a716-446655440020")
)

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

	if err := seedCategories(ctx, categoryRepo); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}

	if err := seedProducts(ctx, productRepo); err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	log.Println("Seeding completed successfully")
}

func seedCategories(ctx context.Context, repo postgres.CategoryRepository) error {
	categories := []struct {
		id       uuid.UUID
		name     string
		parentID *uuid.UUID
		level    int
	}{
		// Super categories (Level 0)
		{produceID, "Produce", nil, 0},
		{dairyID, "Dairy", nil, 0},
		{bakeryID, "Bakery", nil, 0},
		{meatSeafoodID, "Meat & Seafood", nil, 0},
		{pantryID, "Pantry", nil, 0},

		// Produce subcategories
		{fruitsID, "Fruits", &produceID, 1},
		{vegetablesID, "Vegetables", &produceID, 1},
		{herbsID, "Herbs", &produceID, 1},

		// Dairy subcategories
		{milkID, "Milk & Cream", &dairyID, 1},
		{cheeseID, "Cheese", &dairyID, 1},
		{yogurtID, "Yogurt", &dairyID, 1},

		// Bakery subcategories
		{breadID, "Bread", &bakeryID, 1},
		{pastryID, "Pastries", &bakeryID, 1},
		{cakesID, "Cakes & Desserts", &bakeryID, 1},

		// Meat & Seafood subcategories
		{poultryID, "Poultry", &meatSeafoodID, 1},
		{beefID, "Beef", &meatSeafoodID, 1},
		{seafoodID, "Seafood", &meatSeafoodID, 1},

		// Pantry subcategories
		{spicesID, "Spices & Seasonings", &pantryID, 1},
		{cannedGoodsID, "Canned Goods", &pantryID, 1},
		{condimentsID, "Condiments", &pantryID, 1},
	}

	for _, c := range categories {
		path := c.name
		if c.parentID != nil {
			for _, pc := range categories {
				if pc.id == *c.parentID {
					path = fmt.Sprintf("%s/%s", pc.name, c.name)
					break
				}
			}
		}

		category := &domain.Category{
			ID:       c.id,
			Name:     c.name,
			ParentID: c.parentID,
			Level:    c.level,
			Path:     path,
		}
		if err := repo.Create(ctx, category); err != nil {
			return err
		}
	}

	return nil
}

func seedProducts(ctx context.Context, repo postgres.ProductRepository) error {
	productsData := map[uuid.UUID][]productData{
		fruitsID: {
			{"Banana", "Fresh bananas from Ecuador", 0.99, 100},
			{"Apple", "Red delicious apples", 0.75, 150},
			{"Orange", "Sweet navel oranges", 1.29, 120},
		},
		vegetablesID: {
			{"Carrot", "Organic carrots", 1.99, 80},
			{"Tomato", "Vine-ripened tomatoes", 2.49, 90},
			{"Spinach", "Fresh baby spinach", 3.99, 50},
		},
		herbsID: {
			{"Basil", "Fresh basil leaves", 2.99, 40},
			{"Cilantro", "Fresh cilantro bunch", 1.99, 45},
			{"Mint", "Fresh mint leaves", 2.49, 35},
		},
		milkID: {
			{"Whole Milk", "Fresh whole milk, 1 gallon", 3.99, 40},
			{"Heavy Cream", "Fresh heavy cream", 4.29, 25},
			{"Half & Half", "Fresh half & half", 3.49, 30},
		},
		cheeseID: {
			{"Cheddar", "Sharp cheddar cheese", 5.99, 30},
			{"Mozzarella", "Fresh mozzarella", 4.99, 35},
			{"Swiss", "Swiss cheese", 6.99, 25},
		},
		yogurtID: {
			{"Greek Yogurt", "Plain greek yogurt", 1.99, 45},
			{"Vanilla Yogurt", "Vanilla flavored yogurt", 2.49, 40},
			{"Strawberry Yogurt", "Strawberry yogurt", 2.49, 40},
		},
		breadID: {
			{"Whole Wheat", "Whole wheat bread", 3.49, 40},
			{"Sourdough", "Fresh sourdough loaf", 4.99, 30},
			{"Rye Bread", "Fresh rye bread", 4.49, 25},
		},
		pastryID: {
			{"Croissants", "Butter croissants, 4 pack", 6.99, 20},
			{"Danish", "Assorted danish pastries", 5.99, 25},
			{"Muffins", "Blueberry muffins, 4 pack", 5.99, 30},
		},
		cakesID: {
			{"Chocolate Cake", "Rich chocolate cake", 24.99, 15},
			{"Cheesecake", "New York style cheesecake", 19.99, 20},
			{"Cupcakes", "Assorted cupcakes, 6 pack", 8.99, 25},
		},
		poultryID: {
			{"Chicken Breast", "Boneless skinless, per lb", 5.99, 40},
			{"Turkey", "Ground turkey, per lb", 5.99, 35},
			{"Chicken Wings", "Fresh chicken wings, per lb", 4.99, 45},
		},
		beefID: {
			{"Ground Beef", "80/20 ground beef, per lb", 6.99, 35},
			{"Ribeye", "Choice ribeye steak, per lb", 15.99, 20},
			{"Beef Tenderloin", "Premium tenderloin, per lb", 19.99, 15},
		},
		seafoodID: {
			{"Salmon", "Fresh Atlantic salmon, per lb", 12.99, 25},
			{"Shrimp", "Large frozen shrimp, 1lb bag", 14.99, 20},
			{"Cod", "Fresh cod fillet, per lb", 11.99, 30},
		},
		spicesID: {
			{"Black Pepper", "Ground black pepper", 3.99, 85},
			{"Cinnamon", "Ground cinnamon", 4.99, 70},
			{"Garlic Powder", "Pure garlic powder", 3.99, 75},
		},
		cannedGoodsID: {
			{"Diced Tomatoes", "Canned diced tomatoes", 1.99, 100},
			{"Black Beans", "Canned black beans", 1.49, 120},
			{"Tuna", "Chunk light tuna in water", 1.99, 90},
		},
		condimentsID: {
			{"Mayonnaise", "Classic mayonnaise", 4.99, 60},
			{"Ketchup", "Tomato ketchup", 3.99, 70},
			{"Mustard", "Yellow mustard", 2.99, 65},
		},
	}

	for categoryID, products := range productsData {
		for _, p := range products {
			product := &domain.Product{
				ID:          uuid.New(),
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
