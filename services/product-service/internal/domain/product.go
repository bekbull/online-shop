package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a product in the catalog
type Product struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Description string                 `bson:"description" json:"description"`
	Price       float64                `bson:"price" json:"price"`
	ImageURLs   []string               `bson:"image_urls" json:"image_urls"`
	Category    string                 `bson:"category" json:"category"`
	Inventory   InventoryInfo          `bson:"inventory" json:"inventory"`
	Tags        []string               `bson:"tags" json:"tags"`
	Attributes  map[string]string      `bson:"attributes" json:"attributes"`
	Active      bool                   `bson:"active" json:"active"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// InventoryInfo contains product inventory details
type InventoryInfo struct {
	Quantity int    `bson:"quantity" json:"quantity"`
	SKU      string `bson:"sku" json:"sku"`
	InStock  bool   `bson:"in_stock" json:"in_stock"`
	Reserved int    `bson:"reserved" json:"reserved"`
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(product *Product) error
	GetByID(id string) (*Product, error)
	Update(product *Product) error
	Delete(id string) error
	List(params ListProductsParams) ([]*Product, int, error)
	UpdateInventory(productID string, quantityChange int, operationID, operationType string) (*InventoryInfo, error)
	CheckStock(productID string, quantity int) (bool, int, error)
}

// ListProductsParams defines the parameters for listing products
type ListProductsParams struct {
	Page        int
	PageSize    int
	Category    string
	Tags        []string
	MinPrice    float64
	MaxPrice    float64
	InStockOnly bool
	SortBy      string
	SortDesc    bool
	SearchTerm  string
}

// InventoryOperation represents a change to inventory
type InventoryOperation struct {
	ProductID     string    `bson:"product_id" json:"product_id"`
	QuantityChange int       `bson:"quantity_change" json:"quantity_change"`
	OperationID   string    `bson:"operation_id" json:"operation_id"`
	OperationType string    `bson:"operation_type" json:"operation_type"` // e.g., "purchase", "restock"
	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

// NewProduct creates a new product with default values
func NewProduct() *Product {
	return &Product{
		ID:         primitive.NewObjectID(),
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Attributes: make(map[string]string),
	}
} 