package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	Metrics  MetricsConfig
	Logging  LoggingConfig
	Tracing  TracingConfig
	GRPCPort int
	HTTPPort int
	Env      string
}

// ServerConfig holds HTTP and API server configuration
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI          string
	Database     string
	Collection   string
	Username     string
	Password     string
	MaxPoolSize  uint64
	ConnTimeout  time.Duration
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

// MetricsConfig holds configuration for metrics collection
type MetricsConfig struct {
	Enabled bool
	Path    string
}

// LoggingConfig holds configuration for logging
type LoggingConfig struct {
	Level  string
	JSON   bool
	Pretty bool
}

// TracingConfig holds configuration for distributed tracing
type TracingConfig struct {
	Enabled    bool
	ServiceName string
	Endpoint   string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		MongoDB: MongoDBConfig{
			URI:          getEnv("MONGODB_URI", "mongodb://root:root_password@mongodb:27017"),
			Database:     getEnv("MONGODB_DATABASE", "product_db"),
			Collection:   getEnv("MONGODB_COLLECTION", "products"),
			Username:     getEnv("MONGODB_USERNAME", "root"),
			Password:     getEnv("MONGODB_PASSWORD", "root_password"),
			MaxPoolSize:  getEnvUint64("MONGODB_MAX_POOL_SIZE", 100),
			ConnTimeout:  getEnvDuration("MONGODB_CONN_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("MONGODB_WRITE_TIMEOUT", 10*time.Second),
			ReadTimeout:  getEnvDuration("MONGODB_READ_TIMEOUT", 10*time.Second),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			JSON:   getEnvBool("LOG_JSON", true),
			Pretty: getEnvBool("LOG_PRETTY", false),
		},
		Tracing: TracingConfig{
			Enabled:    getEnvBool("TRACING_ENABLED", true),
			ServiceName: getEnv("TRACING_SERVICE_NAME", "product-service"),
			Endpoint:   getEnv("TRACING_ENDPOINT", "http://jaeger:14268/api/traces"),
		},
		GRPCPort: getEnvInt("GRPC_PORT", 50051),
		HTTPPort: getEnvInt("HTTP_PORT", 8080),
		Env:      getEnv("ENV", "development"),
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvUint64(key string, defaultValue uint64) uint64 {
	if value, exists := os.LookupEnv(key); exists {
		if uint64Value, err := strconv.ParseUint(value, 10, 64); err == nil {
			return uint64Value
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ConnectionString returns a formatted MongoDB connection string
func (c *MongoDBConfig) ConnectionString() string {
	if c.URI != "" {
		return c.URI
	}
	
	// Otherwise construct from components
	return fmt.Sprintf("mongodb://%s:%s@mongodb:27017/%s", 
		c.Username, c.Password, c.Database)
} 