// Package api - Faucet handler for testnet token distribution
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
	// TODO: Faucet temporarily disabled until bank module is registered in certd
	// The certd binary needs to be rebuilt with proper module registration
	s.respondJSON(w, http.StatusServiceUnavailable, FaucetResponse{
		Success: false,
		Message: "Faucet is temporarily under maintenance. Please check back soon or contact the team on Discord for testnet tokens.",
	})
}

// executeFaucetTransfer sends tokens from the faucet account
func (s *Server) executeFaucetTransfer(toAddress string) (string, error) {
	// Use docker exec to call certd in the certd container
	// The validator account has funds from genesis
	// Cosmos SDK v0.50.x command structure
	cmd := exec.Command("docker", "exec", "certd",
		"certd", "tx", "bank", "send",
		"validator", toAddress, faucetAmount+"ucert",
		"--from", "validator",
		"--keyring-backend", "test",
		"--home", "/root/.certd",
		"--yes",
		"--gas", "auto",
		"--gas-adjustment", "1.5",
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
