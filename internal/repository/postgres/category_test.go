package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCategoryTestDB(t *testing.T) *db.PostgresDB {
	postgres, err := db.NewTestDB()
	require.NoError(t, err)

	err = postgres.DB.Migrator().DropTable(&domain.Category{})
	require.NoError(t, err)
	err = postgres.DB.AutoMigrate(&domain.Category{})
	require.NoError(t, err)

	return postgres
}

func TestCategoryRepository_Create(t *testing.T) {
	postgres := setupCategoryTestDB(t)
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	parentID := uuid.New()
	category := &domain.Category{
		ID:          uuid.New(),
		Name:        "Test Category",
		Description: "Test Description",
		ParentID:    &parentID,
	}

	err := repo.Create(ctx, category)
	assert.NoError(t, err)

	var found domain.Category
	err = postgres.DB.First(&found, "id = ?", category.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, category.Name, found.Name)
	assert.Equal(t, category.Description, found.Description)
	assert.Equal(t, category.ParentID, found.ParentID)
}

func TestCategoryRepository_GetByID(t *testing.T) {
	postgres := setupCategoryTestDB(t)
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	parentID := uuid.New()
	category := &domain.Category{
		ID:          uuid.New(),
		Name:        "Test Category",
		Description: "Test Description",
		ParentID:    &parentID,
	}

	err := postgres.DB.Create(category).Error
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, category.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, category.ID, found.ID)
	assert.Equal(t, category.Name, found.Name)

	_, err = repo.GetByID(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrCategoryNotFound)
}

func TestCategoryRepository_List(t *testing.T) {
	postgres := setupCategoryTestDB(t)
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	categories := []domain.Category{
		{ID: uuid.New(), Name: "Category 1", Description: "Description 1"},
		{ID: uuid.New(), Name: "Category 2", Description: "Description 2"},
	}

	for _, c := range categories {
		err := postgres.DB.Create(&c).Error
		require.NoError(t, err)
	}

	found, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, len(categories))
}

func TestCategoryRepository_ListByParentID(t *testing.T) {
	postgres := setupCategoryTestDB(t)
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	parentID := uuid.New()
	parentID2 := uuid.New()
	categories := []domain.Category{
		{ID: uuid.New(), Name: "Category 1", ParentID: &parentID},
		{ID: uuid.New(), Name: "Category 2", ParentID: &parentID},
		{ID: uuid.New(), Name: "Category 3", ParentID: &parentID2},
	}

	for _, c := range categories {
		err := postgres.DB.Create(&c).Error
		require.NoError(t, err)
	}

	found, err := repo.ListByParentID(ctx, parentID.String())
	assert.NoError(t, err)
	assert.Len(t, found, 2)
	for _, c := range found {
		assert.Equal(t, parentID, c.ParentID)
	}
}

func TestCategoryRepository_Update(t *testing.T) {
	postgres := setupCategoryTestDB(t)
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	parentID := uuid.New()
	category := &domain.Category{
		ID:          uuid.New(),
		Name:        "Test Category",
		Description: "Test Description",
		ParentID:    &parentID,
	}

	err := postgres.DB.Create(category).Error
	require.NoError(t, err)

	category.Name = "Updated Category"
	err = repo.Update(ctx, category)
	assert.NoError(t, err)

	var found domain.Category
	err = postgres.DB.First(&found, "id = ?", category.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Category", found.Name)

	nonExistent := &domain.Category{ID: uuid.New()}
	err = repo.Update(ctx, nonExistent)
	assert.ErrorIs(t, err, customErrors.ErrCategoryNotFound)
}

func TestCategoryRepository_Delete(t *testing.T) {
	postgres := setupCategoryTestDB(t)
	repo := NewCategoryRepository(postgres)
	ctx := context.Background()

	parentID := uuid.New()
	category := &domain.Category{
		ID:          uuid.New(),
		Name:        "Test Category",
		Description: "Test Description",
		ParentID:    &parentID,
	}

	err := postgres.DB.Create(category).Error
	require.NoError(t, err)

	err = repo.Delete(ctx, category.ID.String())
	assert.NoError(t, err)

	err = postgres.DB.First(&domain.Category{}, "id = ?", category.ID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = repo.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, customErrors.ErrCategoryNotFound)
}
