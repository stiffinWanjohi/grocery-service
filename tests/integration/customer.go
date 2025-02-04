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

func setupCustomerTest() (*serviceMock.CustomerService, *handler.CustomerHandler) {
	mockService := new(serviceMock.CustomerService)
	handler := handler.NewCustomerHandler(mockService)
	return mockService, handler
}

func TestCustomerHandler_Create(t *testing.T) {
	mockService, handler := setupCustomerTest()

	tests := []struct {
		name       string
		customer   *domain.Customer
		setupMock  func(*domain.Customer)
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			customer: &domain.Customer{
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
			setupMock: func(c *domain.Customer) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(cust *domain.Customer) bool {
					return cust.Name == c.Name && cust.Email == c.Email
				})).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid Data",
			customer: &domain.Customer{
				Name:  "",
				Email: "invalid-email",
			},
			setupMock: func(c *domain.Customer) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(cust *domain.Customer) bool {
					return cust.Name == c.Name && cust.Email == c.Email
				})).Return(customErrors.ErrInvalidCustomerData)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid customer data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(tt.customer)

			jsonBody, _ := json.Marshal(tt.customer)
			req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
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
				var returnedCustomer domain.Customer
				customerData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(customerData, &returnedCustomer)
				assert.NoError(t, err)
				assert.Equal(t, tt.customer.Name, returnedCustomer.Name)
				assert.Equal(t, tt.customer.Email, returnedCustomer.Email)
			}
		})
	}
}

func TestCustomerHandler_GetByID(t *testing.T) {
	mockService, handler := setupCustomerTest()

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
				mockService.On("GetByID", mock.Anything, testID.String()).Return(&domain.Customer{
					ID:    testID,
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid customer ID",
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("GetByID", mock.Anything, testID.String()).Return(nil, customErrors.ErrCustomerNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Customer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(http.MethodGet, "/customers/"+tt.id, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
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

func TestCustomerHandler_List(t *testing.T) {
	mockService, handler := setupCustomerTest()

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
				mockService.On("List", mock.Anything).Return([]domain.Customer{
					{
						ID:    uuid.New(),
						Name:  "John Doe",
						Email: "john.doe@example.com",
					},
					{
						ID:    uuid.New(),
						Name:  "Jane Smith",
						Email: "jane.smith@example.com",
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "Internal Error",
			setupMock: func() {
				mockService.On("List", mock.Anything).Return(nil, fmt.Errorf("database error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "Failed to list customers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/customers", nil)
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
				var customers []domain.Customer
				customersData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(customersData, &customers)
				assert.NoError(t, err)
				assert.Len(t, customers, tt.wantCount)
			}
		})
	}
}

func TestCustomerHandler_Update(t *testing.T) {
	mockService, handler := setupCustomerTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		customer   *domain.Customer
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			id:   testID.String(),
			customer: &domain.Customer{
				ID:    testID,
				Name:  "Updated John",
				Email: "updated.john@example.com",
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   testID.String(),
			customer: &domain.Customer{
				ID:    testID,
				Name:  "Updated John",
				Email: "updated.john@example.com",
			},
			setupMock: func() {
				mockService.On("Update", mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(customErrors.ErrCustomerNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Customer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			jsonBody, _ := json.Marshal(tt.customer)
			req := httptest.NewRequest(http.MethodPut, "/customers/"+tt.id, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

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

func TestCustomerHandler_Delete(t *testing.T) {
	mockService, handler := setupCustomerTest()

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
				mockService.On("Delete", mock.Anything, testID.String()).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("Delete", mock.Anything, testID.String()).Return(customErrors.ErrCustomerNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Customer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(http.MethodDelete, "/customers/"+tt.id, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
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
