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

func NewProductHandler(
	service service.ProductService,
) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/category/{categoryID}", h.ListByCategoryID)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Put("/{id}/stock", h.UpdateStock)

	return r
}

// @Summary Create a new product
// @Description Create a new product with the provided data
// @Tags products
// @Accept json
// @Produce json
// @Param product body domain.Product true "Product object"
// @Success 201 {object} api.Response{data=domain.Product}
// @Failure 400 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /products [post]
func (h *ProductHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
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

	if err := h.service.Create(r.Context(), &product); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrInvalidProductData):
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
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			if err := api.ErrorResponse(
				w,
				"Invalid category",
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
				"Failed to create product",
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

	if err := api.SuccessResponse(w, product, http.StatusCreated); err != nil {
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

// @Summary Get a product by ID
// @Description Get a product's details by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Success 200 {object} api.Response{data=domain.Product}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid product ID",
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

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			if err := api.ErrorResponse(
				w,
				"Product not found",
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
				"Failed to get product",
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

	if err := api.SuccessResponse(w, product, http.StatusOK); err != nil {
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

// @Summary List all products
// @Description Get a list of all products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {object} api.Response{data=[]domain.Product}
// @Failure 500 {object} api.Response
// @Router /products [get]
func (h *ProductHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	products, err := h.service.List(r.Context())
	if err != nil {
		if err := api.ErrorResponse(
			w,
			"Failed to list products",
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

	if err := api.SuccessResponse(w, products, http.StatusOK); err != nil {
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

// @Summary List products by category
// @Description Get all products in a specific category
// @Tags products
// @Accept json
// @Produce json
// @Param categoryID path string true "Category ID" format(uuid)
// @Success 200 {object} api.Response{data=[]domain.Product}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /products/category/{categoryID} [get]
func (h *ProductHandler) ListByCategoryID(
	w http.ResponseWriter,
	r *http.Request,
) {
	categoryID := chi.URLParam(r, "categoryID")
	if _, err := uuid.Parse(categoryID); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid category ID",
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

	products, err := h.service.ListByCategoryID(
		r.Context(),
		categoryID,
	)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCategoryNotFound):
			if err := api.ErrorResponse(
				w,
				"Category not found",
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
				"Failed to list products",
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

	if err := api.SuccessResponse(w, products, http.StatusOK); err != nil {
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

// @Summary Update a product
// @Description Update an existing product's details
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Param product body domain.Product true "Product object"
// @Success 200 {object} api.Response{data=domain.Product}
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /products/{id} [put]
func (h *ProductHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid product ID",
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

	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
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

	product.ID = uuid.MustParse(id)
	if err := h.service.Update(r.Context(), &product); err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			if err := api.ErrorResponse(
				w,
				"Product not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrInvalidProductData):
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
				"Failed to update product",
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

	if err := api.SuccessResponse(w, product, http.StatusOK); err != nil {
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

// @Summary Update product stock
// @Description Update the stock quantity of a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Param request body object true "Stock update request" schema(properties(quantity=integer))
// @Success 200 {object} api.Response
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /products/{id}/stock [put]
func (h *ProductHandler) UpdateStock(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid product ID",
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
		Quantity int `json:"quantity"`
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

	err := h.service.UpdateStock(r.Context(), id, request.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrProductNotFound):
			if err := api.ErrorResponse(
				w,
				"Product not found",
				http.StatusNotFound,
			); err != nil {
				http.Error(
					w,
					"Failed to send error response",
					http.StatusInternalServerError,
				)
			}
		case errors.Is(err, customErrors.ErrInvalidProductData):
			if err := api.ErrorResponse(
				w,
				"Invalid product data",
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
				"Failed to update product stock",
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

// @Summary Delete a product
// @Description Delete a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Success 204 {object} api.Response
// @Failure 400 {object} api.Response
// @Failure 404 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		if err := api.ErrorResponse(
			w,
			"Invalid product ID",
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
		case errors.Is(err, customErrors.ErrProductNotFound):
			if err := api.ErrorResponse(
				w,
				"Product not found",
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
				"Failed to delete product",
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
