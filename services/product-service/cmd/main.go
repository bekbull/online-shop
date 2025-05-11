package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bekbull/online-shop/proto/product"
	"github.com/bekbull/online-shop/services/product-service/config"
	grpcHandler "github.com/bekbull/online-shop/services/product-service/internal/api/grpc"
	restHandler "github.com/bekbull/online-shop/services/product-service/internal/api/rest"
	"github.com/bekbull/online-shop/services/product-service/internal/repository/mongodb"
	"github.com/bekbull/online-shop/services/product-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize logger
	logger := setupLogger()
	logger.Info("Starting Product Service")

	// Load configuration
	cfg := config.Load()
	logger.Info("Configuration loaded")

	// Connect to MongoDB
	mongoClient, err := connectToMongoDB(cfg.MongoDB)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Error("Failed to disconnect from MongoDB", "error", err)
		}
	}()
	logger.Info("Connected to MongoDB")

	// Create repository
	productRepo := mongodb.New(mongoClient, &cfg.MongoDB)

	// Create service
	productService := service.New(productRepo, logger)

	// Setup HTTP server
	router := setupHTTPServer(cfg, productService, logger)

	// Setup gRPC server
	grpcServer := setupGRPCServer(cfg, productService, logger)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start servers in goroutines
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		logger.Info("Starting gRPC server", "port", cfg.GRPCPort)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			logger.Error("Failed to listen for gRPC", "error", err)
			os.Exit(1)
		}
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Handle graceful shutdown
	handleGracefulShutdown(httpServer, grpcServer, logger)
}

func setupLogger() *slog.Logger {
	// Create a simple logger for now
	// In a real application, we would configure this based on environment
	// and use structured logging with proper levels and output formats
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func connectToMongoDB(cfg config.MongoDBConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
	defer cancel()

	// Create MongoDB client
	clientOptions := options.Client().
		ApplyURI(cfg.ConnectionString()).
		SetMaxPoolSize(cfg.MaxPoolSize)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func setupHTTPServer(cfg *config.Config, productService *service.ProductService, logger *slog.Logger) *chi.Mux {
	// Create router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	// Create REST handler
	productHandler := restHandler.NewProductHandler(productService, logger)

	// Register routes
	productHandler.RegisterRoutes(router)

	// Add health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add metrics endpoint (in a real application, we'd configure Prometheus here)
	if cfg.Metrics.Enabled {
		router.Get(cfg.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Metrics would be here"))
		})
	}

	return router
}

func setupGRPCServer(cfg *config.Config, productService *service.ProductService, logger *slog.Logger) *grpc.Server {
	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create gRPC handler
	productServer := grpcHandler.New(productService, logger)

	// Register gRPC services
	product.RegisterProductServiceServer(grpcServer, productServer)

	// Enable reflection for development tools
	if cfg.Env != "production" {
		reflection.Register(grpcServer)
	}

	return grpcServer
}

func handleGracefulShutdown(httpServer *http.Server, grpcServer *grpc.Server, logger *slog.Logger) {
	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	logger.Info("Received shutdown signal", "signal", sig)

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	logger.Info("Shutting down HTTP server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	}

	// Shutdown gRPC server
	logger.Info("Shutting down gRPC server")
	grpcServer.GracefulStop()

	logger.Info("Shutdown completed")
}
