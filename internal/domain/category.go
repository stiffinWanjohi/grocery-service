package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID            uuid.UUID   `json:"id"                       gorm:"type:uuid;primary_key"`
	Name          string      `json:"name"                     gorm:"not null"`
	Description   string      `json:"description"`
	ParentID      *uuid.UUID  `json:"parent_id,omitempty"      gorm:"type:uuid"`
	SubCategories []*Category `json:"sub_categories,omitempty" gorm:"foreignKey:ParentID"`
	Level         int         `json:"level"                    gorm:"not null"`
	Path          string      `json:"path"                     gorm:"not null"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func NewCategory(
	name string,
	description string,
	parentID *uuid.UUID,
) *Category {
	return &Category{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		ParentID:    parentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (c *Category) Validate() error {
	if c == nil {
		return fmt.Errorf("category is required")
	}
	if c.Name == "" {
		return fmt.Errorf("category name is required")
	}
	if len(c.Name) > 100 {
		return fmt.Errorf("category name cannot exceed 100 characters")
	}
	if c.Description != "" && len(c.Description) > 500 {
		return fmt.Errorf("category description cannot exceed 500 characters")
	}
	return nil
}

func (c *Category) ValidatePathFormat(path string) bool {
	return path != "" && path[0] != '/' && path[len(path)-1] != '/'
}
