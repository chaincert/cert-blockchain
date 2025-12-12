// CERT API Server Entry Point
// Per Whitepaper Section 8: API Specifications
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/chaincertify/certd/api"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Load configuration from environment
	config := loadConfig()

	// Create API server
	server := api.NewServer(config, logger)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("CERT API Server started",
		zap.String("host", config.Host),
		zap.String("port", config.Port),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// loadConfig loads configuration from environment variables
func loadConfig() *api.Config {
	config := api.DefaultConfig()

	if host := os.Getenv("API_HOST"); host != "" {
		config.Host = host
	}
	if port := os.Getenv("API_PORT"); port != "" {
		config.Port = port
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}
	if ipfsGateway := os.Getenv("IPFS_GATEWAY"); ipfsGateway != "" {
		config.IPFSGateway = ipfsGateway
	}
	if chainRPC := os.Getenv("CHAIN_RPC_URL"); chainRPC != "" {
		config.ChainRPCURL = chainRPC
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = []byte(jwtSecret)
	}

	// Parse timeout values
	if timeout := os.Getenv("READ_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.ReadTimeout = d
		}
	}
	if timeout := os.Getenv("WRITE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.WriteTimeout = d
		}
	}

	return config
}

