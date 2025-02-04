package domain

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Phone     string    `json:"phone" gorm:"not null"`
	Address   string    `json:"address"`
	Orders    []*Order  `json:"orders,omitempty" gorm:"foreignKey:CustomerID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewCustomer(name, email, phone, address string) *Customer {
	return &Customer{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Phone:     phone,
		Address:   address,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *Customer) Validate() error {
	if c == nil {
		return fmt.Errorf("customer is required")
	}
	if c.Name == "" {
		return fmt.Errorf("customer name is required")
	}
	if len(c.Name) > 100 {
		return fmt.Errorf("customer name cannot exceed 100 characters")
	}
	if c.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(c.Email) {
		return fmt.Errorf("invalid email format")
	}
	if c.Phone != "" && !isValidPhone(c.Phone) {
		return fmt.Errorf("invalid phone format")
	}
	return nil
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func isValidPhone(phone string) bool {
	pattern := `^\+?[1-9]\d{1,14}$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}
