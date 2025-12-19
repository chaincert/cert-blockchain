// Package api provides the REST API server for CERT Blockchain
// Per Whitepaper Section 8 - API Specifications
package api

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/chaincertify/certd/api/database"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	router     *mux.Router
	httpServer *http.Server
	logger     *zap.Logger
	config     *Config
	db         *database.DB
}

// Config holds API server configuration
type Config struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
	JWTSecret       []byte
	DatabaseURL     string
	IPFSGateway     string
	ChainRPCURL     string
}

// DefaultConfig returns default API configuration
func DefaultConfig() *Config {
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)
	return &Config{
		Host:            "0.0.0.0",
		Port:            "3000",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		AllowedOrigins:  []string{"*"},
		JWTSecret:       secret,
		IPFSGateway:     "https://ipfs.c3rt.org",
		ChainRPCURL:     "http://localhost:26657",
	}
}

// NewServer creates a new API server instance
func NewServer(config *Config, logger *zap.Logger) *Server {
	router := mux.NewRouter()

	var dbConn *database.DB
	if config.DatabaseURL != "" {
		if d, err := database.NewFromURL(config.DatabaseURL, logger); err != nil {
			logger.Warn("Database disabled (failed to connect)", zap.Error(err))
		} else {
			dbConn = d
		}
	}

	s := &Server{
		router: router,
		logger: logger,
		config: config,
		db:     dbConn,
	}

	s.setupRoutes()
	s.setupMiddleware()

	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API version prefix
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Auth endpoints (EIP-191 challenge/response -> JWT)
	api.HandleFunc("/auth/challenge", s.handleAuthChallenge).Methods("GET")
	api.HandleFunc("/auth/verify", s.handleAuthVerify).Methods("POST")

	// Encrypted Attestation endpoints (Per Whitepaper Section 8)
	api.HandleFunc("/encrypted-attestations", s.handleCreateEncryptedAttestation).Methods("POST")
	api.HandleFunc("/encrypted-attestations/{uid}", s.handleGetEncryptedAttestation).Methods("GET")
	api.HandleFunc("/encrypted-attestations/{uid}/retrieve", s.handleRetrieveEncryptedAttestation).Methods("POST")
	api.HandleFunc("/encrypted-attestations/{uid}/revoke", s.handleRevokeEncryptedAttestation).Methods("POST")

	// Schema endpoints
	api.HandleFunc("/schemas", s.handleCreateSchema).Methods("POST")
	api.HandleFunc("/schemas/{uid}", s.handleGetSchema).Methods("GET")

	// Public attestation endpoints
	api.HandleFunc("/attestations", s.handleCreateAttestation).Methods("POST")
	api.HandleFunc("/attestations/{uid}", s.handleGetAttestation).Methods("GET")
	api.HandleFunc("/attestations/by-attester/{address}", s.handleGetAttestationsByAttester).Methods("GET")
	api.HandleFunc("/attestations/by-recipient/{address}", s.handleGetAttestationsByRecipient).Methods("GET")

	// Wallet + staking (testnet UX)
	api.HandleFunc("/wallet/{address}/balance", s.handleGetWalletBalance).Methods("GET")
	api.HandleFunc("/staking/delegations/{address}", s.handleGetStakingDelegations).Methods("GET")
	api.HandleFunc("/staking/summary/{address}", s.handleGetStakingSummary).Methods("GET")

	// User dashboard summary (aggregates wallet + staking + attestations)
	api.HandleFunc("/dashboard/{address}", s.handleGetDashboard).Methods("GET")

	// CertID Profile endpoints (Per CertID Section 2.2)
	api.HandleFunc("/profile/{address}", s.handleGetProfile).Methods("GET")
	api.HandleFunc("/profile", s.handleUpdateProfile).Methods("POST")
	api.HandleFunc("/profile/verify-social", s.handleVerifySocial).Methods("POST")
	api.HandleFunc("/profile/credentials", s.handleAddCredential).Methods("POST")
	api.HandleFunc("/profile/credentials/{id}", s.handleRemoveCredential).Methods("DELETE")

	// CertID Verifiable Credential (VC) endpoints
	api.HandleFunc("/certid/vc/verify", s.handleVerifyCertIDVC).Methods("POST")

	// Statistics
	api.HandleFunc("/stats", s.handleGetStats).Methods("GET")

	// Faucet endpoint (testnet only)
	api.HandleFunc("/faucet", s.handleFaucet).Methods("POST", "OPTIONS")

	// Governance endpoints
	api.HandleFunc("/governance/proposals", s.handleGetProposals).Methods("GET")
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	s.router.Use(c.Handler)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.recoveryMiddleware)
}

// Start begins serving the API
func (s *Server) Start() error {
	addr := s.config.Host + ":" + s.config.Port

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Info("Starting API server", zap.String("address", addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down API server")
	err := s.httpServer.Shutdown(ctx)
	if s.db != nil {
		if cerr := s.db.Close(); cerr != nil {
			s.logger.Warn("Failed to close database", zap.Error(cerr))
		}
	}
	return err
}
