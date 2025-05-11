package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/bekbull/online-shop/services/product-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductService defines the interface for the product service
type ProductService interface {
	CreateProduct(product *domain.Product) (*domain.Product, error)
	GetProduct(id string) (*domain.Product, error)
	UpdateProduct(product *domain.Product) (*domain.Product, error)
	DeleteProduct(id string) error
	ListProducts(params domain.ListProductsParams) ([]*domain.Product, int, error)
	UpdateInventory(productID string, quantityChange int, operationID, operationType string) (*domain.InventoryInfo, error)
	CheckStock(productID string, quantity int) (bool, int, error)
}

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service ProductService
	logger  *slog.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(service ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the product routes with the given router
func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.ListProducts)
		r.Get("/{id}", h.GetProduct)
		r.Put("/{id}", h.UpdateProduct)
		r.Delete("/{id}", h.DeleteProduct)

		// Inventory management endpoints
		r.Post("/{id}/inventory", h.UpdateInventory)
		r.Get("/{id}/stock", h.CheckStock)
	})
}

// CreateProduct handles POST /v1/products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HTTP CreateProduct called")

	// Decode request body
	var productRequest struct {
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Price       float64              `json:"price"`
		ImageURLs   []string             `json:"image_urls"`
		Category    string               `json:"category"`
		Inventory   domain.InventoryInfo `json:"inventory"`
		Tags        []string             `json:"tags"`
		Attributes  map[string]string    `json:"attributes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&productRequest); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create domain product
	product := &domain.Product{
		Name:        productRequest.Name,
		Description: productRequest.Description,
		Price:       productRequest.Price,
		ImageURLs:   productRequest.ImageURLs,
		Category:    productRequest.Category,
		Inventory:   productRequest.Inventory,
		Tags:        productRequest.Tags,
		Attributes:  productRequest.Attributes,
	}

	// Call service
	createdProduct, err := h.service.CreateProduct(product)
	if err != nil {
		h.logger.Error("Failed to create product", "error", err)
		http.Error(w, "Failed to create product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdProduct); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// GetProduct handles GET /v1/products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.logger.Info("HTTP GetProduct called", "id", id)

	// Call service
	product, err := h.service.GetProduct(id)
	if err != nil {
		h.logger.Error("Failed to get product", "id", id, "error", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// UpdateProduct handles PUT /v1/products/{id}
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.logger.Info("HTTP UpdateProduct called", "id", id)

	// Parse ID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.logger.Error("Invalid product ID format", "id", id)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var productRequest struct {
		Name        string                `json:"name"`
		Description string                `json:"description"`
		Price       float64               `json:"price"`
		ImageURLs   []string              `json:"image_urls"`
		Category    string                `json:"category"`
		Inventory   *domain.InventoryInfo `json:"inventory"`
		Tags        []string              `json:"tags"`
		Attributes  map[string]string     `json:"attributes"`
		Active      *bool                 `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&productRequest); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create domain product
	product := &domain.Product{
		ID:          objectID,
		Name:        productRequest.Name,
		Description: productRequest.Description,
		Price:       productRequest.Price,
		ImageURLs:   productRequest.ImageURLs,
		Category:    productRequest.Category,
		Tags:        productRequest.Tags,
		Attributes:  productRequest.Attributes,
	}

	// Set active status if provided
	if productRequest.Active != nil {
		product.Active = *productRequest.Active
	}

	// Set inventory if provided
	if productRequest.Inventory != nil {
		product.Inventory = *productRequest.Inventory
	}

	// Call service
	updatedProduct, err := h.service.UpdateProduct(product)
	if err != nil {
		h.logger.Error("Failed to update product", "id", id, "error", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedProduct); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// DeleteProduct handles DELETE /v1/products/{id}
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.logger.Info("HTTP DeleteProduct called", "id", id)

	// Call service
	err := h.service.DeleteProduct(id)
	if err != nil {
		h.logger.Error("Failed to delete product", "id", id, "error", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete product: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	w.WriteHeader(http.StatusNoContent)
}

// ListProducts handles GET /v1/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("HTTP ListProducts called")

	// Parse query parameters
	params := domain.ListProductsParams{
		Page:     parseInt(r.URL.Query().Get("page"), 0),
		PageSize: parseInt(r.URL.Query().Get("page_size"), 20),
	}

	// Parse optional filters
	if category := r.URL.Query().Get("category"); category != "" {
		params.Category = category
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		params.Tags = strings.Split(tags, ",")
	}

	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		if p, err := strconv.ParseFloat(minPrice, 64); err == nil {
			params.MinPrice = p
		}
	}

	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		if p, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			params.MaxPrice = p
		}
	}

	if inStock := r.URL.Query().Get("in_stock"); inStock == "true" {
		params.InStockOnly = true
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	if sortDesc := r.URL.Query().Get("sort_desc"); sortDesc == "true" {
		params.SortDesc = true
	}

	if search := r.URL.Query().Get("search"); search != "" {
		params.SearchTerm = search
	}

	// Call service
	products, total, err := h.service.ListProducts(params)
	if err != nil {
		h.logger.Error("Failed to list products", "error", err)
		http.Error(w, "Failed to list products: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination metadata
	totalPages := (total + params.PageSize - 1) / params.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Prepare response
	response := struct {
		Products   []*domain.Product `json:"products"`
		Total      int               `json:"total"`
		Page       int               `json:"page"`
		PageSize   int               `json:"page_size"`
		TotalPages int               `json:"total_pages"`
	}{
		Products:   products,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// UpdateInventory handles POST /v1/products/{id}/inventory
func (h *ProductHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.logger.Info("HTTP UpdateInventory called", "id", id)

	// Decode request body
	var request struct {
		QuantityChange int    `json:"quantity_change"`
		OperationID    string `json:"operation_id"`
		OperationType  string `json:"operation_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call service
	updatedInventory, err := h.service.UpdateInventory(id, request.QuantityChange, request.OperationID, request.OperationType)
	if err != nil {
		h.logger.Error("Failed to update inventory", "id", id, "error", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "insufficient") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to update inventory: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	response := struct {
		Success   bool                  `json:"success"`
		Inventory *domain.InventoryInfo `json:"inventory"`
		Message   string                `json:"message"`
	}{
		Success:   true,
		Inventory: updatedInventory,
		Message:   "Inventory updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// CheckStock handles GET /v1/products/{id}/stock
func (h *ProductHandler) CheckStock(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.logger.Info("HTTP CheckStock called", "id", id)

	// Parse quantity parameter
	quantity := parseInt(r.URL.Query().Get("quantity"), 1)

	// Call service
	available, currentStock, err := h.service.CheckStock(id, quantity)
	if err != nil {
		h.logger.Error("Failed to check stock", "id", id, "error", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to check stock: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	response := struct {
		Available    bool `json:"available"`
		CurrentStock int  `json:"current_stock"`
		Requested    int  `json:"requested"`
	}{
		Available:    available,
		CurrentStock: currentStock,
		Requested:    quantity,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// Helper function to parse int parameters with default value
func parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
