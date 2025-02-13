package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCategoryRepository_Create(t *testing.T) {
	postgres := setupTestDB(t, &domain.Category{})
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	// Create parent categories for testing
	parentID1 := uuid.New()
	parentCategory1 := &domain.Category{
		ID:          parentID1,
		Name:        "Parent Category 1",
		Description: "Parent Description 1",
		Level:       0,
		Path:        "Parent Category 1",
	}
	require.NoError(t, postgres.DB.Create(parentCategory1).Error)

	parentID2 := uuid.New()
	parentCategory2 := &domain.Category{
		ID:          parentID2,
		Name:        "Parent Category 2",
		Description: "Parent Description 2",
		Level:       0,
		Path:        "Parent Category 2",
	}
	require.NoError(t, postgres.DB.Create(parentCategory2).Error)

	tests := []struct {
		name       string
		category   *domain.Category
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success with parent ID - level 1",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Test Category L1",
				Description: "Test Description L1",
				ParentID:    &parentID1,
				Level:       1,
				Path:        "Parent Category 1/Test Category L1",
			},
			wantErr: false,
		},
		{
			name: "success with parent ID - level 2",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Test Category L2",
				Description: "Test Description L2",
				ParentID:    &parentID2,
				Level:       1,
				Path:        "Parent Category 2/Test Category L2",
			},
			wantErr: false,
		},
		{
			name: "success without parent ID - root category",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Root Category",
				Description: "Root Description",
				Level:       0,
				Path:        "Root Category",
			},
			wantErr: false,
		},
		{
			name: "failure - empty name",
			category: &domain.Category{
				ID:          uuid.New(),
				Description: "Test Description",
				Level:       0,
			},
			wantErr:    true,
			wantErrMsg: "invalid category data",
		},
		{
			name: "failure - non-existent parent ID",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Invalid Parent",
				Description: "Should Fail",
				ParentID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Level:       1,
				Path:        "Invalid Parent",
			},
			wantErr:    true,
			wantErrMsg: "parent category not found",
		},
		{
			name: "failure - invalid level for root category",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Invalid Level Root",
				Description: "Should Fail - Root with Level 1",
				Level:       1,
				Path:        "Invalid Level Root",
				ParentID:    nil,
			},
			wantErr:    true,
			wantErrMsg: "invalid category level",
		},
		{
			name: "failure - invalid path format",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Invalid Path",
				Description: "Should Fail",
				ParentID:    &parentID1,
				Level:       1,
				Path:        "/Wrong/Path/Format/",
			},
			wantErr:    true,
			wantErrMsg: customErrors.ErrInvalidCategoryData.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.category)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.ErrorIs(t, err, customErrors.ErrInvalidCategoryData)
				}
				return
			}

			assert.NoError(t, err)

			var found domain.Category
			err = postgres.DB.First(&found, "id = ?", tt.category.ID).Error
			assert.NoError(t, err)
			assert.Equal(t, tt.category.Name, found.Name)
			assert.Equal(t, tt.category.Description, found.Description)
			assert.Equal(t, tt.category.ParentID, found.ParentID)
			assert.Equal(t, tt.category.Level, found.Level)
			assert.Equal(t, tt.category.Path, found.Path)

			// Verify timestamps are set
			assert.False(t, found.CreatedAt.IsZero())
			assert.False(t, found.UpdatedAt.IsZero())

			// Verify parent exists if ParentID is set
			if tt.category.ParentID != nil {
				var parent domain.Category
				err = postgres.DB.First(
					&parent,
					"id = ?",
					tt.category.ParentID,
				).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.category.Level, parent.Level+1)
				assert.True(t, len(found.Path) > len(parent.Path))
			}
		})
	}
}

func TestCategoryRepository_GetByID(t *testing.T) {
	postgres := setupTestDB(t, &domain.Category{})
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   error
	}{
		{
			name: "success",
			setupFunc: func() string {
				category := &domain.Category{
					ID:          uuid.New(),
					Name:        "Test Category",
					Description: "Test Description",
				}
				require.NoError(t, postgres.DB.Create(category).Error)
				return category.ID.String()
			},
			wantErr: nil,
		},
		{
			name: "not found",
			setupFunc: func() string {
				return uuid.New().String()
			},
			wantErr: customErrors.ErrCategoryNotFound,
		},
		{
			name: "invalid uuid",
			setupFunc: func() string {
				return "invalid-uuid"
			},
			wantErr: customErrors.ErrInvalidCategoryData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setupFunc()
			category, err := repo.GetByID(ctx, id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, category)
				if tt.wantErr == customErrors.ErrInvalidCategoryData {
					// Don't proceed with database checks for invalid UUID
					return
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, category)
			assert.Equal(t, id, category.ID.String())
		})
	}
}

func TestCategoryRepository_List(t *testing.T) {
	postgres := setupTestDB(t, &domain.Category{})
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	// Create test categories
	categories := []domain.Category{
		{
			ID:          uuid.New(),
			Name:        "Category 1",
			Description: "Description 1",
		},
		{
			ID:          uuid.New(),
			Name:        "Category 2",
			Description: "Description 2",
		},
		{
			ID:          uuid.New(),
			Name:        "Category 3",
			Description: "Description 3",
		},
	}

	for _, c := range categories {
		require.NoError(t, postgres.DB.Create(&c).Error)
	}

	tests := []struct {
		name          string
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "success with multiple categories",
			expectedCount: 3,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.List(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
		})
	}
}

func TestCategoryRepository_Update(t *testing.T) {
	postgres := setupTestDB(t, &domain.Category{})
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	existingCategory := &domain.Category{
		ID:          uuid.New(),
		Name:        "Original Name",
		Description: "Original Description",
	}

	require.NoError(t, postgres.DB.Create(existingCategory).Error)

	tests := []struct {
		name     string
		category *domain.Category
		wantErr  error
	}{
		{
			name: "success",
			category: &domain.Category{
				ID:          existingCategory.ID,
				Name:        "Updated Name",
				Description: "Updated Description",
			},
			wantErr: nil,
		},
		{
			name: "not found",
			category: &domain.Category{
				ID:          uuid.New(),
				Name:        "Non-existent",
				Description: "Should not update",
			},
			wantErr: customErrors.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.category)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)

			var updated domain.Category
			err = postgres.DB.First(&updated, "id = ?", tt.category.ID).Error
			assert.NoError(t, err)
			assert.Equal(t, tt.category.Name, updated.Name)
			assert.Equal(t, tt.category.Description, updated.Description)
		})
	}
}

func TestCategoryRepository_Delete(t *testing.T) {
	postgres := setupTestDB(t, &domain.Category{})
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	existingCategory := &domain.Category{
		ID:          uuid.New(),
		Name:        "To Delete",
		Description: "Will be deleted",
	}

	require.NoError(t, postgres.DB.Create(existingCategory).Error)

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "success",
			id:      existingCategory.ID.String(),
			wantErr: nil,
		},
		{
			name:    "not found",
			id:      uuid.New().String(),
			wantErr: customErrors.ErrCategoryNotFound,
		},
		{
			name:    "invalid uuid",
			id:      "invalid-uuid",
			wantErr: customErrors.ErrInvalidCategoryData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				// Skip database check for invalid cases
				return
			}

			assert.NoError(t, err)

			// Only verify deletion for successful cases
			var found domain.Category
			err = postgres.DB.First(&found, "id = ?", tt.id).Error
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		})
	}
}
