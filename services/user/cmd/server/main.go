package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bekbull/online-shop/services/user/api/proto"
	"github.com/bekbull/online-shop/services/user/internal/handler"
	"github.com/bekbull/online-shop/services/user/internal/repository"
	"github.com/bekbull/online-shop/services/user/internal/service"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "[USER-SERVICE] ", log.LstdFlags)
	logger.Println("Starting user service...")

	// Load configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "users")
	httpPort := getEnv("HTTP_PORT", "8081")
	grpcPort := getEnv("GRPC_PORT", "9091")

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set up connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create repository
	repo := repository.NewPostgresRepository(db)

	// Initialize database schema
	if err := repo.InitDB(); err != nil {
		logger.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Create service
	userService := service.NewUserService(repo)

	// Create HTTP server
	httpServer := handler.NewHTTPServer(userService)
	httpSrv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpServer.Router(),
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	userGrpcServer := handler.NewGRPCServer(userService)
	proto.RegisterUserServiceServer(grpcServer, userGrpcServer)
	reflection.Register(grpcServer) // Enable reflection for debugging

	// Start HTTP server in a goroutine
	go func() {
		logger.Printf("HTTP server listening on port %s", httpPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server in a goroutine
	go func() {
		listener, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
		}
		logger.Printf("gRPC server listening on port %s", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down servers...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logger.Println("Servers stopped")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
