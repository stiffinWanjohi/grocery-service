package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	handler "github.com/grocery-service/internal/api/handlers"
	"github.com/grocery-service/internal/api/middleware"
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
	userID := uuid.New().String()

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			setupMock: func() {
				customer := &domain.Customer{
					User: &domain.User{
						ID: uuid.MustParse(userID),
					},
				}
				mockService.On("Create", mock.Anything, userID).Return(customer, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Customer Already Exists",
			setupMock: func() {
				mockService.On("Create", mock.Anything, userID).Return(nil, customErrors.ErrCustomerExists)
			},
			wantStatus: http.StatusConflict,
			wantError:  "Customer profile already exists for this user",
		},
		{
			name: "Internal Error",
			setupMock: func() {
				mockService.On("Create", mock.Anything, userID).Return(nil, customErrors.ErrInternalServer)
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "Failed to create customer profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/customers", nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
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
			}
		})
	}
}

func TestCustomerHandler_GetCurrentCustomer(t *testing.T) {
	mockService, handler := setupCustomerTest()
	userID := uuid.New().String()

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			setupMock: func() {
				customer := &domain.Customer{
					User: &domain.User{
						ID: uuid.MustParse(userID),
					},
				}
				mockService.On("GetByUserID", mock.Anything, userID).Return(customer, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			setupMock: func() {
				mockService.On("GetByUserID", mock.Anything, userID).Return(nil, customErrors.ErrCustomerNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Customer profile not found",
		},
		{
			name: "Internal Error",
			setupMock: func() {
				mockService.On("GetByUserID", mock.Anything, userID).Return(nil, customErrors.ErrInternalServer)
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "Failed to get customer profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/customers/me", nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
			w := httptest.NewRecorder()

			handler.GetCurrentCustomer(w, req)

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
				customer := &domain.Customer{
					ID: testID,
					User: &domain.User{
						Name:  "John Doe",
						Email: "john@example.com",
					},
				}
				mockService.On("GetByID", mock.Anything, testID.String()).Return(customer, nil)
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
				customers := []domain.Customer{
					{
						ID: uuid.New(),
						User: &domain.User{
							Name:  "John Doe",
							Email: "john@example.com",
						},
					},
					{
						ID: uuid.New(),
						User: &domain.User{
							Name:  "Jane Smith",
							Email: "jane@example.com",
						},
					},
				}
				mockService.On("List", mock.Anything).Return(customers, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "Internal Error",
			setupMock: func() {
				mockService.On("List", mock.Anything).Return(nil, customErrors.ErrInternalServer)
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

func TestCustomerHandler_Routes(t *testing.T) {
	_, handler := setupCustomerTest()
	router := handler.Routes()

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/"},
		{http.MethodGet, "/me"},
		{http.MethodGet, "/"},
		{http.MethodGet, "/{id}"},
		{http.MethodDelete, "/{id}"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			assert.NotPanics(t, func() {
				router.Match(chi.NewRouteContext(), route.method, route.path)
			})
		})
	}
}
