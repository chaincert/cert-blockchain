package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// BridgeTransfer represents a cross-chain transfer
type BridgeTransfer struct {
	TransferID      string    `json:"transfer_id"`
	Sender          string    `json:"sender"`
	Recipient       string    `json:"recipient"`
	Amount          string    `json:"amount"`
	SourceChainID   uint64    `json:"source_chain_id"`
	TargetChainID   uint64    `json:"target_chain_id"`
	Status          string    `json:"status"` // pending, confirmed, completed, failed
	TxHash          string    `json:"tx_hash,omitempty"`
	TargetTxHash    string    `json:"target_tx_hash,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Confirmations   int       `json:"confirmations"`
	RequiredConfirm int       `json:"required_confirmations"`
}

// SupportedChain represents a chain supported by the bridge
type SupportedChain struct {
	ChainID     uint64 `json:"chain_id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	RpcURL      string `json:"rpc_url,omitempty"`
	BridgeAddr  string `json:"bridge_address"`
	IsActive    bool   `json:"is_active"`
	MinAmount   string `json:"min_amount"`
	MaxAmount   string `json:"max_amount"`
	Fee         string `json:"fee_percent"`
}

// In-memory store for bridge transfers (use DB in production)
var (
	bridgeTransfers     = make(map[string]*BridgeTransfer)
	bridgeTransferMutex sync.RWMutex
)

// Supported chains configuration
var supportedChains = []SupportedChain{
	{ChainID: 951753, Name: "CERT Chain (EVM)", Symbol: "CERT", BridgeAddr: "0x...", IsActive: true, MinAmount: "1", MaxAmount: "1000000", Fee: "0.1"},
	{ChainID: 1, Name: "Ethereum", Symbol: "ETH", BridgeAddr: "0x...", IsActive: true, MinAmount: "1", MaxAmount: "1000000", Fee: "0.1"},
	{ChainID: 42161, Name: "Arbitrum One", Symbol: "ARB", BridgeAddr: "0x...", IsActive: true, MinAmount: "1", MaxAmount: "1000000", Fee: "0.05"},
	{ChainID: 137, Name: "Polygon", Symbol: "MATIC", BridgeAddr: "0x...", IsActive: false, MinAmount: "1", MaxAmount: "1000000", Fee: "0.05"},
}

// handleGetSupportedChains returns all supported chains for bridging
func (s *Server) handleGetSupportedChains(w http.ResponseWriter, r *http.Request) {
	activeChains := make([]SupportedChain, 0)
	for _, chain := range supportedChains {
		if chain.IsActive {
			activeChains = append(activeChains, chain)
		}
	}
	s.respondJSON(w, http.StatusOK, activeChains)
}

// handleGetBridgeFees returns fee estimation for a bridge transfer
func (s *Server) handleGetBridgeFees(w http.ResponseWriter, r *http.Request) {
	sourceChain := r.URL.Query().Get("source_chain")
	targetChain := r.URL.Query().Get("target_chain")
	amount := r.URL.Query().Get("amount")

	if sourceChain == "" || targetChain == "" || amount == "" {
		s.respondError(w, http.StatusBadRequest, "Missing required params: source_chain, target_chain, amount")
		return
	}

	// Calculate fees (simplified)
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		s.respondError(w, http.StatusBadRequest, "Invalid amount")
		return
	}

	// 0.1% fee
	feeRate := big.NewInt(1)
	feeDivisor := big.NewInt(1000)
	fee := new(big.Int).Div(new(big.Int).Mul(amountBig, feeRate), feeDivisor)

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"source_chain":   sourceChain,
		"target_chain":   targetChain,
		"amount":         amount,
		"fee":            fee.String(),
		"fee_percent":    "0.1",
		"estimated_time": "5-15 minutes",
		"gas_estimate":   "150000",
	})
}

// LockTokensRequest represents a request to lock tokens for bridging
type LockTokensRequest struct {
	Sender        string `json:"sender"`
	Recipient     string `json:"recipient"`
	Amount        string `json:"amount"`
	SourceChainID uint64 `json:"source_chain_id"`
	TargetChainID uint64 `json:"target_chain_id"`
}

