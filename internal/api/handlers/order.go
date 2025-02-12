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

func NewOrderHandler(
	service service.OrderService,
) *OrderHandler {
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
// @Router /orders [post]
func (h *OrderHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var order domain.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	if err := h.service.Create(r.Context(), &order); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidOrderData):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrInsufficientStock):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		default:
			if err := api.ErrorResponse(
				w,
				"Failed to create order",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		}
		return
	}
	if err := api.SuccessResponse(w, order, http.StatusCreated); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
	}
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
// @Router /orders/{id} [get]
func (h *OrderHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid order ID",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	order, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			if err := api.ErrorResponse(
				w,
				"Order not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		default:
			if err := api.ErrorResponse(
				w,
				"Failed to get order",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		}
		return
	}

	if err := api.SuccessResponse(w, order, http.StatusOK); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
	}
}

// @Summary List all orders
// @Description Get a list of all orders
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {object} api.Response{data=[]domain.Order}
// @Failure 500 {object} api.Response
// @Router /orders [get]
func (h *OrderHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to list orders",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}
	if err := api.SuccessResponse(w, orders, http.StatusOK); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
	}
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
// @Router /orders/customer/{customerID} [get]
func (h *OrderHandler) ListByCustomerID(
	w http.ResponseWriter,
	r *http.Request,
) {
	customerID := chi.URLParam(r, "customerID")
	if _, err := uuid.Parse(customerID); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid customer ID",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}
	orders, err := h.service.ListByCustomerID(
		r.Context(),
		customerID,
	)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCustomerNotFound):
			if err := api.ErrorResponse(
				w,
				"Customer not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		default:
			if err := api.ErrorResponse(
				w,
				"Failed to list customer orders",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		}
		return
	}
	if err := api.SuccessResponse(w, orders, http.StatusOK); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
	}
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
// @Router /orders/{id}/status [put]
func (h *OrderHandler) UpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid order ID",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	var request struct {
		Status domain.OrderStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, request.Status); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			if err := api.ErrorResponse(
				w,
				"Order not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrOrderStatusInvalid):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		default:
			if err := api.ErrorResponse(
				w,
				"Failed to update order status",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		}
		return
	}

	if err := api.SuccessResponse(w, nil, http.StatusOK); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
	}
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
// @Router /orders/{id}/items [post]
func (h *OrderHandler) AddOrderItem(
	w http.ResponseWriter,
	r *http.Request,
) {
	orderID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(orderID); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid order ID",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	var item domain.OrderItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	if err := h.service.AddOrderItem(r.Context(), orderID, &item); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			if err := api.ErrorResponse(
				w,
				"Order not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrInvalidOrderItemData):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrInsufficientStock):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrOrderStatusInvalid):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		default:
			if err := api.ErrorResponse(
				w,
				"Failed to add order item",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		}
		return
	}

	if err := api.SuccessResponse(w, item, http.StatusCreated); err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to send response",
			http.StatusInternalServerError,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
	}
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
// @Router /orders/{id}/items/{itemID} [delete]
func (h *OrderHandler) RemoveOrderItem(
	w http.ResponseWriter,
	r *http.Request,
) {
	orderID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(orderID); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid order ID",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	itemID := chi.URLParam(r, "itemID")
	if _, err := uuid.Parse(itemID); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid item ID",
			http.StatusBadRequest,
		); err != nil {
			http.Error(
				w,
				"Failed to send error response",
				http.StatusInternalServerError,
			)
		}
		return
	}

	err := h.service.RemoveOrderItem(
		r.Context(),
		orderID,
		itemID,
	)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrOrderNotFound):
			if err := api.ErrorResponse(
				w,
				"Order not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrOrderItemNotFound):
			if err := api.ErrorResponse(
				w,
				"Order item not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrOrderStatusInvalid):
			if err := api.ErrorResponse(
				w,
				err.Error(),
				http.StatusBadRequest,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		default:
			if err := api.ErrorResponse(
				w,
				"Failed to remove order item",
				http.StatusInternalServerError,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
