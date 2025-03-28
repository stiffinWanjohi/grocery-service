package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/grocery-service/internal/api/middleware"
	_ "github.com/grocery-service/internal/domain" // we need this
	"github.com/grocery-service/internal/service"
	"github.com/grocery-service/utils/api"
	customErrors "github.com/grocery-service/utils/errors"
)

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(
	service service.CustomerService,
) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequireAuth)

	// Regular user routes
	r.Post("/", h.Create)
	r.Get("/me", h.GetCurrentCustomer)

	// Admin only routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAdmin)
		r.Get("/", h.List)
		r.Get("/{id}", h.GetByID)
		r.Delete("/{id}", h.Delete)
	})

	return r
}

// @Summary Create customer profile
// @Description Create a customer profile for the authenticated user
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} api.Response{data=domain.Customer}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.Response
// @Failure 409 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /customers [post]
func (h *CustomerHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	customer, err := h.service.Create(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidCustomerData):
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
		case errors.Is(err, customErrors.ErrCustomerExists):
			if err := api.ErrorResponse(
				w,
				"Customer profile already exists for this user",
				http.StatusConflict,
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
				"Failed to create customer profile",
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

	if err := api.SuccessResponse(w, customer, http.StatusCreated); err != nil {
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

// @Summary Get current customer profile
// @Description Get the customer profile for the authenticated user
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response{data=domain.Customer}
// @Failure 401 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /customers/me [get]
func (h *CustomerHandler) GetCurrentCustomer(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	customer, err := h.service.GetByUserID(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCustomerNotFound):
			if err := api.ErrorResponse(
				w,
				"Customer profile not found",
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
				"Failed to get customer profile",
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

	if err := api.SuccessResponse(w, customer, http.StatusOK); err != nil {
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

// @Summary Get customer by ID
// @Description Get a customer profile by ID (admin only)
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 200 {object} api.Response{data=domain.Customer}
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
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

	customer, err := h.service.GetByID(r.Context(), id)
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
				"Failed to get customer",
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

	if err := api.SuccessResponse(w, customer, http.StatusOK); err != nil {
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

// @Summary List all customers
// @Description Get a list of all customers (admin only)
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} api.Response{data=[]domain.Customer}
// @Failure 401 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /customers [get]
func (h *CustomerHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	customers, err := h.service.List(r.Context())
	if err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to list customers",
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

	if err := api.SuccessResponse(w, customers, http.StatusOK); err != nil {
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

// @Summary Delete a customer
// @Description Delete a customer profile by ID (admin only)
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 204 {object} api.Response
// @Failure 400 {object} api.Response
// @Failure 401 {object} api.Response
// @Failure 403 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /customers/{id} [delete]
func (h *CustomerHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
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

	if err := h.service.Delete(r.Context(), id); err != nil {
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
				"Failed to delete customer",
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

	if err := api.SuccessResponse(w, nil, http.StatusNoContent); err != nil {
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
