package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/service"
	"github.com/grocery-service/utils/api"
	customErrors "github.com/grocery-service/utils/errors"
)

type OrderHandler struct {
	service service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Get("/customer/{customerID}", h.ListByCustomerID)
	r.Put("/{id}/status", h.UpdateStatus)
	r.Post("/{id}/items", h.AddOrderItem)
	r.Delete("/{id}/items/{itemID}", h.RemoveOrderItem)

	return r
}

// @Summary Create a new order
// @Description Create a new order with the provided data
// @Tags orders
// @Accept json
// @Produce json
// @Param order body domain.Order true "Order object"
// @Success 201 {object} api.Response{data=domain.Order}
// @Failure 400 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var order domain.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), &order); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidOrderData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, customErrors.ErrInsufficientStock):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to create order", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, order, http.StatusCreated)
}

// @Summary Get an order by ID
// @Description Get an order's details by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID" format(uuid)
// @Success 200 {object} api.Response{data=domain.Order}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			api.ErrorResponse(w, "Order not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to get order", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, order, http.StatusOK)
}

// @Summary List all orders
// @Description Get a list of all orders
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {object} api.Response{data=[]domain.Order}
// @Failure 500 {object} api.Response
// @Router /api/v1/orders [get]
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		api.ErrorResponse(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	api.SuccessResponse(w, orders, http.StatusOK)
}

// @Summary List customer orders
// @Description Get all orders for a specific customer
// @Tags orders
// @Accept json
// @Produce json
// @Param customerID path string true "Customer ID" format(uuid)
// @Success 200 {object} api.Response{data=[]domain.Order}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/orders/customer/{customerID} [get]
func (h *OrderHandler) ListByCustomerID(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerID")
	if _, err := uuid.Parse(customerID); err != nil {
		api.ErrorResponse(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	orders, err := h.service.ListByCustomerID(r.Context(), customerID)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCustomerNotFound):
			api.ErrorResponse(w, "Customer not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to list customer orders", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, orders, http.StatusOK)
}

// @Summary Update order status
// @Description Update the status of an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID" format(uuid)
// @Param status body object true "Status object" schema(properties(status=string))
// @Success 200 {object} api.Response
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/orders/{id}/status [put]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Status domain.OrderStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, request.Status); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			api.ErrorResponse(w, "Order not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrOrderStatusInvalid):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to update order status", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, nil, http.StatusOK)
}

// @Summary Add order item
// @Description Add a new item to an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID" format(uuid)
// @Param item body domain.OrderItem true "Order item object"
// @Success 201 {object} api.Response{data=domain.OrderItem}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/orders/{id}/items [post]
func (h *OrderHandler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(orderID); err != nil {
		api.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var item domain.OrderItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AddOrderItem(r.Context(), orderID, &item); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			api.ErrorResponse(w, "Order not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrInvalidOrderData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, customErrors.ErrInsufficientStock):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, customErrors.ErrOrderStatusInvalid):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to add order item", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, item, http.StatusCreated)
}

// @Summary Remove order item
// @Description Remove an item from an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID" format(uuid)
// @Param itemID path string true "Item ID" format(uuid)
// @Success 204 {object} api.Response
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/orders/{id}/items/{itemID} [delete]
func (h *OrderHandler) RemoveOrderItem(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(orderID); err != nil {
		api.ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	itemID := chi.URLParam(r, "itemID")
	if _, err := uuid.Parse(itemID); err != nil {
		api.ErrorResponse(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveOrderItem(r.Context(), orderID, itemID); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			api.ErrorResponse(w, "Order not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrOrderItemNotFound):
			api.ErrorResponse(w, "Order item not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrOrderStatusInvalid):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to remove order item", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, nil, http.StatusNoContent)
}
