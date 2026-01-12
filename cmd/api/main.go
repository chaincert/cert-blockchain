// CERT API Server Entry Point
// Per Whitepaper Section 8: API Specifications
package main

import (
	"context"
	"net/http"
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
			// `http.ErrServerClosed` is the expected error returned by `ListenAndServe`
			// when we call `Shutdown()`.
			if err == http.ErrServerClosed {
				logger.Info("API server closed")
				return
			}
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
	if chainID := os.Getenv("CERT_TX_CHAIN_ID"); chainID != "" {
		config.ChainID = chainID
	} else if chainID := os.Getenv("CERT_CHAIN_ID"); chainID != "" {
		// Back-compat / convenience: if a single chain-id env var is used, apply it to tx as well.
		config.ChainID = chainID
	}

	if v := os.Getenv("CERT_TX_FROM"); v != "" {
		config.TxFrom = v
	}
	if v := os.Getenv("CERT_TX_KEYRING_BACKEND"); v != "" {
		config.TxKeyringBackend = v
	}
	if v := os.Getenv("CERT_TX_HOME"); v != "" {
		config.TxHome = v
	}
	if v := os.Getenv("CERT_TX_NODE"); v != "" {
		config.TxNode = v
	}
	if v := os.Getenv("CERT_TX_GAS"); v != "" {
		config.TxGas = v
	}
	if v := os.Getenv("CERT_TX_FEES"); v != "" {
		config.TxFees = v
	}
	if v := os.Getenv("CERT_TX_GAS_PRICES"); v != "" {
		config.TxGasPrices = v
	}
	if v := os.Getenv("CERT_TX_BROADCAST_MODE"); v != "" {
		config.TxBroadcastMode = v
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
