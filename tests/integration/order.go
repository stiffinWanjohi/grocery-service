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

func setupOrderTest() (*serviceMock.OrderService, *handler.OrderHandler) {
	mockService := new(serviceMock.OrderService)
	handler := handler.NewOrderHandler(mockService)
	return mockService, handler
}

func TestOrderHandler_Create(t *testing.T) {
	mockService, handler := setupOrderTest()

	customerID := uuid.New()
	productID := uuid.New()
	tests := []struct {
		name       string
		order      *domain.Order
		setupMock  func(*domain.Order)
		wantStatus int
		wantError  string
	}{
		{
			name: "Success",
			order: &domain.Order{
				CustomerID: customerID,
				Status:     domain.OrderStatusPending,
				TotalPrice: 29.99,
				Items: []domain.OrderItem{
					{
						ProductID: productID,
						Quantity:  2,
						Price:     14.99,
					},
				},
			},
			setupMock: func(o *domain.Order) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(order *domain.Order) bool {
					return order.CustomerID == o.CustomerID && order.TotalPrice == o.TotalPrice
				})).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid Order",
			order: &domain.Order{
				CustomerID: customerID,
				Status:     domain.OrderStatusPending,
				TotalPrice: -1, // Invalid price
			},
			setupMock: func(o *domain.Order) {
				mockService.On("Create", mock.Anything, mock.MatchedBy(func(order *domain.Order) bool {
					return order.CustomerID == o.CustomerID && order.TotalPrice == o.TotalPrice
				})).Return(customErrors.ErrInvalidOrderData)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid order data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(tt.order)

			jsonBody, _ := json.Marshal(tt.order)
			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(jsonBody))
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
				var returnedOrder domain.Order
				orderData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(orderData, &returnedOrder)
				assert.NoError(t, err)
				assert.Equal(t, tt.order.CustomerID, returnedOrder.CustomerID)
				assert.Equal(t, tt.order.TotalPrice, returnedOrder.TotalPrice)
			}
		})
	}
}

func TestOrderHandler_GetByID(t *testing.T) {
	mockService, handler := setupOrderTest()

	testID := uuid.New()
	customerID := uuid.New()
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
				mockService.On("GetByID", mock.Anything, testID.String()).Return(&domain.Order{
					ID:         testID,
					CustomerID: customerID,
					TotalPrice: 49.98,
					Status:     domain.OrderStatusPreparing,
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid UUID",
			id:         "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid order ID",
		},
		{
			name: "Not Found",
			id:   testID.String(),
			setupMock: func() {
				mockService.On("GetByID", mock.Anything, testID.String()).Return(nil, customErrors.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Order not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req := httptest.NewRequest(http.MethodGet, "/orders/"+tt.id, nil)
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
				var returnedOrder domain.Order
				orderData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(orderData, &returnedOrder)
				assert.NoError(t, err)
				assert.Equal(t, testID, returnedOrder.ID)
				assert.Equal(t, customerID, returnedOrder.CustomerID)
			}
		})
	}
}

func TestOrderHandler_List(t *testing.T) {
	mockService, handler := setupOrderTest()

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
				mockService.On("List", mock.Anything).Return([]domain.Order{
					{
						ID:         uuid.New(),
						CustomerID: uuid.New(),
						TotalPrice: 29.99,
						Status:     domain.OrderStatusPending,
					},
					{
						ID:         uuid.New(),
						CustomerID: uuid.New(),
						TotalPrice: 49.98,
						Status:     domain.OrderStatusConfirmed,
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
			wantError:  "Failed to list orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/orders", nil)
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
				var orders []domain.Order
				ordersData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(ordersData, &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, tt.wantCount)
			}
		})
	}
}

