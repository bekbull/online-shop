package grpc

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/bekbull/online-shop/proto/product"
	"github.com/bekbull/online-shop/services/product-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductServer implements the gRPC ProductService
type ProductServer struct {
	pb.UnimplementedProductServiceServer
	productService ProductService
	logger         *slog.Logger
}

// ProductService represents the business logic interface for product operations
type ProductService interface {
	CreateProduct(product *domain.Product) (*domain.Product, error)
	GetProduct(id string) (*domain.Product, error)
	UpdateProduct(product *domain.Product) (*domain.Product, error)
	DeleteProduct(id string) error
	ListProducts(params domain.ListProductsParams) ([]*domain.Product, int, error)
	UpdateInventory(productID string, quantityChange int, operationID, operationType string) (*domain.InventoryInfo, error)
	CheckStock(productID string, quantity int) (bool, int, error)
}

// New creates a new ProductServer
func New(service ProductService, logger *slog.Logger) *ProductServer {
	return &ProductServer{
		productService: service,
		logger:         logger,
	}
}

// CreateProduct implements the CreateProduct RPC method
func (s *ProductServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC CreateProduct called", "name", req.Name)

	// Map protobuf request to domain model
	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURLs:   req.ImageUrls,
		Category:    req.Category,
		Inventory: domain.InventoryInfo{
			Quantity: int(req.Inventory.Quantity),
			SKU:      req.Inventory.Sku,
			InStock:  req.Inventory.InStock,
			Reserved: int(req.Inventory.Reserved),
		},
		Tags:       req.Tags,
		Attributes: req.Attributes,
	}

	// Call business logic
	createdProduct, err := s.productService.CreateProduct(product)
	if err != nil {
		s.logger.Error("Failed to create product", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	// Map domain model to protobuf response
	return &pb.ProductResponse{
		Product: domainToProtoProduct(createdProduct),
	}, nil
}

// GetProduct implements the GetProduct RPC method
func (s *ProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC GetProduct called", "id", req.Id)

	// Call business logic
	product, err := s.productService.GetProduct(req.Id)
	if err != nil {
		s.logger.Error("Failed to get product", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	// Map domain model to protobuf response
	return &pb.ProductResponse{
		Product: domainToProtoProduct(product),
	}, nil
}

// UpdateProduct implements the UpdateProduct RPC method
func (s *ProductServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC UpdateProduct called", "id", req.Id)

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		s.logger.Error("Invalid product ID format", "id", req.Id)
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID: %v", err)
	}

	// Map protobuf request to domain model
	product := &domain.Product{
		ID: objectID,
	}

	// Handle optional fields
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if len(req.ImageUrls) > 0 {
		product.ImageURLs = req.ImageUrls
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if len(req.Tags) > 0 {
		product.Tags = req.Tags
	}
	if len(req.Attributes) > 0 {
		product.Attributes = req.Attributes
	}
	if req.Active != nil {
		product.Active = *req.Active
	}

	// Set inventory if provided
	if req.Inventory != nil {
		product.Inventory = domain.InventoryInfo{
			SKU:      req.Inventory.Sku,
			Quantity: int(req.Inventory.Quantity),
			InStock:  req.Inventory.InStock,
			Reserved: int(req.Inventory.Reserved),
		}
	}

	// Call business logic
	updatedProduct, err := s.productService.UpdateProduct(product)
	if err != nil {
		s.logger.Error("Failed to update product", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	// Map domain model to protobuf response
	return &pb.ProductResponse{
		Product: domainToProtoProduct(updatedProduct),
	}, nil
}

// DeleteProduct implements the DeleteProduct RPC method
func (s *ProductServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	s.logger.Info("gRPC DeleteProduct called", "id", req.Id)

	// Call business logic
	err := s.productService.DeleteProduct(req.Id)
	if err != nil {
		s.logger.Error("Failed to delete product", "id", req.Id, "error", err)
		return &pb.DeleteProductResponse{
			Success: false,
			Message: err.Error(),
		}, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &pb.DeleteProductResponse{
		Success: true,
		Message: "Product deleted successfully",
	}, nil
}

// ListProducts implements the ListProducts RPC method
func (s *ProductServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	s.logger.Info("gRPC ListProducts called",
		"page", req.Page,
		"pageSize", req.PageSize,
		"category", req.Category)

	// Map protobuf request to domain params
	params := domain.ListProductsParams{
		Page:        int(req.Page),
		PageSize:    int(req.PageSize),
		Category:    req.Category,
		Tags:        req.Tags,
		MinPrice:    req.MinPrice,
		MaxPrice:    req.MaxPrice,
		InStockOnly: req.InStockOnly,
		SortBy:      req.SortBy,
		SortDesc:    req.SortDesc,
		SearchTerm:  req.SearchTerm,
	}

	// Call business logic
	products, total, err := s.productService.ListProducts(params)
	if err != nil {
		s.logger.Error("Failed to list products", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	// Map domain models to protobuf response
	protoProducts := make([]*pb.Product, len(products))
	for i, product := range products {
		protoProducts[i] = domainToProtoProduct(product)
	}

	return &pb.ListProductsResponse{
		Products:   protoProducts,
		Total:      int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int32((total + int(req.PageSize) - 1) / int(req.PageSize)), // Calculate total pages
	}, nil
}

// UpdateInventory implements the UpdateInventory RPC method
func (s *ProductServer) UpdateInventory(ctx context.Context, req *pb.UpdateInventoryRequest) (*pb.UpdateInventoryResponse, error) {
	s.logger.Info("gRPC UpdateInventory called",
		"productID", req.ProductId,
		"quantityChange", req.QuantityChange,
		"operationType", req.OperationType)

	// Call business logic
	updatedInventory, err := s.productService.UpdateInventory(
		req.ProductId,
		int(req.QuantityChange),
		req.OperationId,
		req.OperationType,
	)
	if err != nil {
		s.logger.Error("Failed to update inventory", "productID", req.ProductId, "error", err)
		return &pb.UpdateInventoryResponse{
			Success: false,
			Message: err.Error(),
		}, status.Errorf(codes.Internal, "failed to update inventory: %v", err)
	}

	// Map domain model to protobuf response
	return &pb.UpdateInventoryResponse{
		Success: true,
		UpdatedInventory: &pb.InventoryInfo{
			Quantity: int32(updatedInventory.Quantity),
			Sku:      updatedInventory.SKU,
			InStock:  updatedInventory.InStock,
			Reserved: int32(updatedInventory.Reserved),
		},
		Message: "Inventory updated successfully",
	}, nil
}

// CheckStock implements the CheckStock RPC method
func (s *ProductServer) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	s.logger.Info("gRPC CheckStock called", "productID", req.ProductId, "quantity", req.Quantity)

	// Call business logic
	available, currentStock, err := s.productService.CheckStock(req.ProductId, int(req.Quantity))
	if err != nil {
		s.logger.Error("Failed to check stock", "productID", req.ProductId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to check stock: %v", err)
	}

	return &pb.CheckStockResponse{
		Available:    available,
		CurrentStock: int32(currentStock),
	}, nil
}

// WatchInventory implements the WatchInventory RPC method
func (s *ProductServer) WatchInventory(req *pb.WatchInventoryRequest, stream pb.ProductService_WatchInventoryServer) error {
	s.logger.Info("gRPC WatchInventory called", "productIDs", req.ProductIds, "threshold", req.Threshold)

	// In a real implementation, we would:
	// 1. Set up a database watch/change stream or subscribe to a message queue
	// 2. Monitor inventory changes for the requested products
	// 3. Send updates to the client when stock changes

	// For this implementation, we'll simulate periodic inventory updates
	// This is just a placeholder for the real implementation

	// Continue sending updates until the client disconnects
	for i := 0; i < 5; i++ {
		// Check if client has cancelled the request
		if stream.Context().Err() != nil {
			return status.Errorf(codes.Canceled, "client cancelled request")
		}

		// Simulate an inventory update
		if len(req.ProductIds) > 0 {
			update := &pb.InventoryUpdate{
				ProductId:   req.ProductIds[0],
				ProductName: "Sample Product",
				Inventory: &pb.InventoryInfo{
					Quantity: 100 - int32(i*10),
					Sku:      "SKU123",
					InStock:  true,
					Reserved: 0,
				},
				Timestamp: time.Now().Unix(),
			}

			if err := stream.Send(update); err != nil {
				s.logger.Error("Failed to send inventory update", "error", err)
				return status.Errorf(codes.Internal, "failed to send update: %v", err)
			}
		}

		// Simulate delay between updates
		time.Sleep(1 * time.Second)
	}

	return nil
}

// Helper function to convert domain Product to proto Product
func domainToProtoProduct(product *domain.Product) *pb.Product {
	return &pb.Product{
		Id:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		ImageUrls:   product.ImageURLs,
		Category:    product.Category,
		Inventory: &pb.InventoryInfo{
			Quantity: int32(product.Inventory.Quantity),
			Sku:      product.Inventory.SKU,
			InStock:  product.Inventory.InStock,
			Reserved: int32(product.Inventory.Reserved),
		},
		Tags:       product.Tags,
		Attributes: product.Attributes,
		Active:     product.Active,
		CreatedAt:  product.CreatedAt.Unix(),
		UpdatedAt:  product.UpdatedAt.Unix(),
	}
}
