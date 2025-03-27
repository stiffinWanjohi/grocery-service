package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	// Product IDs
	// Fruits
	bananaID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	appleID  = uuid.MustParse("660e8400-e29b-41d4-a716-446655440002")
	orangeID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440003")

	// Vegetables
	carrotID  = uuid.MustParse("660e8400-e29b-41d4-a716-446655440004")
	tomatoID  = uuid.MustParse("660e8400-e29b-41d4-a716-446655440005")
	spinachID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440006")

	// Herbs
	basilID    = uuid.MustParse("660e8400-e29b-41d4-a716-446655440007")
	cilantroID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440008")
	mintID     = uuid.MustParse("660e8400-e29b-41d4-a716-446655440009")

	// Milk
	wholeMilkID   = uuid.MustParse("660e8400-e29b-41d4-a716-446655440010")
	heavyCreamID  = uuid.MustParse("660e8400-e29b-41d4-a716-446655440011")
	halfAndHalfID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440012")

	// Cheese
	cheddarID    = uuid.MustParse("660e8400-e29b-41d4-a716-446655440013")
	mozzarellaID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440014")
	swissID      = uuid.MustParse("660e8400-e29b-41d4-a716-446655440015")

	// Yogurt
	greekYogurtID      = uuid.MustParse("660e8400-e29b-41d4-a716-446655440016")
	vanillaYogurtID    = uuid.MustParse("660e8400-e29b-41d4-a716-446655440017")
	strawberryYogurtID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440018")

	// Bread
	wholeWheatID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440019")
	sourdoughID  = uuid.MustParse("660e8400-e29b-41d4-a716-446655440020")
	ryeBreadID   = uuid.MustParse("660e8400-e29b-41d4-a716-446655440021")

	// Pastry
	croissantsID = uuid.MustParse("660e8400-e29b-41d4-a716-446655440022")
	danishID     = uuid.MustParse("660e8400-e29b-41d4-a716-446655440023")
	muffinsID    = uuid.MustParse("660e8400-e29b-41d4-a716-446655440024")

	// User IDs
	testUserID = uuid.MustParse("880e8400-e29b-41d4-a716-446655440001")

	// Customer IDs
	testCustomerID = uuid.MustParse("770e8400-e29b-41d4-a716-446655440001")
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

	categoryRepo := postgres.NewCategoryRepository(database)
	productRepo := postgres.NewProductRepository(database)
	userRepo := postgres.NewUserRepository(database)
	customerRepo := postgres.NewCustomerRepository(database)

	if err := seedCategories(ctx, categoryRepo); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}

	if err := seedProducts(ctx, productRepo); err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	if err := seedUsers(ctx, userRepo); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	if err := seedCustomers(ctx, customerRepo); err != nil {
		log.Fatalf("Failed to seed customers: %v", err)
	}

	log.Println("Seeding completed successfully")
}

func seedCategories(
	ctx context.Context,
	repo postgres.CategoryRepository,
) error {
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
	products := []struct {
		id          uuid.UUID
		name        string
		description string
		price       float64
		stock       int
		categoryID  uuid.UUID
	}{
		// Fruits
		{bananaID, "Banana", "Fresh bananas from Ecuador", 0.99, 100, fruitsID},
		{appleID, "Apple", "Red delicious apples", 0.75, 150, fruitsID},
		{orangeID, "Orange", "Sweet navel oranges", 1.29, 120, fruitsID},

		// Vegetables
		{carrotID, "Carrot", "Organic carrots", 1.99, 80, vegetablesID},
		{tomatoID, "Tomato", "Vine-ripened tomatoes", 2.49, 90, vegetablesID},
		{spinachID, "Spinach", "Fresh baby spinach", 3.99, 50, vegetablesID},

		// Herbs
		{basilID, "Basil", "Fresh basil leaves", 2.99, 40, herbsID},
		{cilantroID, "Cilantro", "Fresh cilantro bunch", 1.99, 45, herbsID},
		{mintID, "Mint", "Fresh mint leaves", 2.49, 35, herbsID},

		// Milk
		{wholeMilkID, "Whole Milk", "Fresh whole milk, 1 gallon", 3.99, 40, milkID},
		{heavyCreamID, "Heavy Cream", "Fresh heavy cream", 4.29, 25, milkID},
		{halfAndHalfID, "Half & Half", "Fresh half & half", 3.49, 30, milkID},

		// Cheese
		{cheddarID, "Cheddar", "Sharp cheddar cheese", 5.99, 30, cheeseID},
		{mozzarellaID, "Mozzarella", "Fresh mozzarella", 4.99, 35, cheeseID},
		{swissID, "Swiss", "Swiss cheese", 6.99, 25, cheeseID},

		// Yogurt
		{greekYogurtID, "Greek Yogurt", "Plain greek yogurt", 1.99, 45, yogurtID},
		{vanillaYogurtID, "Vanilla Yogurt", "Vanilla flavored yogurt", 2.49, 40, yogurtID},
		{strawberryYogurtID, "Strawberry Yogurt", "Strawberry yogurt", 2.49, 40, yogurtID},

		// Bread
		{wholeWheatID, "Whole Wheat", "Whole wheat bread", 3.49, 40, breadID},
		{sourdoughID, "Sourdough", "Fresh sourdough loaf", 4.99, 30, breadID},
		{ryeBreadID, "Rye Bread", "Fresh rye bread", 4.49, 25, breadID},

		// Pastry
		{croissantsID, "Croissants", "Butter croissants, 4 pack", 6.99, 20, pastryID},
		{danishID, "Danish", "Assorted danish pastries", 5.99, 25, pastryID},
		{muffinsID, "Muffins", "Blueberry muffins, 4 pack", 5.99, 30, pastryID},
	}

	for _, p := range products {
		product := &domain.Product{
			ID:          p.id,
			Name:        p.name,
			Description: p.description,
			Price:       p.price,
			Stock:       p.stock,
			CategoryID:  p.categoryID,
		}

		if err := repo.Create(ctx, product); err != nil {
			return err
		}
	}

	return nil
}

func seedUsers(ctx context.Context, repo postgres.UserRepository) error {
	testUser := &domain.User{
		ID:        testUserID,
		Email:     "wanjohisteve04@gmail.com",
		Password:  "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password: test123
		Name:      "Test User",
		Phone:     "+254706595191",
		Address:   "123 Test Street, Test City, 12345",
		Picture:   "",
		Role:      domain.CustomerRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, testUser); err != nil {
		return err
	}

	return nil
}

func seedCustomers(ctx context.Context, repo postgres.CustomerRepository) error {
	testCustomer := &domain.Customer{
		ID:        testCustomerID,
		UserID:    testUserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, testCustomer); err != nil {
		return err
	}

	return nil
}
