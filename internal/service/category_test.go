package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	mocks "github.com/grocery-service/tests/mocks/repository"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
)

func TestCategoryService_Create(t *testing.T) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	parentID := uuid.New()
	category := &domain.Category{
		ID:       uuid.New(),
		Name:     "Test Category",
		ParentID: &parentID,
	}

	// Test successful creation without parent
	categoryWithoutParent := &domain.Category{
		ID:   uuid.New(),
		Name: "Another Category",
	}
	mockRepo.On("Create", ctx, categoryWithoutParent).Return(nil)

	err := service.Create(ctx, categoryWithoutParent)
	assert.NoError(t, err)

	// Test creation with valid parent
	mockRepo.On("GetByID", ctx, parentID.String()).Return(&domain.Category{ID: parentID}, nil)
	mockRepo.On("Create", ctx, category).Return(nil)

	err = service.Create(ctx, category)
	assert.NoError(t, err)

	// Test invalid parent
	invalidParentID := uuid.New()
	invalidParentCategory := &domain.Category{
		ID:       uuid.New(),
		Name:     "Invalid Parent Category",
		ParentID: &invalidParentID,
	}
	mockRepo.On("GetByID", ctx, invalidParentID.String()).Return(nil, customErrors.ErrCategoryNotFound)

	err = service.Create(ctx, invalidParentCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent category")

	// Test validation error
	invalidCategory := &domain.Category{
		ID: uuid.New(),
	}
	err = service.Create(ctx, invalidCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category name is required")
}

func TestCategoryService_GetByID(t *testing.T) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	category := &domain.Category{
		ID:   uuid.New(),
		Name: "Test Category",
	}

	// Test empty ID
	_, err := service.GetByID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category ID is required")

	// Test successful retrieval
	mockRepo.On("GetByID", ctx, category.ID.String()).Return(category, nil)
	found, err := service.GetByID(ctx, category.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, category.ID, found.ID)

	// Test not found case
	mockRepo.On("GetByID", ctx, "non-existent").Return(nil, customErrors.ErrCategoryNotFound)
	_, err = service.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, customErrors.ErrCategoryNotFound)
}

func TestCategoryService_List(t *testing.T) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	categories := []domain.Category{
		{ID: uuid.New(), Name: "Category 1"},
		{ID: uuid.New(), Name: "Category 2"},
	}

	mockRepo.On("List", ctx).Return(categories, nil)

	found, err := service.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, len(categories))
}

func TestCategoryService_ListByParentID(t *testing.T) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	parentID := uuid.New()

	// Test empty parent ID
	_, err := service.ListByParentID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent ID is required")

	// Test successful retrieval
	categories := []domain.Category{
		{ID: uuid.New(), Name: "Category 1", ParentID: &parentID},
		{ID: uuid.New(), Name: "Category 2", ParentID: &parentID},
	}

	mockRepo.On("ListByParentID", ctx, parentID.String()).Return(categories, nil)

	found, err := service.ListByParentID(ctx, parentID.String())
	assert.NoError(t, err)
	assert.Len(t, found, len(categories))
}

func TestCategoryService_Update(t *testing.T) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	parentID := uuid.New()
	category := &domain.Category{
		ID:       uuid.New(),
		Name:     "Test Category",
		ParentID: &parentID,
	}

	// Test successful update without parent
	categoryWithoutParent := &domain.Category{
		ID:   uuid.New(),
		Name: "Another Category",
	}
	mockRepo.On("Update", ctx, categoryWithoutParent).Return(nil)

	err := service.Update(ctx, categoryWithoutParent)
	assert.NoError(t, err)

	// Test update with valid parent
	mockRepo.On("GetByID", ctx, parentID.String()).Return(&domain.Category{ID: parentID}, nil)
	mockRepo.On("Update", ctx, category).Return(nil)

	err = service.Update(ctx, category)
	assert.NoError(t, err)

	// Test category cannot be its own parent
	selfParentCategory := &domain.Category{
		ID:   uuid.New(),
		Name: "Self Parent Category",
	}
	selfParentID := selfParentCategory.ID
	selfParentCategory.ParentID = &selfParentID

	err = service.Update(ctx, selfParentCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category cannot be its own parent")

	// Test invalid parent
	invalidParentID := uuid.New()
	invalidParentCategory := &domain.Category{
		ID:       uuid.New(),
		Name:     "Invalid Parent Category",
		ParentID: &invalidParentID,
	}
	mockRepo.On("GetByID", ctx, invalidParentID.String()).Return(nil, customErrors.ErrCategoryNotFound)

	err = service.Update(ctx, invalidParentCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent category")

	// Test validation error
	invalidCategory := &domain.Category{
		ID: uuid.New(),
	}
	err = service.Update(ctx, invalidCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category name is required")
}

func TestCategoryService_Delete(t *testing.T) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	id := uuid.New().String()

	// Test empty ID
	err := service.Delete(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category ID is required")

	// Test category with subcategories
	subcategories := []domain.Category{
		{ID: uuid.New(), Name: "Subcategory", ParentID: func() *uuid.UUID {
			parsedID, _ := uuid.Parse(id)
			return &parsedID
		}()},
	}
	mockRepo.On("ListByParentID", ctx, id).Return(subcategories, nil)

	err = service.Delete(ctx, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete category with subcategories")

	// Test successful deletion
	mockRepo.On("ListByParentID", ctx, id).Return([]domain.Category{}, nil)
	mockRepo.On("Delete", ctx, id).Return(nil)

	err = service.Delete(ctx, id)
	assert.NoError(t, err)
}
