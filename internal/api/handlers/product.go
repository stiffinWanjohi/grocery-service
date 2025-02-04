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

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/category/{categoryID}", h.ListByCategoryID)
	r.Put("/{id}/stock", h.UpdateStock)

	return r
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), &product); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidProductData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			api.ErrorResponse(w, "Invalid category", http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to create product", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, product, http.StatusCreated)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			api.ErrorResponse(w, "Product not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to get product", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, product, http.StatusOK)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.List(r.Context())
	if err != nil {
		api.ErrorResponse(w, "Failed to list products", http.StatusInternalServerError)
		return
	}

	api.SuccessResponse(w, products, http.StatusOK)
}

func (h *ProductHandler) ListByCategoryID(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryID")
	if _, err := uuid.Parse(categoryID); err != nil {
		api.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	products, err := h.service.ListByCategoryID(r.Context(), categoryID)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			api.ErrorResponse(w, "Category not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to list products", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, products, http.StatusOK)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product.ID = uuid.MustParse(id)
	if err := h.service.Update(r.Context(), &product); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			api.ErrorResponse(w, "Product not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrInvalidProductData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to update product", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, product, http.StatusOK)
}

func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStock(r.Context(), id, request.Quantity); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			api.ErrorResponse(w, "Product not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrInvalidProductData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to update product stock", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, nil, http.StatusOK)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			api.ErrorResponse(w, "Product not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to delete product", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, nil, http.StatusNoContent)
}