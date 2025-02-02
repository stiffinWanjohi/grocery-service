package repository

import (
	"context"

	"github.com/grocery-service/internal/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context) ([]domain.Category, error)
	ListByParentID(ctx context.Context, parentID string) ([]domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id string) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	List(ctx context.Context) ([]domain.Product, error)
	ListByCategoryID(ctx context.Context, categoryID string) ([]domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, quantity int) error
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id string) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
	List(ctx context.Context) ([]domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, id string) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	ListByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	AddOrderItem(ctx context.Context, orderItem *domain.OrderItem) error
	RemoveOrderItem(ctx context.Context, orderID, orderItemID string) error
}

type Repository interface {
	Category() CategoryRepository
	Product() ProductRepository
	Customer() CustomerRepository
	Order() OrderRepository
}