// handleLockTokens initiates a bridge transfer by preparing a lock transaction
func (s *Server) handleLockTokens(w http.ResponseWriter, r *http.Request) {
	var req LockTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Sender == "" || req.Amount == "" || req.TargetChainID == 0 {
		s.respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	if req.Recipient == "" {
		req.Recipient = req.Sender
	}

	// Validate amount
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok || amount.Sign() <= 0 {
		s.respondError(w, http.StatusBadRequest, "Invalid amount")
		return
	}

	// Generate transfer ID
	transferID := fmt.Sprintf("0x%x", time.Now().UnixNano())

	// Create unsigned transaction for EVM bridge contract
	unsignedTx := map[string]interface{}{
		"to":   "0xBridgeContractAddress", // TODO: Use actual bridge address
		"data": fmt.Sprintf("lockTokens(%s,%d,%s)", req.Amount, req.TargetChainID, req.Recipient),
		"value": "0",
		"gas":   "200000",
	}

	// Store pending transfer
	transfer := &BridgeTransfer{
		TransferID:      transferID,
		Sender:          req.Sender,
		Recipient:       req.Recipient,
		Amount:          req.Amount,
		SourceChainID:   req.SourceChainID,
		TargetChainID:   req.TargetChainID,
		Status:          "pending",
		CreatedAt:       time.Now(),
		Confirmations:   0,
		RequiredConfirm: 12,
	}
	bridgeTransferMutex.Lock()
	bridgeTransfers[transferID] = transfer
	bridgeTransferMutex.Unlock()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"transfer_id":  transferID,
		"message":      "Sign and broadcast this transaction to lock tokens",
		"unsigned_tx":  unsignedTx,
		"source_chain": req.SourceChainID,
		"target_chain": req.TargetChainID,
	})
}

// handleGetTransferStatus returns the status of a bridge transfer
func (s *Server) handleGetTransferStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transferID := vars["transfer_id"]

	bridgeTransferMutex.RLock()
	transfer, exists := bridgeTransfers[transferID]
	bridgeTransferMutex.RUnlock()

	if !exists {
		s.respondError(w, http.StatusNotFound, "Transfer not found")
		return
	}

	s.respondJSON(w, http.StatusOK, transfer)
}

// handleGetTransferHistory returns bridge transfer history for an address
func (s *Server) handleGetTransferHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	bridgeTransferMutex.RLock()
	transfers := make([]*BridgeTransfer, 0)
	for _, t := range bridgeTransfers {
		if t.Sender == address || t.Recipient == address {
			transfers = append(transfers, t)
		}
	}
	bridgeTransferMutex.RUnlock()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":   address,
		"transfers": transfers,
		"total":     len(transfers),
	})
}

// handleConfirmTransfer updates transfer status (called by validator service)
func (s *Server) handleConfirmTransfer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transferID := vars["transfer_id"]

	var req struct {
		TxHash        string `json:"tx_hash"`
		Status        string `json:"status"`
		Confirmations int    `json:"confirmations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	bridgeTransferMutex.Lock()
	transfer, exists := bridgeTransfers[transferID]
	if exists {
		if req.TxHash != "" {
			transfer.TxHash = req.TxHash
		}
		if req.Status != "" {
			transfer.Status = req.Status
		}
		if req.Confirmations > 0 {
			transfer.Confirmations = req.Confirmations
		}
		if req.Status == "completed" {
			now := time.Now()
			transfer.CompletedAt = &now
		}
	}
	bridgeTransferMutex.Unlock()

	if !exists {
		s.respondError(w, http.StatusNotFound, "Transfer not found")
		return
	}

	s.respondJSON(w, http.StatusOK, transfer)
}

// handleGetBridgeStats returns overall bridge statistics
func (s *Server) handleGetBridgeStats(w http.ResponseWriter, r *http.Request) {
	bridgeTransferMutex.RLock()
	totalTransfers := len(bridgeTransfers)
	pendingCount := 0
	completedCount := 0
	totalVolume := big.NewInt(0)

	for _, t := range bridgeTransfers {
		if t.Status == "pending" || t.Status == "confirmed" {
			pendingCount++
		} else if t.Status == "completed" {
			completedCount++
		}
		if amt, ok := new(big.Int).SetString(t.Amount, 10); ok {
			totalVolume.Add(totalVolume, amt)
		}
	}
	bridgeTransferMutex.RUnlock()

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_transfers":     totalTransfers,
		"pending_transfers":   pendingCount,
		"completed_transfers": completedCount,
		"total_volume":        totalVolume.String(),
		"supported_chains":    len(supportedChains),
	})
}