func TestOrderHandler_ListByCustomerID(t *testing.T) {
	mockService, handler := setupOrderTest()

	customerID := uuid.New()
	tests := []struct {
		name       string
		customerID string
		setupMock  func()
		wantStatus int
		wantCount  int
		wantError  string
	}{
		{
			name:       "Success",
			customerID: customerID.String(),
			setupMock: func() {
				mockService.On("ListByCustomerID", mock.Anything, customerID.String()).Return([]domain.Order{
					{
						ID:         uuid.New(),
						CustomerID: customerID,
						TotalPrice: 29.99,
						Status:     domain.OrderStatusPending,
					},
					{
						ID:         uuid.New(),
						CustomerID: customerID,
						TotalPrice: 49.98,
						Status:     domain.OrderStatusShipped,
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "Invalid Customer ID",
			customerID: "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid customer ID",
		},
		{
			name:       "No Orders Found",
			customerID: customerID.String(),
			setupMock: func() {
				mockService.On("ListByCustomerID", mock.Anything, customerID.String()).Return([]domain.Order{}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("customerID", tt.customerID)
			req := httptest.NewRequest(http.MethodGet, "/orders/customer/"+tt.customerID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			handler.ListByCustomerID(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
				var orders []domain.Order
				ordersData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(ordersData, &orders)
				assert.NoError(t, err)
				assert.Len(t, orders, tt.wantCount)
				if tt.wantCount > 0 {
					assert.Equal(t, customerID, orders[0].CustomerID)
				}
			}
		})
	}
}

func TestOrderHandler_UpdateStatus(t *testing.T) {
	mockService, handler := setupOrderTest()

	testID := uuid.New()
	tests := []struct {
		name       string
		id         string
		status     domain.OrderStatus
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name:   "Success",
			id:     testID.String(),
			status: domain.OrderStatusDelivered,
			setupMock: func() {
				mockService.On("UpdateStatus", mock.Anything, testID.String(), domain.OrderStatusDelivered).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Invalid Status",
			id:     testID.String(),
			status: "invalid_status",
			setupMock: func() {
				mockService.On("UpdateStatus", mock.Anything, testID.String(), mock.AnythingOfType("domain.OrderStatus")).Return(customErrors.ErrOrderStatusInvalid)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid order status",
		},
		{
			name:   "Order Not Found",
			id:     testID.String(),
			status: domain.OrderStatusDelivered,
			setupMock: func() {
				mockService.On("UpdateStatus", mock.Anything, testID.String(), domain.OrderStatusDelivered).Return(customErrors.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Order not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			statusUpdate := struct {
				Status domain.OrderStatus `json:"status"`
			}{
				Status: tt.status,
			}

			jsonBody, _ := json.Marshal(statusUpdate)
			req := httptest.NewRequest(http.MethodPut, "/orders/"+tt.id+"/status", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.UpdateStatus(w, req)

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

func TestOrderHandler_AddOrderItem(t *testing.T) {
	mockService, handler := setupOrderTest()

	orderID := uuid.New()
	productID := uuid.New()
	tests := []struct {
		name       string
		orderID    string
		item       *domain.OrderItem
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name:    "Success",
			orderID: orderID.String(),
			item: &domain.OrderItem{
				ProductID: productID,
				OrderID:   orderID,
				Quantity:  3,
				Price:     9.99,
			},
			setupMock: func() {
				mockService.On("AddOrderItem", mock.Anything, orderID.String(), mock.AnythingOfType("*domain.OrderItem")).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:    "Invalid Order ID",
			orderID: "invalid-uuid",
			item: &domain.OrderItem{
				ProductID: productID,
				Quantity:  3,
				Price:     9.99,
			},
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid order ID",
		},
		{
			name:    "Order Not Found",
			orderID: orderID.String(),
			item: &domain.OrderItem{
				ProductID: productID,
				OrderID:   orderID,
				Quantity:  3,
				Price:     9.99,
			},
			setupMock: func() {
				mockService.On("AddOrderItem", mock.Anything, orderID.String(), mock.AnythingOfType("*domain.OrderItem")).Return(customErrors.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Order not found",
		},
		{
			name:    "Invalid Item Data",
			orderID: orderID.String(),
			item: &domain.OrderItem{
				ProductID: productID,
				OrderID:   orderID,
				Quantity:  0, // Invalid quantity
				Price:     9.99,
			},
			setupMock: func() {
				mockService.On("AddOrderItem", mock.Anything, orderID.String(), mock.AnythingOfType("*domain.OrderItem")).Return(customErrors.ErrInvalidOrderItemData)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid order item data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			jsonBody, _ := json.Marshal(tt.item)
			req := httptest.NewRequest(http.MethodPost, "/orders/"+tt.orderID+"/items", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.orderID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.AddOrderItem(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response api.Response
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.wantError != "" {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, tt.wantError)
			} else {
				assert.True(t, response.Success)
				var returnedItem domain.OrderItem
				itemData, err := json.Marshal(response.Data)
				assert.NoError(t, err)
				err = json.Unmarshal(itemData, &returnedItem)
				assert.NoError(t, err)
				assert.Equal(t, tt.item.ProductID, returnedItem.ProductID)
				assert.Equal(t, tt.item.Quantity, returnedItem.Quantity)
				assert.Equal(t, tt.item.Price, returnedItem.Price)
			}
		})
	}
}

func TestOrderHandler_RemoveOrderItem(t *testing.T) {
	mockService, handler := setupOrderTest()

	orderID := uuid.New()
	itemID := uuid.New()
	tests := []struct {
		name       string
		orderID    string
		itemID     string
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name:    "Success",
			orderID: orderID.String(),
			itemID:  itemID.String(),
			setupMock: func() {
				mockService.On("RemoveOrderItem", mock.Anything, orderID.String(), itemID.String()).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Invalid Order ID",
			orderID:    "invalid-uuid",
			itemID:     itemID.String(),
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid order ID",
		},
		{
			name:       "Invalid Item ID",
			orderID:    orderID.String(),
			itemID:     "invalid-uuid",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid item ID",
		},
		{
			name:    "Order Not Found",
			orderID: orderID.String(),
			itemID:  itemID.String(),
			setupMock: func() {
				mockService.On("RemoveOrderItem", mock.Anything, orderID.String(), itemID.String()).Return(customErrors.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Order not found",
		},
		{
			name:    "Item Not Found",
			orderID: orderID.String(),
			itemID:  itemID.String(),
			setupMock: func() {
				mockService.On("RemoveOrderItem", mock.Anything, orderID.String(), itemID.String()).Return(customErrors.ErrOrderItemNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "Order item not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.orderID)
			rctx.URLParams.Add("itemID", tt.itemID)
			req := httptest.NewRequest(http.MethodDelete, "/orders/"+tt.orderID+"/items/"+tt.itemID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			handler.RemoveOrderItem(w, req)

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
