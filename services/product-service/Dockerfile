# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service ./services/product-service/cmd

# Final stage
FROM alpine:latest

# Add necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/product-service .

# Set ownership
RUN chown -R appuser:appgroup /app

# Use the non-root user
USER appuser

# Expose ports
EXPOSE 8080 50051

# Command to run the application
CMD ["./product-service"] 