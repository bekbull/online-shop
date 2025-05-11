package service

import (
	"errors"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/bekbull/online-shop/services/product-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockProductRepository is a mock implementation of the domain.ProductRepository interface
type MockProductRepository struct {
	mock.Mock
}

// Methods implementing the domain.ProductRepository interface

func (m *MockProductRepository) Create(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(id string) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) List(params domain.ListProductsParams) ([]*domain.Product, int, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Product), args.Int(1), args.Error(2)
}

func (m *MockProductRepository) UpdateInventory(productID string, quantityChange int, operationID, operationType string) (*domain.InventoryInfo, error) {
	args := m.Called(productID, quantityChange, operationID, operationType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryInfo), args.Error(1)
}

func (m *MockProductRepository) CheckStock(productID string, quantity int) (bool, int, error) {
	args := m.Called(productID, quantity)
	return args.Bool(0), args.Int(1), args.Error(2)
}

// Helper function to create a test product
func createTestProduct() *domain.Product {
	return &domain.Product{
		ID:          primitive.NewObjectID(),
		Name:        "Test Product",
		Description: "This is a test product",
		Price:       99.99,
		ImageURLs:   []string{"http://example.com/image.jpg"},
		Category:    "Electronics",
		Inventory: domain.InventoryInfo{
			Quantity: 100,
			SKU:      "TEST-SKU-123",
			InStock:  true,
			Reserved: 0,
		},
		Tags:       []string{"test", "electronics"},
		Attributes: map[string]string{"color": "black", "size": "medium"},
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestCreateProduct(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Create test product
	product := createTestProduct()

	// Setup expectations
	mockRepo.On("Create", mock.AnythingOfType("*domain.Product")).Return(nil)

	// Call the service method
	createdProduct, err := service.CreateProduct(product)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, createdProduct)
	assert.Equal(t, product.Name, createdProduct.Name)
	assert.Equal(t, product.Price, createdProduct.Price)
	assert.True(t, createdProduct.Active)
	assert.False(t, createdProduct.CreatedAt.IsZero())
	assert.False(t, createdProduct.UpdatedAt.IsZero())

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestCreateProduct_ValidationError(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Test cases for validation errors
	testCases := []struct {
		name        string
		product     *domain.Product
		expectedErr string
	}{
		{
			name: "Empty name",
			product: &domain.Product{
				Price:    99.99,
				Category: "Electronics",
				Inventory: domain.InventoryInfo{
					SKU: "TEST-SKU-123",
				},
			},
			expectedErr: "product name is required",
		},
		{
			name: "Zero price",
			product: &domain.Product{
				Name:     "Test Product",
				Price:    0,
				Category: "Electronics",
				Inventory: domain.InventoryInfo{
					SKU: "TEST-SKU-123",
				},
			},
			expectedErr: "product price must be greater than zero",
		},
		{
			name: "Empty category",
			product: &domain.Product{
				Name:  "Test Product",
				Price: 99.99,
				Inventory: domain.InventoryInfo{
					SKU: "TEST-SKU-123",
				},
			},
			expectedErr: "product category is required",
		},
		{
			name: "Empty SKU",
			product: &domain.Product{
				Name:     "Test Product",
				Price:    99.99,
				Category: "Electronics",
			},
			expectedErr: "product SKU is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the service method
			_, err := service.CreateProduct(tc.product)

			// Assert error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestGetProduct(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Create test product
	product := createTestProduct()
	productID := product.ID.Hex()

	// Setup expectations
	mockRepo.On("GetByID", productID).Return(product, nil)

	// Call the service method
	fetchedProduct, err := service.GetProduct(productID)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, fetchedProduct)
	assert.Equal(t, product.ID, fetchedProduct.ID)
	assert.Equal(t, product.Name, fetchedProduct.Name)

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestGetProduct_NotFound(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup expectations
	mockRepo.On("GetByID", "non-existent-id").Return(nil, errors.New("product not found"))

	// Call the service method
	fetchedProduct, err := service.GetProduct("non-existent-id")

	// Assert expectations
	assert.Error(t, err)
	assert.Nil(t, fetchedProduct)
	assert.Contains(t, err.Error(), "product not found")

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestUpdateProduct(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Create test product
	existingProduct := createTestProduct()
	productID := existingProduct.ID.Hex()

	// Create update product
	updateProduct := &domain.Product{
		ID:          existingProduct.ID,
		Name:        "Updated Product",
		Description: "This is an updated product",
		Price:       129.99,
	}

	// Expected updated product
	expectedProduct := createTestProduct()
	expectedProduct.Name = updateProduct.Name
	expectedProduct.Description = updateProduct.Description
	expectedProduct.Price = updateProduct.Price
	expectedProduct.UpdatedAt = time.Now() // This will be different in the actual result

	// Setup expectations
	mockRepo.On("GetByID", productID).Return(existingProduct, nil)
	mockRepo.On("Update", mock.AnythingOfType("*domain.Product")).Return(nil)

	// Call the service method
	updatedProduct, err := service.UpdateProduct(updateProduct)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, updatedProduct)
	assert.Equal(t, updateProduct.Name, updatedProduct.Name)
	assert.Equal(t, updateProduct.Description, updatedProduct.Description)
	assert.Equal(t, updateProduct.Price, updatedProduct.Price)
	assert.Equal(t, existingProduct.Category, updatedProduct.Category) // Unchanged
	assert.Equal(t, existingProduct.Tags, updatedProduct.Tags)         // Unchanged
	assert.Equal(t, existingProduct.Active, updatedProduct.Active)     // Unchanged
	assert.True(t, updatedProduct.UpdatedAt.After(existingProduct.UpdatedAt) ||
		updatedProduct.UpdatedAt.Equal(existingProduct.UpdatedAt))

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestDeleteProduct(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup expectations
	productID := primitive.NewObjectID().Hex()
	mockRepo.On("Delete", productID).Return(nil)

	// Call the service method
	err := service.DeleteProduct(productID)

	// Assert expectations
	assert.NoError(t, err)

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestListProducts(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Create test products
	product1 := createTestProduct()
	product2 := createTestProduct()
	product2.Name = "Another Product"

	products := []*domain.Product{product1, product2}
	totalCount := 2

	// Define list parameters
	params := domain.ListProductsParams{
		Page:     0,
		PageSize: 10,
		Category: "Electronics",
	}

	// Setup expectations
	mockRepo.On("List", params).Return(products, totalCount, nil)

	// Call the service method
	listedProducts, total, err := service.ListProducts(params)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, listedProducts)
	assert.Equal(t, totalCount, total)
	assert.Len(t, listedProducts, 2)
	assert.Equal(t, product1.Name, listedProducts[0].Name)
	assert.Equal(t, product2.Name, listedProducts[1].Name)

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestUpdateInventory(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup test data
	productID := primitive.NewObjectID().Hex()
	quantityChange := 10
	operationID := "op-123"
	operationType := "restock"

	updatedInventory := &domain.InventoryInfo{
		Quantity: 110,
		SKU:      "TEST-SKU-123",
		InStock:  true,
		Reserved: 0,
	}

	// Setup expectations for restocking
	mockRepo.On("UpdateInventory", productID, quantityChange, operationID, operationType).
		Return(updatedInventory, nil)

	// Call the service method
	inventory, err := service.UpdateInventory(productID, quantityChange, operationID, operationType)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, inventory)
	assert.Equal(t, updatedInventory.Quantity, inventory.Quantity)
	assert.Equal(t, updatedInventory.InStock, inventory.InStock)

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)

	// Test purchase operation with stock check
	mockRepo.ExpectedCalls = nil
	operationType = "purchase"
	quantityChange = -5

	// First check stock
	mockRepo.On("CheckStock", productID, 5).Return(true, 110, nil)

	// Then update inventory
	mockRepo.On("UpdateInventory", productID, quantityChange, operationID, operationType).
		Return(updatedInventory, nil)

	// Call the service method
	inventory, err = service.UpdateInventory(productID, quantityChange, operationID, operationType)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, inventory)

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestUpdateInventory_InsufficientStock(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup test data
	productID := primitive.NewObjectID().Hex()
	quantityChange := -20
	operationID := "op-123"
	operationType := "purchase"

	// Setup expectations - not enough stock
	mockRepo.On("CheckStock", productID, 20).Return(false, 10, nil)

	// Call the service method
	inventory, err := service.UpdateInventory(productID, quantityChange, operationID, operationType)

	// Assert expectations
	assert.Error(t, err)
	assert.Nil(t, inventory)
	assert.Contains(t, err.Error(), "insufficient stock")

	// Verify that mock expectations were met
	mockRepo.AssertExpectations(t)
}

func TestCheckStock(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup test data
	productID := primitive.NewObjectID().Hex()

	// Test cases
	testCases := []struct {
		name            string
		quantity        int
		mockAvailable   bool
		mockStock       int
		mockErr         error
		expectAvailable bool
		expectStock     int
		expectErr       bool
	}{
		{
			name:            "Available stock",
			quantity:        5,
			mockAvailable:   true,
			mockStock:       10,
			mockErr:         nil,
			expectAvailable: true,
			expectStock:     10,
			expectErr:       false,
		},
		{
			name:            "Insufficient stock",
			quantity:        15,
			mockAvailable:   false,
			mockStock:       10,
			mockErr:         nil,
			expectAvailable: false,
			expectStock:     10,
			expectErr:       false,
		},
		{
			name:            "Product not found",
			quantity:        5,
			mockAvailable:   false,
			mockStock:       0,
			mockErr:         errors.New("product not found"),
			expectAvailable: false,
			expectStock:     0,
			expectErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup expectations
			mockRepo.On("CheckStock", productID, tc.quantity).
				Return(tc.mockAvailable, tc.mockStock, tc.mockErr).Once()

			// Call the service method
			available, stock, err := service.CheckStock(productID, tc.quantity)

			// Assert expectations
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectAvailable, available)
				assert.Equal(t, tc.expectStock, stock)
			}
		})
	}

	// Verify that all mock expectations were met
	mockRepo.AssertExpectations(t)
}
