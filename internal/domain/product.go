package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"                 gorm:"type:uuid;primary_key"`
	Name        string    `json:"name"               gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price"              gorm:"not null"`
	Stock       int       `json:"stock"              gorm:"not null"`
	CategoryID  uuid.UUID `json:"category_id"        gorm:"type:uuid;not null"`
	Category    *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewProduct(
	name, description string,
	price float64,
	stock int,
	categoryID uuid.UUID,
) *Product {
	return &Product{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		CategoryID:  categoryID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if p.Price <= 0 {
		return fmt.Errorf("product price must be greater than zero")
	}
	if p.Stock < 0 {
		return fmt.Errorf("product stock cannot be negative")
	}
	return nil
}
