package service

import (
	"log/slog"
	"os"
	"testing"

	"github.com/bekbull/online-shop/services/product-service/internal/domain"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BenchmarkCreateProduct(b *testing.B) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger with minimal logging to avoid affecting benchmark
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup expectations for all iterations
	mockRepo.On("Create", mock.AnythingOfType("*domain.Product")).Return(nil)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		product := &domain.Product{
			Name:        "Benchmark Product",
			Description: "This is a product for benchmarking",
			Price:       99.99,
			Category:    "Benchmark",
			Inventory: domain.InventoryInfo{
				Quantity: 100,
				SKU:      "BENCH-123",
				InStock:  true,
			},
		}

		_, _ = service.CreateProduct(product)
	}
}

func BenchmarkListProducts(b *testing.B) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger with minimal logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Create test products
	products := make([]*domain.Product, 50)
	for i := 0; i < 50; i++ {
		products[i] = &domain.Product{
			ID:          primitive.NewObjectID(),
			Name:        "Product" + string(rune(i)),
			Description: "Description",
			Price:       float64(i * 10),
			Category:    "Category",
		}
	}

	// Define list parameters
	params := domain.ListProductsParams{
		Page:     0,
		PageSize: 20,
		Category: "Category",
	}

	// Setup expectations
	mockRepo.On("List", params).Return(products[:20], 50, nil)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.ListProducts(params)
	}
}

func BenchmarkUpdateInventory(b *testing.B) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger with minimal logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup test data
	productID := primitive.NewObjectID().Hex()
	operationID := "bench-op-123"
	operationType := "restock"

	updatedInventory := &domain.InventoryInfo{
		Quantity: 110,
		SKU:      "TEST-SKU-123",
		InStock:  true,
		Reserved: 0,
	}

	// Setup expectations for restocking
	mockRepo.On("UpdateInventory", productID, 10, operationID, operationType).
		Return(updatedInventory, nil)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.UpdateInventory(productID, 10, operationID, operationType)
	}
}

func BenchmarkCheckStock(b *testing.B) {
	// Create mock repository
	mockRepo := new(MockProductRepository)

	// Create logger with minimal logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create service with mock repository
	service := New(mockRepo, logger)

	// Setup test data
	productID := primitive.NewObjectID().Hex()

	// Setup expectations
	mockRepo.On("CheckStock", productID, 5).Return(true, 10, nil)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.CheckStock(productID, 5)
	}
}
