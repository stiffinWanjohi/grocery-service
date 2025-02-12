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

func setupProductTest() (
	*serviceMock.ProductService,
	*handler.ProductHandler,
) {
	mockService := new(serviceMock.ProductService)
	handler := handler.NewProductHandler(mockService)
	return mockService, handler
}

func TestProductHandler_Create(t *testing.T) {
	mockService, handler := setupProductTest()

	categoryID := uuid.New()
	tests := []struct {
		name       string
		product    *domain.Product
		setupMock  func(*domain.Product)
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			product: &domain.Product{
				Name:        "Apple",
				Description: "Fresh red apple",
				Price:       1.99,
				CategoryID:  categoryID,
				Stock:       100,
			},
			setupMock: func(p *domain.Product) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(product *domain.Product) bool {
					return product.Name == p.Name &&
						product.Price == p.Price
				})).
					Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid Product",
			product: &domain.Product{
				Name:  "",
				Price: -1,
			},
			setupMock: func(p *domain.Product) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(product *domain.Product) bool {
					return product.Name == p.Name &&
						product.Price == p.Price
				})).
					Return(customErrors.ErrInvalidProductData)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid product data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(tt.product)

			jsonBody, _ := json.Marshal(tt.product)
			req := httptest.NewRequest(
				http.MethodPost,
				"/products",
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
				var returnedProduct domain.Product

				productData, err := json.Marshal(response.Data)
				assert.NoError(t, err)

				err = json.Unmarshal(productData, &returnedProduct)
				assert.NoError(t, err)
				assert.Equal(t, tt.product.Name, returnedProduct.Name)
				assert.Equal(t, tt.product.Price, returnedProduct.Price)
			}
		})
	}
}

func TestProductHandler_GetByID(t *testing.T) {
	mockService, handler := setupProductTest()

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
					Return(&domain.Product{
						ID:    testID,
						Name:  "Banana",
						Price: 0.99,
						Stock: 50,
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid product ID",
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("GetByID", mock.Anything, testID.String()).
					Return(nil, customErrors.ErrProductNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.id, nil)
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
				assert.Equal(t, tt.wantError, response.Error)
			} else {
				assert.True(t, response.Success)
				var returnedProduct domain.Product
				productData, err := json.Marshal(response.Data)
				assert.NoError(t, err)

				err = json.Unmarshal(productData, &returnedProduct)
				assert.NoError(t, err)
				assert.Equal(t, testID, returnedProduct.ID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_List(t *testing.T) {
	mockService, handler := setupProductTest()

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
					Return([]domain.Product{
						{
							ID:    uuid.New(),
							Name:  "Apple",
							Price: 1.99,
						},
						{
							ID:    uuid.New(),
							Name:  "Banana",
							Price: 0.99,
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
					Return(nil, fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "Failed to list products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Equal(t, tt.wantError, response.Error)
			} else {
				assert.True(t, response.Success)
				var products []domain.Product
				productsData, err := json.Marshal(response.Data)
				assert.NoError(t, err)

				err = json.Unmarshal(productsData, &products)
				assert.NoError(t, err)
				assert.Len(t, products, tt.wantCount)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_Update(t *testing.T) {
	mockService, handler := setupProductTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		product    *domain.Product
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			id:   testID.String(),
			product: &domain.Product{
				ID:    testID,
				Name:  "Updated Apple",
				Price: 2.49,
				Stock: 75,
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   testID.String(),
			product: &domain.Product{
				ID:    testID,
				Name:  "Updated Apple",
				Price: 2.49,
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
					Return(customErrors.ErrProductNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Product not found",
		},
		{
			name: "Invalid Data",
			id:   testID.String(),
			product: &domain.Product{
				ID:    testID,
				Name:  "",
				Price: -1,
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Product")).
					Return(customErrors.ErrInvalidProductData)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid product data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			tt.setupMock()

			jsonBody, _ := json.Marshal(tt.product)
			req := httptest.NewRequest(
				http.MethodPut,
				"/products/"+tt.id,
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
				assert.Equal(t, tt.wantError, response.Error)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_Delete(t *testing.T) {
	mockService, handler := setupProductTest()

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
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid product ID",
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("Delete", mock.Anything, testID.String()).
					Return(customErrors.ErrProductNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(
				http.MethodDelete,
				"/products/"+tt.id,
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
				assert.Equal(t, tt.wantError, response.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_ListByCategoryID(t *testing.T) {
	mockService, handler := setupProductTest()

	categoryID := uuid.New()
	tests := []struct {
		name       string
		categoryID string
		setupMock  func()
		wantStatus int
		wantCount  int
		wantError  string
	}{
		{
			name:       "Success",
			categoryID: categoryID.String(),
			setupMock: func() {
				mockService.On("ListByCategoryID", mock.Anything, categoryID.String()).
					Return([]domain.Product{
						{
							ID:         uuid.New(),
							Name:       "Red Apple",
							CategoryID: categoryID,
						},
						{
							ID:         uuid.New(),
							Name:       "Green Apple",
							CategoryID: categoryID,
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "Invalid Category ID",
			categoryID: "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid category ID",
		},
		{
			name:       "Category Not Found",
			categoryID: categoryID.String(),
			setupMock: func() {
				mockService.On("ListByCategoryID", mock.Anything, categoryID.String()).
					Return(nil, customErrors.ErrCategoryNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("categoryID", tt.categoryID)
			req := httptest.NewRequest(
				http.MethodGet,
				"/products/category/"+tt.categoryID,
				nil,
			)
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)
			w := httptest.NewRecorder()

			handler.ListByCategoryID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Equal(t, tt.wantError, response.Error)
			} else {
				assert.True(t, response.Success)
				var products []domain.Product
				productsData, err := json.Marshal(response.Data)
				assert.NoError(t, err)

				err = json.Unmarshal(productsData, &products)
				assert.NoError(t, err)
				assert.Len(t, products, tt.wantCount)

				if tt.wantCount > 0 {
					assert.Equal(t, categoryID, products[0].CategoryID)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestProductHandler_UpdateStock(t *testing.T) {
	mockService, handler := setupProductTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		quantity   int
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name:     "Success",
			id:       testID.String(),
			quantity: 50,
			setupMock: func() {
				mockService.On("UpdateStock", mock.Anything, testID.String(), 50).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			quantity:   50,
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid product ID",
		},
		{
			name:     "Invalid Quantity",
			id:       testID.String(),
			quantity: -1,
			setupMock: func() {
				mockService.On("UpdateStock", mock.Anything, testID.String(), -1).
					Return(customErrors.ErrInvalidProductData)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid product data",
		},
		{
			name:     "Product Not Found",
			id:       testID.String(),
			quantity: 50,
			setupMock: func() {
				mockService.On("UpdateStock",
					mock.Anything,
					testID.String(),
					50,
				).Return(customErrors.ErrProductNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			tt.setupMock()

			stockUpdate := struct {
				Quantity int `json:"quantity"`
			}{
				Quantity: tt.quantity,
			}

			jsonBody, _ := json.Marshal(stockUpdate)
			req := httptest.NewRequest(
				http.MethodPut,
				"/products/"+tt.id+"/stock",
				bytes.NewBuffer(jsonBody),
			)
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)

			w := httptest.NewRecorder()

			handler.UpdateStock(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Equal(t, tt.wantError, response.Error)
			} else {
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}
