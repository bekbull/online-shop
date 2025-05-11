# Product Service

The Product Service is a microservice for managing product catalog and inventory in the online shop.

## Features

- Product CRUD operations
- Inventory management
- Stock availability checking
- Real-time inventory updates with gRPC streaming
- Extensive filtering and search capabilities
- Idempotent inventory operations

## Architecture

- **Domain-Driven Design**: Clear separation of concerns with domain models, repositories, and services
- **Hexagonal Architecture**: Core business logic is independent of external frameworks and databases
- **REST API**: Public HTTP endpoints for external clients
- **gRPC API**: Internal communication for other microservices
- **MongoDB**: Document storage for flexible product data

## Implementation Details

### API Endpoints

#### RESTful API

- **Create Product**: `POST /v1/products`
- **Get Product**: `GET /v1/products/{id}`
- **Update Product**: `PUT /v1/products/{id}`
- **Delete Product**: `DELETE /v1/products/{id}`
- **List Products**: `GET /v1/products?page=0&page_size=20`
- **Update Inventory**: `POST /v1/products/{id}/inventory`
- **Check Stock**: `GET /v1/products/{id}/stock?quantity=5`

#### gRPC Service

The service implements the `ProductService` interface defined in `proto/product/product.proto`:

- `CreateProduct`
- `GetProduct`
- `UpdateProduct`
- `DeleteProduct`
- `ListProducts`
- `UpdateInventory`
- `CheckStock`
- `WatchInventory` (streaming)

### Configuration

The service is configured via environment variables:

- `ENV`: Environment (development, staging, production)
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: MongoDB database name
- `MONGODB_COLLECTION`: MongoDB collection name
- `MONGODB_USERNAME`: MongoDB username
- `MONGODB_PASSWORD`: MongoDB password
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `LOG_JSON`: Whether to output logs as JSON
- `LOG_PRETTY`: Whether to format JSON logs
- `GRPC_PORT`: gRPC server port
- `HTTP_PORT`: HTTP server port
- `METRICS_ENABLED`: Whether to enable metrics endpoints
- `METRICS_PATH`: Path for metrics endpoint
- `TRACING_ENABLED`: Whether to enable distributed tracing

### Testing

The service includes both unit tests and benchmarks:

- Tests for business logic in the service layer
- Benchmarks for performance-critical operations

Run tests with:

```sh
go test ./...
```

Run benchmarks with:

```sh
go test -bench=. ./internal/service
```

### Performance

Based on benchmark tests, the service can handle:

- Product creation in under 10ms
- Product listing in under 10ms
- Inventory updates in under 10ms
- Stock checks in under 10ms

### Deployment

The service can be deployed as a Docker container:

```sh
docker build -t product-service -f Dockerfile .
docker run -p 8080:8080 -p 50051:50051 product-service
```

Or using docker-compose:

```sh
docker-compose -f ../../docker-compose.dev.yml up
``` 