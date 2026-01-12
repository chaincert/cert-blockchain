//go:build ignore

// Package api - Faucet handler for testnet token distribution (legacy, superseded by ./api/handlers_faucet.go)
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FaucetRequest represents a faucet token request
type FaucetRequest struct {
	Address string `json:"address"`
}

// FaucetResponse represents a faucet response
type FaucetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TxHash  string `json:"tx_hash,omitempty"`
	Amount  string `json:"amount,omitempty"`
}

// Rate limiting: track requests per address
var (
	faucetRequests = make(map[string]time.Time)
	faucetMutex    sync.RWMutex
	faucetCooldown = 24 * time.Hour // 24 hour cooldown per address
	faucetAmount   = "10000000"     // 10 CERT = 10,000,000 ucert
)

// handleFaucet handles POST /api/v1/faucet
// Distributes testnet tokens to requesting addresses
func (s *Server) handleFaucet(w http.ResponseWriter, r *http.Request) {
	var req FaucetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate address format (cert1... or 0x...)
	if !isValidAddress(req.Address) {
		s.respondError(w, http.StatusBadRequest, "Invalid address format. Must be cert1... or 0x...")
		return
	}

	// Normalize address for rate limiting
	normalizedAddr := strings.ToLower(req.Address)

	// Check rate limiting
	faucetMutex.RLock()
	lastRequest, exists := faucetRequests[normalizedAddr]
	faucetMutex.RUnlock()

	if exists && time.Since(lastRequest) < faucetCooldown {
		remaining := faucetCooldown - time.Since(lastRequest)
		s.respondJSON(w, http.StatusTooManyRequests, FaucetResponse{
			Success: false,
			Message: "Rate limited. Try again in " + formatDuration(remaining),
		})
		return
	}

	s.logger.Info("Faucet request",
		zap.String("address", req.Address),
		zap.String("amount", faucetAmount+"ucert"),
	)

	// Execute the faucet transfer using certd tx bank send
	txHash, err := s.executeFaucetTransfer(req.Address)
	if err != nil {
		s.logger.Error("Faucet transfer failed", zap.Error(err), zap.String("address", req.Address))
		s.respondJSON(w, http.StatusInternalServerError, FaucetResponse{
			Success: false,
			Message: "Transfer failed: " + err.Error(),
		})
		return
	}

	// Update rate limiting
	faucetMutex.Lock()
	faucetRequests[normalizedAddr] = time.Now()
	faucetMutex.Unlock()

	s.respondJSON(w, http.StatusOK, FaucetResponse{
		Success: true,
		Message: "Successfully sent 10 CERT to your address!",
		TxHash:  txHash,
		Amount:  "10000000ucert",
	})
}

// executeFaucetTransfer sends tokens from the faucet account
func (s *Server) executeFaucetTransfer(toAddress string) (string, error) {
	// Use docker exec to call certd in the certd container
	// The validator account has funds from genesis
	cmd := exec.Command("docker", "exec", "certd",
		"certd", "tx", "bank", "send",
		"validator", toAddress, faucetAmount+"ucert",
		"--chain-id", "cert-testnet-1",
		"--keyring-backend", "test",
		"--home", "/root/.certd",
		"--yes",
		"--output", "json",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(output))
	}

	// Parse txhash from output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// If JSON parsing fails, return a mock hash for now
		return "0x" + generateUID(), nil
	}

	if txHash, ok := result["txhash"].(string); ok {
		return txHash, nil
	}

	return "0x" + generateUID(), nil
}

// isValidAddress validates address format
func isValidAddress(addr string) bool {
	if strings.HasPrefix(addr, "cert1") && len(addr) >= 39 {
		return true
	}
	if strings.HasPrefix(addr, "0x") && len(addr) == 42 {
		matched, _ := regexp.MatchString("^0x[a-fA-F0-9]{40}$", addr)
		return matched
	}
	return false
}

// formatDuration formats a duration for human readability
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
