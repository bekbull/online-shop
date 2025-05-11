# Go Microservices Online Shop

A microservices-based online shop application implemented in Go.

## Architecture

This project implements a microservices architecture for an online shop with the following services:

- **Product Service**: Manages product catalog and inventory
- **User Service**: Handles user accounts and profiles
- **Order Service**: Manages the order process and history
- **Auth Service**: Provides centralized authentication

For detailed architecture information, see [Architecture Documentation](memory-bank/architecture.md).

## Technologies

- **Backend**: Go 1.24+
- **Communication**: gRPC with Protocol Buffers
- **API**: RESTful HTTP endpoints
- **Databases**:
  - MongoDB (Product Service)
  - PostgreSQL (User and Order Services)
  - Redis (Auth Service)
- **Deployment**: Docker and Kubernetes

## Prerequisites

- Go 1.24+
- Protocol Buffers compiler (protoc)
- Docker and Docker Compose
- Git

## Getting Started

### Clone the repository

```sh
git clone https://github.com/bekbull/online-shop.git
cd online-shop
```

### Install dependencies

```sh
go mod tidy
```

### Start the development environment

For local development with all required databases:

```sh
docker-compose up -d
```

For running the Product Service with MongoDB:

```sh
docker-compose -f docker-compose.dev.yml up -d
```

### Running the services locally

Each service can be run individually in development mode:

```sh
# Run the Product Service
go run services/product-service/cmd/main.go

# Run the User Service (when implemented)
go run services/user-service/cmd/main.go

# Run the Order Service (when implemented)
go run services/order-service/cmd/main.go

# Run the Auth Service (when implemented)
go run services/auth-service/cmd/main.go
```

### Testing

Run unit tests with:

```sh
go test ./...
```

Run benchmarks with:

```sh
go test -bench=. ./services/product-service/internal/service
```

## API Documentation

### Product Service

#### RESTful API Endpoints

- **Create Product**: `POST /v1/products`
- **Get Product**: `GET /v1/products/{id}`
- **Update Product**: `PUT /v1/products/{id}`
- **Delete Product**: `DELETE /v1/products/{id}`
- **List Products**: `GET /v1/products?page=0&page_size=20`
  - Query parameters:
    - `page`: Page number (default: 0)
    - `page_size`: Items per page (default: 20)
    - `category`: Filter by category
    - `tags`: Filter by tags (comma-separated)
    - `min_price`: Minimum price
    - `max_price`: Maximum price
    - `in_stock`: Whether to only show in-stock items (true/false)
    - `sort_by`: Field to sort by
    - `sort_desc`: Whether to sort in descending order (true/false)
    - `search`: Search term

- **Update Inventory**: `POST /v1/products/{id}/inventory`
- **Check Stock**: `GET /v1/products/{id}/stock?quantity=5`

#### gRPC Service

The Product Service also provides a gRPC interface defined in `proto/product/product.proto`.

## Development

### Project Structure

```
online-shop/
├── proto/                  # Protocol Buffer definitions
│   ├── product/            # Product service protobuf files
│   ├── user/               # User service protobuf files
│   ├── order/              # Order service protobuf files
│   └── auth/               # Auth service protobuf files
├── services/               # Microservices
│   ├── product-service/    # Product microservice
│   ├── user-service/       # User microservice
│   ├── order-service/      # Order microservice
│   └── auth-service/       # Auth microservice
├── docker-compose.yml      # Docker Compose for all databases
├── docker-compose.dev.yml  # Docker Compose for Product Service
└── memory-bank/            # Project documentation
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 