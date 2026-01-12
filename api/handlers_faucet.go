// Package api - Faucet handler for testnet token distribution
package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"
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
	// Parse request body
	var req FaucetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, FaucetResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate address format
	if !isValidAddress(req.Address) {
		s.respondJSON(w, http.StatusBadRequest, FaucetResponse{
			Success: false,
			Message: "Invalid address format. Must be cert1... (bech32) or 0x... (hex)",
		})
		return
	}

	// Check rate limiting
	faucetMutex.RLock()
	lastRequest, exists := faucetRequests[req.Address]
	faucetMutex.RUnlock()

	if exists && time.Since(lastRequest) < faucetCooldown {
		remaining := faucetCooldown - time.Since(lastRequest)
		s.respondJSON(w, http.StatusTooManyRequests, FaucetResponse{
			Success: false,
			Message: fmt.Sprintf("Rate limited. Please wait %s before requesting again.", formatDuration(remaining)),
		})
		return
	}

	// Execute the transfer
	txHash, err := s.executeFaucetTransfer(req.Address)
	if err != nil {
		s.logger.Error("Faucet transfer failed", zap.Error(err))
		s.respondJSON(w, http.StatusInternalServerError, FaucetResponse{
			Success: false,
			Message: "Failed to send tokens. Please try again later.",
		})
		return
	}

	// Update rate limiting
	faucetMutex.Lock()
	faucetRequests[req.Address] = time.Now()
	faucetMutex.Unlock()

	// Record the transaction in the database for balance tracking
	// This helps when SDK state queries fail (known v0.50.x bug)
	if s.db != nil && txHash != "" {
		bech32Addr, _ := toBech32Address(req.Address)
		if bech32Addr != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.db.RecordFaucetTransaction(ctx, txHash, bech32Addr, 10000000); err != nil {
				s.logger.Warn("Failed to record faucet transaction",
					zap.String("tx_hash", txHash),
					zap.Error(err))
			}
		}
	}

	// Return success
	s.respondJSON(w, http.StatusOK, FaucetResponse{
		Success: true,
		Message: "Tokens sent successfully!",
		TxHash:  txHash,
		Amount:  "10 CERT",
	})
}

// Faucet sequence tracking for offline mode
var (
	faucetSequence    uint64 = 0
	faucetAccountNum  uint64 = 0
	faucetSeqMutex    sync.Mutex
	faucetInitialized bool = false
)

// executeFaucetTransfer sends tokens from the faucet account
// Uses 3-step offline signing flow to bypass gRPC state query issues in SDK v0.50.x:
// 1. generate-only: Creates unsigned tx JSON (no state query needed)
// 2. sign --offline: Signs tx without querying chain
// 3. broadcast: Sends signed tx to the network
func (s *Server) executeFaucetTransfer(toAddress string) (string, error) {
	// Convert EVM address (0x...) to bech32 (cert1...) if needed
	bech32Addr, err := toBech32Address(toAddress)
	if err != nil {
		return "", fmt.Errorf("address conversion failed: %w", err)
	}

	faucetSeqMutex.Lock()
	defer faucetSeqMutex.Unlock()

	// Initialize sequence on first use
	if !faucetInitialized {
		s.initializeFaucetSequence()
		faucetInitialized = true
	}

	// Build the shell script that performs generate -> sign -> broadcast.
	//
	// IMPORTANT: Do NOT pipe through `head -1` when producing the signed tx.
	// Depending on CLI output formatting, that can truncate JSON or capture a non-tx line.
	// If `/tmp/signed.json` is not an actual `Tx` JSON, `certd tx broadcast` fails with
	// errors like: `unknown field "codespace" in tx.Tx`.
	shellScript := fmt.Sprintf(`
set -eu

certd tx bank send validator %s %sucert \
  --keyring-backend test --home /root/.certd --chain-id cert-testnet-1 \
  --gas 200000 --fees 10000ucert --generate-only > /tmp/unsigned.json

# certd tx sign sometimes emits the signed tx JSON on stderr (SDK/CLI quirk).
# Capture its combined output, then extract the first full-line JSON object.
certd tx sign /tmp/unsigned.json \
  --from validator --keyring-backend test --home /root/.certd \
  --chain-id cert-testnet-1 --offline --account-number %d --sequence %d \
  --output json 2>&1 | sed -n 's/^\({.*}\)$/\1/p' | head -n 1 > /tmp/signed.json

test -s /tmp/signed.json

certd tx broadcast /tmp/signed.json \
  --node tcp://localhost:26657 --broadcast-mode sync --output json
`, bech32Addr, faucetAmount, faucetAccountNum, faucetSequence)

	// Execute the script inside the certd container
	cmd := exec.Command("docker", "exec", "certd", "sh", "-c", shellScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := string(output)
		// Check for sequence mismatch and retry
		if strings.Contains(errMsg, "account sequence mismatch") {
			if newSeq := extractSequenceFromError(errMsg); newSeq >= 0 {
				s.logger.Info("Sequence mismatch, retrying", zap.Int64("new_seq", newSeq))
				faucetSequence = uint64(newSeq)
				faucetSeqMutex.Unlock()
				return s.executeFaucetTransfer(toAddress)
			}
		}
		return "", fmt.Errorf("faucet tx failed: %v: %s", err, errMsg)
	}

	// Parse output for txhash
	txHash := s.extractTxHash(string(output))
	if txHash == "" {
		s.logger.Warn("Could not extract txhash from output", zap.String("output", string(output)))
		txHash = "tx_sent"
	}

	// Check for tx errors in the response
	if err := s.checkBroadcastForErrors(string(output)); err != nil {
		return "", err
	}

	// Increment sequence for next transaction
	faucetSequence++

	return txHash, nil
}

