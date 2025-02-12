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

	categoryWithoutParent := &domain.Category{
		ID:   uuid.New(),
		Name: "Another Category",
	}
	mockRepo.On("Create", ctx, categoryWithoutParent).Return(nil)

	err := service.Create(ctx, categoryWithoutParent)
	assert.NoError(t, err)

	mockRepo.On("GetByID", ctx, parentID.String()).
		Return(&domain.Category{ID: parentID}, nil)
	mockRepo.On("Create", ctx, category).Return(nil)

	err = service.Create(ctx, category)
	assert.NoError(t, err)

	invalidParentID := uuid.New()
	invalidParentCategory := &domain.Category{
		ID:       uuid.New(),
		Name:     "Invalid Parent Category",
		ParentID: &invalidParentID,
	}
	mockRepo.On("GetByID", ctx, invalidParentID.String()).
		Return(nil, customErrors.ErrCategoryNotFound)

	err = service.Create(ctx, invalidParentCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent category")

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

	_, err := service.GetByID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category ID is required")

	mockRepo.On("GetByID", ctx, category.ID.String()).Return(category, nil)
	found, err := service.GetByID(ctx, category.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, category.ID, found.ID)

	mockRepo.On("GetByID", ctx, "non-existent").
		Return(nil, customErrors.ErrCategoryNotFound)
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

	_, err := service.ListByParentID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent ID is required")

	categories := []domain.Category{
		{
			ID:       uuid.New(),
			Name:     "Category 1",
			ParentID: &parentID,
		},
		{
			ID:       uuid.New(),
			Name:     "Category 2",
			ParentID: &parentID,
		},
	}

	mockRepo.On("ListByParentID", ctx, parentID.String()).
		Return(categories, nil)

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

	categoryWithoutParent := &domain.Category{
		ID:   uuid.New(),
		Name: "Another Category",
	}
	mockRepo.On("Update", ctx, categoryWithoutParent).Return(nil)
	err := service.Update(ctx, categoryWithoutParent)
	assert.NoError(t, err)

	mockRepo.On("GetByID", ctx, parentID.String()).
		Return(&domain.Category{ID: parentID}, nil)
	mockRepo.On("Update", ctx, category).Return(nil)
	err = service.Update(ctx, category)
	assert.NoError(t, err)

	selfParentCategory := &domain.Category{
		ID:   uuid.New(),
		Name: "Self Parent Category",
	}
	selfParentID := selfParentCategory.ID
	selfParentCategory.ParentID = &selfParentID

	err = service.Update(ctx, selfParentCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category cannot be its own parent")

	invalidParentID := uuid.New()
	invalidParentCategory := &domain.Category{
		ID:       uuid.New(),
		Name:     "Invalid Parent Category",
		ParentID: &invalidParentID,
	}

	mockRepo.On("GetByID", ctx, invalidParentID.String()).
		Return(nil, customErrors.ErrCategoryNotFound)
	err = service.Update(ctx, invalidParentCategory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent category")

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

	t.Run(
		"Empty ID",
		func(t *testing.T) {
			err := service.Delete(ctx, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "category ID is required")
		})

	t.Run(
		"Category with subcategories",
		func(t *testing.T) {
			id := uuid.New().String()
			parsedID, _ := uuid.Parse(id)
			subcategories := []domain.Category{
				{
					ID:       uuid.New(),
					Name:     "Subcategory",
					ParentID: &parsedID,
				},
			}
			mockRepo.On("ListByParentID", ctx, id).
				Return(subcategories, nil).
				Once()
			err := service.Delete(ctx, id)
			assert.Error(t, err)
			assert.Contains(
				t,
				err.Error(),
				"cannot delete category with subcategories",
			)
			mockRepo.AssertExpectations(t)
		})

	t.Run(
		"Successful deletion",
		func(t *testing.T) {
			id := uuid.New().String()
			mockRepo.On("ListByParentID", ctx, id).
				Return([]domain.Category{}, nil).
				Once()
			mockRepo.On("Delete", ctx, id).Return(nil).Once()
			err := service.Delete(ctx, id)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})

	t.Run(
		"Repository error",
		func(t *testing.T) {
			id := uuid.New().String()
			mockRepo.On("ListByParentID", ctx, id).
				Return([]domain.Category{}, nil).
				Once()
			mockRepo.On("Delete", ctx, id).
				Return(customErrors.ErrDBQuery).
				Once()
			err := service.Delete(ctx, id)
			assert.Error(t, err)
			assert.ErrorIs(t, err, customErrors.ErrDBQuery)
			mockRepo.AssertExpectations(t)
		})
}
