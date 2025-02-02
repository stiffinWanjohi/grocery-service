package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	Name      string     `json:"name" gorm:"not null"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	Level     int        `json:"level" gorm:"not null"`
	Path      string     `json:"path" gorm:"not null"`
	Products  []Product  `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func NewCategory(name string, parentID *uuid.UUID) *Category {
	return &Category{
		ID:        uuid.New(),
		Name:      name,
		ParentID:  parentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}