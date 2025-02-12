package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	handler "github.com/grocery-service/internal/api/handlers"
	"github.com/grocery-service/internal/domain"
	serviceMock "github.com/grocery-service/tests/mocks/service"
	"github.com/grocery-service/utils/api"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCategoryTest() (
	*serviceMock.CategoryService,
	*handler.CategoryHandler,
) {
	mockService := new(serviceMock.CategoryService)
	handler := handler.NewCategoryHandler(mockService)
	return mockService, handler
}

func TestCategoryHandler_Create(t *testing.T) {
	mockService, handler := setupCategoryTest()

	tests := []struct {
		name       string
		category   *domain.Category
		setupMock  func(*domain.Category)
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			category: &domain.Category{
				Name:        "Fruits",
				Description: "Fresh fruits category",
			},
			setupMock: func(c *domain.Category) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(cat *domain.Category) bool {
					return cat.Name == c.Name &&
						cat.Description == c.Description
				})).
					Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid Data",
			category: &domain.Category{
				Name: "", // Invalid: empty name
			},
			setupMock: func(c *domain.Category) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(cat *domain.Category) bool {
					return cat.Name == c.Name
				})).
					Return(customErrors.ErrInvalidCategoryData).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid category data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(tt.category)

			jsonBody, _ := json.Marshal(tt.category)
			req := httptest.NewRequest(
				http.MethodPost,
				"/categories",
				bytes.NewBuffer(jsonBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
				var returnedCategory domain.Category
				categoryData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(categoryData, &returnedCategory)
				assert.NoError(t, err)
				assert.Equal(t, tt.category.Name, returnedCategory.Name)
				assert.Equal(t, tt.category.Description, returnedCategory.Description)
			}
		})
	}
}

func TestCategoryHandler_GetByID(t *testing.T) {
	mockService, handler := setupCategoryTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("GetByID", mock.Anything, testID.String()).
					Return(&domain.Category{
						ID:          testID,
						Name:        "Vegetables",
						Description: "Fresh vegetables category",
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid category ID",
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("GetByID", mock.Anything, testID.String()).
					Return(nil, customErrors.ErrCategoryNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(
				http.MethodGet,
				"/categories/"+tt.id,
				nil,
			)
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)
			w := httptest.NewRecorder()

			handler.GetByID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
			}
		})
	}
}

func TestCategoryHandler_List(t *testing.T) {
	mockService, handler := setupCategoryTest()

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
		wantCount  int
		wantError  string
	}{
		{
			name: "Success",
			setupMock: func() {
				mockService.On("List", mock.Anything).
					Return([]domain.Category{
						{
							ID:          uuid.New(),
							Name:        "Fruits",
							Description: "Fresh fruits",
						},
						{
							ID:          uuid.New(),
							Name:        "Vegetables",
							Description: "Fresh vegetables",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "Internal Error",
			setupMock: func() {
				mockService.On("List", mock.Anything).
					Return(nil, fmt.Errorf("database error")).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "Failed to list categories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/categories", nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
				var categories []domain.Category
				categoriesData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(categoriesData, &categories)
				assert.NoError(t, err)
				assert.Len(t, categories, tt.wantCount)
			}
		})
	}
}

func TestCategoryHandler_Update(t *testing.T) {
	mockService, handler := setupCategoryTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		category   *domain.Category
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			id:   testID.String(),
			category: &domain.Category{
				ID:          testID,
				Name:        "Updated Fruits",
				Description: "Updated fruits category",
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Category")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   testID.String(),
			category: &domain.Category{
				ID:          testID,
				Name:        "Updated Fruits",
				Description: "Updated fruits category",
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Category")).
					Return(customErrors.ErrCategoryNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMock()

			jsonBody, _ := json.Marshal(tt.category)
			req := httptest.NewRequest(
				http.MethodPut,
				"/categories/"+tt.id,
				bytes.NewBuffer(jsonBody),
			)

			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)

			w := httptest.NewRecorder()

			handler.Update(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
			}
		})
	}
}

func TestCategoryHandler_Delete(t *testing.T) {
	mockService, handler := setupCategoryTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("Delete", mock.Anything, testID.String()).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("Delete", mock.Anything, testID.String()).
					Return(customErrors.ErrCategoryNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(
				http.MethodDelete,
				"/categories/"+tt.id,
				nil,
			)
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)
			w := httptest.NewRecorder()

			handler.Delete(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response api.Response
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			}
		})
	}
}

func TestCategoryHandler_ListByParentID(t *testing.T) {
	mockService, handler := setupCategoryTest()

	parentID := uuid.New()
	tests := []struct {
		name       string
		id         string
		setupMock  func()
		wantStatus int
		wantCount  int
		wantError  string
	}{
		{
			name: "Success",
			id:   parentID.String(),
			setupMock: func() {
				mockService.On("ListByParentID", mock.Anything, parentID.String()).
					Return([]domain.Category{
						{
							ID:          uuid.New(),
							Name:        "Green Fruits",
							Description: "All green fruits",
							ParentID:    &parentID,
						},
						{
							ID:          uuid.New(),
							Name:        "Red Fruits",
							Description: "All red fruits",
							ParentID:    &parentID,
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid parent category ID",
		},
		{
			name: "Not Found",
			id:   parentID.String(),
			setupMock: func() {
				mockService.On("ListByParentID", mock.Anything, parentID.String()).
					Return(nil, customErrors.ErrCategoryNotFound).
					Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Parent category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(
				http.MethodGet,
				"/categories/"+tt.id+"/subcategories",
				nil,
			)
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)
			w := httptest.NewRecorder()

			handler.ListByParentID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
				var categories []domain.Category
				categoriesData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(categoriesData, &categories)
				assert.NoError(t, err)
				assert.Len(t, categories, tt.wantCount)
				for _, category := range categories {
					assert.Equal(t, parentID, *category.ParentID)
				}
			}
		})
	}
}
