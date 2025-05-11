package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/bekbull/online-shop/services/product-service/config"
	"github.com/bekbull/online-shop/services/product-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProductRepository implements the domain.ProductRepository interface with MongoDB
type ProductRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
	config     *config.MongoDBConfig
}

// New creates a new ProductRepository with MongoDB
func New(client *mongo.Client, cfg *config.MongoDBConfig) *ProductRepository {
	collection := client.Database(cfg.Database).Collection(cfg.Collection)
	return &ProductRepository{
		client:     client,
		collection: collection,
		config:     cfg,
	}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.WriteTimeout)
	defer cancel()

	// Ensure ID and timestamps are set
	if product.ID.IsZero() {
		product.ID = primitive.NewObjectID()
	}
	if product.CreatedAt.IsZero() {
		product.CreatedAt = time.Now()
	}
	product.UpdatedAt = time.Now()

	// Ensure inventory.InStock is set correctly
	product.Inventory.InStock = product.Inventory.Quantity > 0

	_, err := r.collection.InsertOne(ctx, product)
	return err
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(id string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.ReadTimeout)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var product domain.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return &product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.WriteTimeout)
	defer cancel()

	// Update timestamps
	product.UpdatedAt = time.Now()

	// Ensure inventory.InStock is set correctly
	product.Inventory.InStock = product.Inventory.Quantity > 0

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": product.ID}, product)
	return err
}

// Delete removes a product by its ID
func (r *ProductRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.WriteTimeout)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("product not found")
	}

	return nil
}

// List retrieves products based on filter parameters
func (r *ProductRepository) List(params domain.ListProductsParams) ([]*domain.Product, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.ReadTimeout)
	defer cancel()

	// Build filter
	filter := bson.M{}

	// Add category filter if provided
	if params.Category != "" {
		filter["category"] = params.Category
	}

	// Add tags filter if provided
	if len(params.Tags) > 0 {
		filter["tags"] = bson.M{"$all": params.Tags}
	}

	// Add price range filter if provided
	if params.MinPrice > 0 || params.MaxPrice > 0 {
		priceFilter := bson.M{}
		if params.MinPrice > 0 {
			priceFilter["$gte"] = params.MinPrice
		}
		if params.MaxPrice > 0 {
			priceFilter["$lte"] = params.MaxPrice
		}
		filter["price"] = priceFilter
	}

	// Add in-stock filter if requested
	if params.InStockOnly {
		filter["inventory.in_stock"] = true
	}

	// Add text search if provided
	if params.SearchTerm != "" {
		filter["$text"] = bson.M{"$search": params.SearchTerm}
	}

	// Count total matching documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set up pagination
	findOptions := options.Find()
	if params.PageSize > 0 {
		findOptions.SetLimit(int64(params.PageSize))
		findOptions.SetSkip(int64(params.Page * params.PageSize))
	}

	// Set up sorting
	if params.SortBy != "" {
		sortDirection := 1
		if params.SortDesc {
			sortDirection = -1
		}
		findOptions.SetSort(bson.D{{Key: params.SortBy, Value: sortDirection}})
	} else {
		// Default sort by creation date descending
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	// Execute query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode results
	var products []*domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}

	return products, int(total), nil
}

// UpdateInventory updates a product's inventory
func (r *ProductRepository) UpdateInventory(productID string, quantityChange int, operationID, operationType string) (*domain.InventoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.WriteTimeout)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, err
	}

	// Use a session with transaction to ensure atomicity
	session, err := r.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var updatedInventory *domain.InventoryInfo
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// Start transaction
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// Check for duplicate operation if operationID is provided (idempotency)
		if operationID != "" {
			// Create a collection for inventory operations if we need to track them
			opCollection := r.client.Database(r.config.Database).Collection("inventory_operations")
			
			// Check if this operation already exists
			var existingOp domain.InventoryOperation
			err := opCollection.FindOne(sc, bson.M{"operation_id": operationID}).Decode(&existingOp)
			if err == nil {
				// Operation already processed
				// Fetch current inventory and return
				var product domain.Product
				err = r.collection.FindOne(sc, bson.M{"_id": objID}).Decode(&product)
				if err != nil {
					return err
				}
				updatedInventory = &product.Inventory
				return nil
			} else if err != mongo.ErrNoDocuments {
				// Unexpected error
				return err
			}
			
			// Record the operation
			_, err = opCollection.InsertOne(sc, domain.InventoryOperation{
				ProductID:      productID,
				QuantityChange: quantityChange,
				OperationID:    operationID,
				OperationType:  operationType,
				Timestamp:      time.Now(),
			})
			if err != nil {
				return err
			}
		}

		// Update the product's inventory
		update := bson.M{
			"$inc": bson.M{"inventory.quantity": quantityChange},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		}

		// Execute update and get the updated document
		var product domain.Product
		err = r.collection.FindOneAndUpdate(
			sc, 
			bson.M{"_id": objID}, 
			update,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&product)
		
		if err != nil {
			return err
		}

		// Update the in_stock status based on the new quantity
		inStock := product.Inventory.Quantity > 0
		if inStock != product.Inventory.InStock {
			_, err = r.collection.UpdateOne(
				sc,
				bson.M{"_id": objID},
				bson.M{"$set": bson.M{"inventory.in_stock": inStock}},
			)
			if err != nil {
				return err
			}
			product.Inventory.InStock = inStock
		}

		updatedInventory = &product.Inventory

		// Commit the transaction
		return session.CommitTransaction(sc)
	})

	if err != nil {
		return nil, err
	}

	return updatedInventory, nil
}

// CheckStock checks if a product has sufficient stock
func (r *ProductRepository) CheckStock(productID string, quantity int) (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.ReadTimeout)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return false, 0, err
	}

	var product domain.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, 0, errors.New("product not found")
		}
		return false, 0, err
	}

	// Available quantity is (total - reserved)
	availableQuantity := product.Inventory.Quantity - product.Inventory.Reserved
	return availableQuantity >= quantity, availableQuantity, nil
} 