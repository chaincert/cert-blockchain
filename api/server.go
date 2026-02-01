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
	ChainID         string

	// certd tx signing/broadcast config (used by POST create endpoints)
	TxFrom           string
	TxKeyringBackend string
	TxHome           string
	TxNode           string
	TxGas            string
	TxFees           string
	TxGasPrices      string
	TxBroadcastMode  string
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
		ChainID:         "cert-testnet-1",

		TxFrom:           "validator",
		TxKeyringBackend: "test",
		TxHome:           "/root/.certd",
		TxNode:           "tcp://localhost:26657",
		TxGas:            "200000",
		TxFees:           "10000ucert",
		TxGasPrices:      "",
		TxBroadcastMode:  "block",
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
	api.HandleFunc("/staking/validators", s.handleGetValidators).Methods("GET")
	api.HandleFunc("/staking/validators/{validator_address}", s.handleGetValidator).Methods("GET")
	api.HandleFunc("/staking/params", s.handleGetStakingParams).Methods("GET")
	api.HandleFunc("/staking/delegate", s.handleDelegate).Methods("POST")
	api.HandleFunc("/staking/undelegate", s.handleUndelegate).Methods("POST")
	api.HandleFunc("/staking/rewards/{address}", s.handleGetRewards).Methods("GET")

	// User dashboard summary (aggregates wallet + staking + attestations)
	api.HandleFunc("/dashboard/{address}", s.handleGetDashboard).Methods("GET")

	// CertID Profile endpoints (Per CertID Section 2.2)
	api.HandleFunc("/profile/{address}", s.handleGetProfile).Methods("GET")
	api.HandleFunc("/profile", s.handleUpdateProfile).Methods("POST")
	api.HandleFunc("/profile/verify-social", s.handleVerifySocial).Methods("POST")
	api.HandleFunc("/profile/credentials", s.handleAddCredential).Methods("POST")
	api.HandleFunc("/profile/credentials/{id}", s.handleRemoveCredential).Methods("DELETE")

	// CertID Identity Resolution (Per Cert ID Evolution spec)
	api.HandleFunc("/identity/{address}", s.handleGetFullIdentity).Methods("GET")
	api.HandleFunc("/identity/{address}/badges", s.handleGetBadges).Methods("GET")
	api.HandleFunc("/identity/{address}/trust-score", s.handleGetTrustScore).Methods("GET")
	api.HandleFunc("/identity/resolve/{handle}", s.handleResolveHandle).Methods("GET")

	// CertID Verifiable Credential (VC) endpoints
	api.HandleFunc("/certid/vc/verify", s.handleVerifyCertIDVC).Methods("POST")

	// KYC Verification endpoints (Didit.me integration)
	api.HandleFunc("/kyc/start", s.requireAuth(s.handleStartKYC)).Methods("POST", "OPTIONS")
	api.HandleFunc("/kyc/status", s.requireAuth(s.handleGetKYCStatus)).Methods("GET", "OPTIONS")
	api.HandleFunc("/kyc/session/{sessionId}", s.requireAuth(s.handleGetKYCSession)).Methods("GET", "OPTIONS")
	api.HandleFunc("/kyc/webhook", s.handleKYCWebhook).Methods("POST") // No auth - verified by signature

	// Social Verification endpoints
	api.HandleFunc("/social/generate", s.requireAuth(s.handleSocialGenerate)).Methods("POST", "OPTIONS")
	api.HandleFunc("/social/verify", s.requireAuth(s.handleSocialVerify)).Methods("POST", "OPTIONS")
	api.HandleFunc("/social/{address}", s.handleSocialStatus).Methods("GET")

	// API Key Management endpoints
	api.HandleFunc("/api-keys", s.requireAuth(s.handleCreateAPIKey)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api-keys", s.requireAuth(s.handleListAPIKeys)).Methods("GET")
	api.HandleFunc("/api-keys/{keyId}", s.requireAuth(s.handleRevokeAPIKey)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/api-keys/{keyId}/usage", s.requireAuth(s.handleGetAPIKeyUsage)).Methods("GET")
	api.HandleFunc("/api-keys/tiers", s.handleGetAPITiers).Methods("GET")

	// Sybil Resistance API (Trust Score Validation)
	api.HandleFunc("/sybil/check/{address}", s.handleSybilCheck).Methods("GET")
	api.HandleFunc("/sybil/batch", s.handleSybilBatchCheck).Methods("POST", "OPTIONS")
	api.HandleFunc("/sybil/history/{address}", s.handleGetTrustScoreHistory).Methods("GET")

	// DID:web Support (W3C Decentralized Identifiers)
	api.HandleFunc("/.well-known/did.json", s.handleGetWellKnownDID).Methods("GET")
	api.HandleFunc("/identity/{address}/did.json", s.handleGetDIDDocument).Methods("GET")
	api.HandleFunc("/identity/{address}/presentation", s.handleGetDIDVerifiablePresentation).Methods("GET")
	api.HandleFunc("/identity/{address}/did/export", s.handleExportDIDtoJSON).Methods("GET")
	api.HandleFunc("/did/resolve", s.handleResolveDID).Methods("GET")

	// Statistics
	api.HandleFunc("/stats", s.handleGetStats).Methods("GET")

	// Explorer endpoints (Block Explorer)
	api.HandleFunc("/explorer/tx/{hash}", s.handleGetTransaction).Methods("GET")
	api.HandleFunc("/explorer/block/{height}", s.handleGetBlock).Methods("GET")
	api.HandleFunc("/explorer/address/{address}", s.handleGetAddress).Methods("GET")
	api.HandleFunc("/explorer/address/{address}/transactions", s.handleGetAddressTransactions).Methods("GET")
	api.HandleFunc("/explorer/transactions", s.handleGetRecentTransactions).Methods("GET")
	api.HandleFunc("/explorer/verify/{hash}", s.handleVerifyDocument).Methods("GET")
	api.HandleFunc("/explorer/stats", s.handleGetExplorerStats).Methods("GET")
	api.HandleFunc("/explorer/search", s.handleSearchExplorer).Methods("GET")

	// Faucet endpoint (testnet only)
	api.HandleFunc("/faucet", s.handleFaucet).Methods("POST", "OPTIONS")

	// Governance endpoints
	api.HandleFunc("/governance/proposals", s.handleGetAllProposals).Methods("GET")
	api.HandleFunc("/governance/proposals", s.handleCreateProposal).Methods("POST")
	api.HandleFunc("/governance/proposals/{proposal_id}", s.handleGetProposal).Methods("GET")
	api.HandleFunc("/governance/proposals/{proposal_id}/tally", s.handleGetProposalTally).Methods("GET")
	api.HandleFunc("/governance/proposals/{proposal_id}/votes", s.handleGetVotes).Methods("GET")
	api.HandleFunc("/governance/proposals/{proposal_id}/vote", s.handleVoteOnProposal).Methods("POST")
	api.HandleFunc("/governance/params", s.handleGetGovParams).Methods("GET")

	// Additional staking endpoints
	api.HandleFunc("/staking/redelegate", s.handleRedelegate).Methods("POST")
	api.HandleFunc("/staking/claim-rewards", s.handleClaimRewards).Methods("POST")

	// Developer API Key Management
	api.HandleFunc("/developer/keys", s.handleGetApiKeys).Methods("GET")
	api.HandleFunc("/developer/keys", s.handleCreateApiKey).Methods("POST")
	api.HandleFunc("/developer/keys/{keyId}", s.handleDeleteApiKey).Methods("DELETE")
	api.HandleFunc("/developer/usage", s.handleGetApiUsage).Methods("GET")

	// Bridge endpoints (Per Whitepaper Section 13)
	api.HandleFunc("/bridge/chains", s.handleGetSupportedChains).Methods("GET")
	api.HandleFunc("/bridge/fees", s.handleGetBridgeFees).Methods("GET")
	api.HandleFunc("/bridge/lock", s.handleLockTokens).Methods("POST")
	api.HandleFunc("/bridge/transfer/{transfer_id}", s.handleGetTransferStatus).Methods("GET")
	api.HandleFunc("/bridge/transfer/{transfer_id}/confirm", s.handleConfirmTransfer).Methods("POST")
	api.HandleFunc("/bridge/history/{address}", s.handleGetTransferHistory).Methods("GET")
	api.HandleFunc("/bridge/stats", s.handleGetBridgeStats).Methods("GET")

	// Enterprise Contact (Sales inquiries)
	api.HandleFunc("/enterprise/contact", s.handleEnterpriseContact).Methods("POST", "OPTIONS")

	// Discourse SSO (Community Forum Integration)
	s.RegisterDiscourseRoutes(api)
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With", "X-API-Key"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	s.router.Use(c.Handler)
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.recoveryMiddleware)
	s.router.Use(s.apiKeyMiddleware) // Validate API keys and track usage
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
