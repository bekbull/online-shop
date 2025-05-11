# User Service

This microservice manages user accounts for the Online Shop application. It provides both gRPC and RESTful HTTP APIs for user management operations.

## Features

- User CRUD operations (Create, Read, Update, Delete)
- Role-based user management
- Secure password hashing
- PostgreSQL database backend
- RESTful HTTP API with versioning
- gRPC API for inter-service communication

## Technology Stack

- Go 1.24+
- PostgreSQL
- gRPC/Protocol Buffers
- RESTful API (Chi router)
- Docker/Docker Compose

## API Endpoints

### RESTful API (HTTP)

All routes are prefixed with `/v1`.

- `GET /users` - List users (with pagination and filtering)
- `POST /users` - Create a new user
- `GET /users/{id}` - Get a specific user
- `PUT /users/{id}` - Update a user
- `DELETE /users/{id}` - Delete a user

### gRPC API

- `CreateUser` - Create a new user
- `GetUser` - Get a user by ID
- `GetUserByEmail` - Get a user by email address
- `UpdateUser` - Update an existing user
- `DeleteUser` - Delete a user
- `ListUsers` - List users with pagination and filtering

## Setup

### Prerequisites

- Go 1.24+
- Docker and Docker Compose (for local development)
- PostgreSQL (if running without Docker)

### Environment Variables

The service is configured using the following environment variables:

- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL user (default: postgres)
- `DB_PASSWORD` - PostgreSQL password (default: postgres)
- `DB_NAME` - PostgreSQL database name (default: users)
- `HTTP_PORT` - HTTP server port (default: 8081)
- `GRPC_PORT` - gRPC server port (default: 9091)

### Running Locally (with Docker)

```bash
docker-compose up
```

This will start both the PostgreSQL database and the User Service.

### Running Locally (without Docker)

1. Make sure PostgreSQL is running and accessible
2. Set the environment variables as needed
3. Run the service:

```bash
go run cmd/server/main.go
```

### Building

```bash
go build -o userservice ./cmd/server
```

## Development

### Regenerating Protocol Buffers

If you make changes to the `.proto` files, regenerate the code with:

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/user.proto
``` 