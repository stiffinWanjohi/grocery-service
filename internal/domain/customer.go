package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;unique"`
	Orders    []*Order  `json:"orders,omitempty" gorm:"foreignKey:CustomerID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user" gorm:"foreignKey:UserID"`
}

func NewCustomer(userID uuid.UUID) *Customer {
	return &Customer{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *Customer) Validate() error {
	if c == nil {
		return fmt.Errorf("customer is required")
	}
	if c.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	return nil
}