type txBroadcastResponse struct {
	TxHash    string `json:"txhash"`
	Code      int    `json:"code"`
	Codespace string `json:"codespace"`
	RawLog    string `json:"raw_log"`
	Log       string `json:"log"`
}

// extractTxHash extracts txhash from CLI broadcast output.
// Prefer JSON (when broadcast uses `--output json`), fallback to YAML parsing.
func (s *Server) extractTxHash(output string) string {
	if resp, ok := extractFirstJSONObject[txBroadcastResponse](output); ok {
		if resp.TxHash != "" {
			return resp.TxHash
		}
	}

	// Fallback: YAML output format from `certd tx broadcast` (legacy / non-json):
	// txhash: AD81519A98F2964CE19E15096552CC9FAF5DF7DDA8382F775568104C6A38B8E3
	re := regexp.MustCompile(`txhash:\s*([A-Fa-f0-9]{64})`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func (s *Server) checkBroadcastForErrors(output string) error {
	if resp, ok := extractFirstJSONObject[txBroadcastResponse](output); ok {
		// Cosmos SDK: code == 0 indicates success
		// code == 19 (sdk.ErrTxInMempoolCache) means the tx is already in the mempool cache;
		// treat it as success for faucet purposes to avoid repeated duplicate-broadcast failures.
		if resp.Codespace == "sdk" && resp.Code == 19 {
			s.logger.Warn("broadcast returned tx already in mempool cache; treating as success",
				zap.String("txhash", resp.TxHash),
				zap.Int("code", resp.Code),
			)
			return nil
		}
		if resp.Code != 0 {
			msg := resp.RawLog
			if msg == "" {
				msg = resp.Log
			}
			if msg == "" {
				msg = output
			}
			return fmt.Errorf("broadcast failed (codespace=%s code=%d): %s", resp.Codespace, resp.Code, msg)
		}
		return nil
	}

	// Fallback: best-effort detect YAML style failures
	if strings.Contains(output, "code:") && !strings.Contains(output, "code: 0") {
		return fmt.Errorf("broadcast failed: %s", output)
	}

	return nil
}

func extractFirstJSONObject[T any](s string) (T, bool) {
	var zero T
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start < 0 || end < 0 || end <= start {
		return zero, false
	}

	var v T
	if err := json.Unmarshal([]byte(s[start:end+1]), &v); err != nil {
		return zero, false
	}
	return v, true
}

// initializeFaucetSequence sets up initial sequence (starts at 0 for new chains)
func (s *Server) initializeFaucetSequence() {
	// Get validator address for logging
	addrCmd := exec.Command("docker", "exec", "certd",
		"certd", "keys", "show", "validator",
		"--keyring-backend", "test",
		"--home", "/root/.certd",
		"-a",
	)
	addrOutput, _ := addrCmd.CombinedOutput()
	validatorAddr := strings.TrimSpace(string(addrOutput))
	s.logger.Info("Faucet validator address", zap.String("address", validatorAddr))

	// For new chains, validator is account 0 with sequence 0
	faucetAccountNum = 0
	faucetSequence = 0
}

// extractSequenceFromError parses sequence from "account sequence mismatch" error
func extractSequenceFromError(output string) int64 {
	// Error format: "account sequence mismatch, expected X, got Y"
	re := regexp.MustCompile(`expected (\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		var seq int64
		fmt.Sscanf(matches[1], "%d", &seq)
		return seq
	}
	return -1
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

// toBech32Address converts an EVM hex address (0x...) to bech32 (cert1...)
// If already in bech32 format, returns as-is
func toBech32Address(addr string) (string, error) {
	// Already bech32 format
	if strings.HasPrefix(addr, "cert1") {
		return addr, nil
	}

	// Convert from EVM hex address
	if strings.HasPrefix(addr, "0x") || strings.HasPrefix(addr, "0X") {
		// Remove 0x prefix and decode hex
		hexAddr := strings.TrimPrefix(strings.ToLower(addr), "0x")
		addrBytes, err := hex.DecodeString(hexAddr)
		if err != nil {
			return "", fmt.Errorf("invalid hex address: %w", err)
		}

		// Convert to bech32 with "cert" prefix
		bech32Addr, err := bech32.ConvertAndEncode("cert", addrBytes)
		if err != nil {
			return "", fmt.Errorf("bech32 encoding failed: %w", err)
		}

		return bech32Addr, nil
	}

	return "", fmt.Errorf("unsupported address format: %s", addr)
}
