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

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/subcategories", h.ListByParentID)

	return r
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var category domain.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), &category); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidCategoryData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to create category", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, category, http.StatusCreated)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			api.ErrorResponse(w, "Category not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to get category", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, category, http.StatusOK)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		api.ErrorResponse(w, "Failed to list categories", http.StatusInternalServerError)
		return
	}

	api.SuccessResponse(w, categories, http.StatusOK)
}

func (h *CategoryHandler) ListByParentID(w http.ResponseWriter, r *http.Request) {
	parentID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(parentID); err != nil {
		api.ErrorResponse(w, "Invalid parent category ID", http.StatusBadRequest)
		return
	}

	categories, err := h.service.ListByParentID(r.Context(), parentID)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			api.ErrorResponse(w, "Parent category not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to list subcategories", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, categories, http.StatusOK)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category domain.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		api.ErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	category.ID = uuid.MustParse(id)
	if err := h.service.Update(r.Context(), &category); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			api.ErrorResponse(w, "Category not found", http.StatusNotFound)
		case errors.Is(err, customErrors.ErrInvalidCategoryData):
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			api.ErrorResponse(w, "Failed to update category", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, category, http.StatusOK)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		api.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			api.ErrorResponse(w, "Category not found", http.StatusNotFound)
		default:
			api.ErrorResponse(w, "Failed to delete category", http.StatusInternalServerError)
		}
		return
	}

	api.SuccessResponse(w, nil, http.StatusNoContent)
}
