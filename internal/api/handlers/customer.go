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

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var customer domain.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), &customer); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidCustomerData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to create customer", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, customer, http.StatusCreated)
}

func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	customer, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCustomerNotFound):
			api.ErrorResponse(w, "Customer not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to get customer", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, customer, http.StatusOK)
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	customers, err := h.service.List(r.Context())
	if err != nil {
		api.ErrorResponse(w, "Failed to list customers", http.StatusInternalServerError)
		return
	}

	api.SuccessResponse(w, customers, http.StatusOK)
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var customer domain.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	customer.ID = uuid.MustParse(id)
	if err := h.service.Update(r.Context(), &customer); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCustomerNotFound):
			api.ErrorResponse(w, "Customer not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrInvalidCustomerData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to update customer", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, customer, http.StatusOK)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCustomerNotFound):
			api.ErrorResponse(w, "Customer not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to delete customer", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, nil, http.StatusNoContent)
}
