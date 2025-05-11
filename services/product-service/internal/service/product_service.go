package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bekbull/online-shop/services/product-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductService provides business logic for product operations
type ProductService struct {
	repo   domain.ProductRepository
	logger *slog.Logger
}

// New creates a new ProductService
func New(repo domain.ProductRepository, logger *slog.Logger) *ProductService {
	return &ProductService{
		repo:   repo,
		logger: logger,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(product *domain.Product) (*domain.Product, error) {
	s.logger.Info("Creating new product", "name", product.Name)

	// Validate product data
	if err := validateProduct(product); err != nil {
		s.logger.Error("Product validation failed", "error", err)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Set default values
	if product.ID.IsZero() {
		product.ID = primitive.NewObjectID()
	}
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.Active = true
	product.Inventory.InStock = product.Inventory.Quantity > 0

	// Persist product
	if err := s.repo.Create(product); err != nil {
		s.logger.Error("Failed to create product", "error", err)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	s.logger.Info("Product created successfully", "id", product.ID.Hex())
	return product, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(id string) (*domain.Product, error) {
	s.logger.Info("Getting product", "id", id)

	product, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to get product", "id", id, "error", err)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	return product, nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(product *domain.Product) (*domain.Product, error) {
	s.logger.Info("Updating product", "id", product.ID.Hex())

	// Check if product exists
	existingProduct, err := s.repo.GetByID(product.ID.Hex())
	if err != nil {
		s.logger.Error("Failed to find product for update", "id", product.ID.Hex(), "error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Update fields that can be changed
	if product.Name != "" {
		existingProduct.Name = product.Name
	}
	if product.Description != "" {
		existingProduct.Description = product.Description
	}
	if product.Price > 0 {
		existingProduct.Price = product.Price
	}
	if len(product.ImageURLs) > 0 {
		existingProduct.ImageURLs = product.ImageURLs
	}
	if product.Category != "" {
		existingProduct.Category = product.Category
	}
	if len(product.Tags) > 0 {
		existingProduct.Tags = product.Tags
	}
	if len(product.Attributes) > 0 {
		existingProduct.Attributes = product.Attributes
	}

	// Handle inventory update if provided
	if product.Inventory.SKU != "" {
		existingProduct.Inventory.SKU = product.Inventory.SKU
	}

	// Only update quantity through dedicated inventory update methods
	// This prevents accidental inventory changes

	// Update timestamp and status
	existingProduct.UpdatedAt = time.Now()

	// Update active status if it was explicitly set
	// This allows for product activation/deactivation
	if product.Active != existingProduct.Active {
		existingProduct.Active = product.Active
	}

	// Persist changes
	if err := s.repo.Update(existingProduct); err != nil {
		s.logger.Error("Failed to update product", "id", existingProduct.ID.Hex(), "error", err)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	s.logger.Info("Product updated successfully", "id", existingProduct.ID.Hex())
	return existingProduct, nil
}

// DeleteProduct removes a product
func (s *ProductService) DeleteProduct(id string) error {
	s.logger.Info("Deleting product", "id", id)

	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("Failed to delete product", "id", id, "error", err)
		return fmt.Errorf("repository error: %w", err)
	}

	s.logger.Info("Product deleted successfully", "id", id)
	return nil
}

// ListProducts retrieves a list of products based on filters
func (s *ProductService) ListProducts(params domain.ListProductsParams) ([]*domain.Product, int, error) {
	s.logger.Info("Listing products",
		"page", params.Page,
		"pageSize", params.PageSize,
		"category", params.Category,
		"inStockOnly", params.InStockOnly)

	// Apply default values if not provided
	if params.PageSize <= 0 {
		params.PageSize = 20 // Default page size
	}
	if params.Page < 0 {
		params.Page = 0 // First page
	}

	products, total, err := s.repo.List(params)
	if err != nil {
		s.logger.Error("Failed to list products", "error", err)
		return nil, 0, fmt.Errorf("repository error: %w", err)
	}

	s.logger.Info("Products listed successfully", "count", len(products), "total", total)
	return products, total, nil
}

// UpdateInventory updates a product's inventory
func (s *ProductService) UpdateInventory(productID string, quantityChange int, operationID, operationType string) (*domain.InventoryInfo, error) {
	s.logger.Info("Updating inventory",
		"productID", productID,
		"quantityChange", quantityChange,
		"operationType", operationType)

	// Validate operation type
	validOperationTypes := map[string]bool{
		"purchase":    true,
		"restock":     true,
		"reservation": true,
		"release":     true,
		"adjustment":  true,
	}
	if !validOperationTypes[operationType] {
		return nil, errors.New("invalid operation type")
	}

	// For purchase and reservation operations, check if there's enough stock
	if (operationType == "purchase" || operationType == "reservation") && quantityChange < 0 {
		available, current, err := s.repo.CheckStock(productID, -quantityChange)
		if err != nil {
			s.logger.Error("Failed to check stock", "productID", productID, "error", err)
			return nil, fmt.Errorf("stock check error: %w", err)
		}
		if !available {
			s.logger.Error("Insufficient stock", "productID", productID, "required", -quantityChange, "available", current)
			return nil, errors.New("insufficient stock")
		}
	}

	// Update inventory
	updatedInventory, err := s.repo.UpdateInventory(productID, quantityChange, operationID, operationType)
	if err != nil {
		s.logger.Error("Failed to update inventory", "productID", productID, "error", err)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	s.logger.Info("Inventory updated successfully",
		"productID", productID,
		"newQuantity", updatedInventory.Quantity)
	return updatedInventory, nil
}

// CheckStock checks if a product has sufficient stock
func (s *ProductService) CheckStock(productID string, quantity int) (bool, int, error) {
	s.logger.Info("Checking stock", "productID", productID, "quantity", quantity)

	available, current, err := s.repo.CheckStock(productID, quantity)
	if err != nil {
		s.logger.Error("Failed to check stock", "productID", productID, "error", err)
		return false, 0, fmt.Errorf("repository error: %w", err)
	}

	s.logger.Info("Stock check completed",
		"productID", productID,
		"available", available,
		"currentStock", current)
	return available, current, nil
}

// Helper functions

// validateProduct performs basic validation on product data
func validateProduct(product *domain.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Price <= 0 {
		return errors.New("product price must be greater than zero")
	}
	if product.Category == "" {
		return errors.New("product category is required")
	}
	if product.Inventory.SKU == "" {
		return errors.New("product SKU is required")
	}
	return nil
}
