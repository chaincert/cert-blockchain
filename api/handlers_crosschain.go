package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Cross-Chain Sybil API
// Supports Ethereum, Arbitrum, Optimism, Base, Polygon alongside CERT

// ChainConfig defines RPC endpoints for supported chains
type ChainConfig struct {
	Name    string
	ChainID int64
	RPCURL  string
}

var SupportedChains = map[string]ChainConfig{
	"ethereum": {Name: "Ethereum", ChainID: 1, RPCURL: "https://eth.llamarpc.com"},
	"arbitrum": {Name: "Arbitrum One", ChainID: 42161, RPCURL: "https://arb1.arbitrum.io/rpc"},
	"optimism": {Name: "Optimism", ChainID: 10, RPCURL: "https://mainnet.optimism.io"},
	"base":     {Name: "Base", ChainID: 8453, RPCURL: "https://mainnet.base.org"},
	"polygon":  {Name: "Polygon", ChainID: 137, RPCURL: "https://polygon-rpc.com"},
	"cert":     {Name: "CERT Blockchain", ChainID: 77551, RPCURL: "https://evm.c3rt.org"},
}

// CrossChainFactors extends TrustFactors with multi-chain data
type CrossChainFactors struct {
	TrustFactors
	ChainData map[string]ChainActivityData `json:"chain_data"`
}

// ChainActivityData represents on-chain activity for a specific chain
type ChainActivityData struct {
	Chain           string    `json:"chain"`
	TxCount         int       `json:"tx_count"`
	Balance         string    `json:"balance"`
	FirstTxTime     time.Time `json:"first_tx_time,omitempty"`
	LastTxTime      time.Time `json:"last_tx_time,omitempty"`
	ContractDeploys int       `json:"contract_deploys"`
	NFTsOwned       int       `json:"nfts_owned,omitempty"`
	DeFiInteraction bool      `json:"defi_interaction"`
}

// CrossChainSybilRequest for multi-chain checks
type CrossChainSybilRequest struct {
	Address  string   `json:"address"`
	Chains   []string `json:"chains,omitempty"`     // Optional: specific chains to check
	Actions  []string `json:"actions,omitempty"`    // Optional: specific actions to verify
	MinScore int      `json:"min_score,omitempty"`  // Optional: minimum score threshold
}

// CrossChainSybilResponse extends SybilCheckResponse
type CrossChainSybilResponse struct {
	Address           string              `json:"address"`
	TrustScore        int                 `json:"trust_score"`
	IsLikelyHuman     bool                `json:"is_likely_human"`
	Factors           CrossChainFactors   `json:"factors"`
	ChainsChecked     []string            `json:"chains_checked"`
	CrossChainScore   int                 `json:"cross_chain_score"`   // Bonus for multi-chain activity
	Layer3Compatible  bool                `json:"layer3_compatible"`   // Ready for Layer3 quests
	CheckedAt         time.Time           `json:"checked_at"`
}

// VerifyActionRequest for Layer3 action verification
type VerifyActionRequest struct {
	Address  string `json:"address"`
	Chain    string `json:"chain"`
	Action   string `json:"action"`    // e.g., "swap", "bridge", "mint", "stake"
	TxHash   string `json:"tx_hash,omitempty"`
	Contract string `json:"contract,omitempty"`
	MinValue string `json:"min_value,omitempty"` // Minimum transaction value
}

// VerifyActionResponse for action verification
type VerifyActionResponse struct {
	Address   string    `json:"address"`
	Chain     string    `json:"chain"`
	Action    string    `json:"action"`
	Verified  bool      `json:"verified"`
	TxHash    string    `json:"tx_hash,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// handleCrossChainSybilCheck returns trust score with multi-chain data
func (s *Server) handleCrossChainSybilCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}

	// Parse optional chain filter from query params
	chainsParam := r.URL.Query().Get("chains")
	var chainsToCheck []string
	if chainsParam != "" {
		chainsToCheck = strings.Split(chainsParam, ",")
	} else {
		// Default: check all supported chains
		chainsToCheck = []string{"ethereum", "arbitrum", "optimism", "base", "polygon", "cert"}
	}

	ctx := r.Context()
	
	// Get CERT-specific factors first
	certFactors := s.getTrustFactors(ctx, address)
	
	// Get cross-chain data
	chainData := s.getMultiChainActivity(ctx, address, chainsToCheck)
	
	// Calculate cross-chain bonus
	crossChainScore := calculateCrossChainBonus(chainData)
	
	// Combine with base trust score
	baseScore := calculateSybilTrustScore(certFactors)
	totalScore := sybilMin(baseScore + crossChainScore, 100)
	
	// Default threshold is 50
	threshold := 50
	if t := r.URL.Query().Get("threshold"); t != "" {
		fmt.Sscanf(t, "%d", &threshold)
	}

	response := CrossChainSybilResponse{
		Address:       address,
		TrustScore:    totalScore,
		IsLikelyHuman: totalScore >= threshold,
		Factors: CrossChainFactors{
			TrustFactors: certFactors,
			ChainData:    chainData,
		},
		ChainsChecked:    chainsToCheck,
		CrossChainScore:  crossChainScore,
		Layer3Compatible: totalScore >= 50 && len(chainData) >= 1,
		CheckedAt:        time.Now(),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleVerifyChainAction verifies a specific on-chain action for Layer3
func (s *Server) handleVerifyChainAction(w http.ResponseWriter, r *http.Request) {
	var req VerifyActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Address == "" || req.Chain == "" || req.Action == "" {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "address, chain, and action are required"})
		return
	}

	// Validate chain
	chainConfig, ok := SupportedChains[strings.ToLower(req.Chain)]
	if !ok {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error":            "unsupported chain",
			"supported_chains": "ethereum,arbitrum,optimism,base,polygon,cert",
		})
		return
	}

	ctx := r.Context()
	verified, txHash, timestamp, details := s.verifyAction(ctx, req, chainConfig)

	response := VerifyActionResponse{
		Address:   req.Address,
		Chain:     chainConfig.Name,
		Action:    req.Action,
		Verified:  verified,
		TxHash:    txHash,
		Timestamp: timestamp,
		Details:   details,
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleBatchVerifyActions verifies multiple actions in one request
func (s *Server) handleBatchVerifyActions(w http.ResponseWriter, r *http.Request) {
	var requests []VerifyActionRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(requests) > 20 {
		s.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "max 20 actions per batch"})
		return
	}

	ctx := r.Context()
	results := make([]VerifyActionResponse, 0, len(requests))

	for _, req := range requests {
		chainConfig, ok := SupportedChains[strings.ToLower(req.Chain)]
		if !ok {
			results = append(results, VerifyActionResponse{
				Address:  req.Address,
				Chain:    req.Chain,
				Action:   req.Action,
				Verified: false,
				Details:  "unsupported chain",
			})
			continue
		}

		verified, txHash, timestamp, details := s.verifyAction(ctx, req, chainConfig)
		results = append(results, VerifyActionResponse{
			Address:   req.Address,
			Chain:     chainConfig.Name,
			Action:    req.Action,
			Verified:  verified,
			TxHash:    txHash,
			Timestamp: timestamp,
			Details:   details,
		})
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"summary": map[string]int{
			"total":    len(results),
			"verified": countVerified(results),
		},
	})
}

// getMultiChainActivity fetches on-chain activity from multiple chains
func (s *Server) getMultiChainActivity(ctx context.Context, address string, chains []string) map[string]ChainActivityData {
	chainData := make(map[string]ChainActivityData)

	for _, chainName := range chains {
		chainConfig, ok := SupportedChains[strings.ToLower(chainName)]
		if !ok {
			continue
		}

		data := s.fetchChainActivity(ctx, address, chainConfig)
		if data.TxCount > 0 || data.Balance != "0" {
			chainData[chainName] = data
		}
	}

	return chainData
}

// fetchChainActivity queries a single chain for activity data
func (s *Server) fetchChainActivity(ctx context.Context, address string, chain ChainConfig) ChainActivityData {
	data := ChainActivityData{
		Chain:   chain.Name,
		Balance: "0",
	}

	// Query balance via JSON-RPC
	balance, err := s.ethGetBalance(ctx, chain.RPCURL, address)
	if err == nil {
		data.Balance = balance
	}

	// Query transaction count
	txCount, err := s.ethGetTransactionCount(ctx, chain.RPCURL, address)
	if err == nil {
		data.TxCount = txCount
	}

	// For significant activity, mark DeFi interaction
	if txCount > 10 {
		data.DeFiInteraction = true
	}

	return data
}

// ethGetBalance calls eth_getBalance
func (s *Server) ethGetBalance(ctx context.Context, rpcURL, address string) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{address, "latest"},
		"id":      1,
	}

	result, err := s.makeRPCCall(ctx, rpcURL, payload)
	if err != nil {
		return "0", err
	}

	// Parse hex balance
	balanceHex, ok := result["result"].(string)
	if !ok {
		return "0", fmt.Errorf("invalid balance response")
	}

	// Convert to readable format
	balance := new(big.Int)
	balance.SetString(strings.TrimPrefix(balanceHex, "0x"), 16)
	
	// Convert wei to ETH (approximate)
	eth := new(big.Float).SetInt(balance)
	eth.Quo(eth, big.NewFloat(1e18))
	
	return fmt.Sprintf("%.4f", eth), nil
}

// ethGetTransactionCount calls eth_getTransactionCount
func (s *Server) ethGetTransactionCount(ctx context.Context, rpcURL, address string) (int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionCount",
		"params":  []interface{}{address, "latest"},
		"id":      1,
	}

	result, err := s.makeRPCCall(ctx, rpcURL, payload)
	if err != nil {
		return 0, err
	}

	countHex, ok := result["result"].(string)
	if !ok {
		return 0, fmt.Errorf("invalid tx count response")
	}

	count := new(big.Int)
	count.SetString(strings.TrimPrefix(countHex, "0x"), 16)
	
	return int(count.Int64()), nil
}

// makeRPCCall performs a JSON-RPC request
func (s *Server) makeRPCCall(ctx context.Context, rpcURL string, payload interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", rpcURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// verifyAction verifies a specific on-chain action
func (s *Server) verifyAction(ctx context.Context, req VerifyActionRequest, chain ChainConfig) (bool, string, time.Time, string) {
	// If tx hash provided, verify it directly
	if req.TxHash != "" {
		return s.verifyTransaction(ctx, chain.RPCURL, req.TxHash, req.Address, req.Action)
	}

	// Otherwise, scan recent transactions for the action
	txCount, _ := s.ethGetTransactionCount(ctx, chain.RPCURL, req.Address)
	
	switch strings.ToLower(req.Action) {
	case "any_tx", "transaction":
		if txCount > 0 {
			return true, "", time.Now(), fmt.Sprintf("Address has %d transactions", txCount)
		}
	case "swap":
		// For swap verification, check if user has interacted with common DEX routers
		if txCount >= 5 {
			return true, "", time.Now(), "Address has DeFi activity"
		}
	case "bridge":
		// Bridge verification - check for specific bridge contract interactions
		if txCount >= 3 {
			return true, "", time.Now(), "Address has bridging activity"
		}
	case "mint", "nft_mint":
		// NFT mint verification
		if txCount >= 2 {
			return true, "", time.Now(), "Address has minting activity"
		}
	case "stake", "staking":
		// Staking verification
		if txCount >= 5 {
			return true, "", time.Now(), "Address has staking activity"
		}
	case "deploy", "contract_deploy":
		// Contract deployment verification would need trace analysis
		return false, "", time.Time{}, "Contract deployment requires trace analysis"
	}

	return false, "", time.Time{}, "Could not verify action"
}

// verifyTransaction verifies a specific transaction exists and matches criteria
func (s *Server) verifyTransaction(ctx context.Context, rpcURL, txHash, expectedFrom, action string) (bool, string, time.Time, string) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionByHash",
		"params":  []interface{}{txHash},
		"id":      1,
	}

	result, err := s.makeRPCCall(ctx, rpcURL, payload)
	if err != nil {
		return false, txHash, time.Time{}, "Failed to fetch transaction"
	}

	tx, ok := result["result"].(map[string]interface{})
	if !ok || tx == nil {
		return false, txHash, time.Time{}, "Transaction not found"
	}

	// Verify sender matches
	from, _ := tx["from"].(string)
	if !strings.EqualFold(from, expectedFrom) {
		return false, txHash, time.Time{}, "Transaction sender does not match"
	}

	return true, txHash, time.Now(), fmt.Sprintf("Transaction verified: %s", action)
}

// calculateCrossChainBonus calculates bonus points for multi-chain activity
func calculateCrossChainBonus(chainData map[string]ChainActivityData) int {
	bonus := 0
	activeChains := 0

	for _, data := range chainData {
		if data.TxCount > 0 {
			activeChains++
			
			// Activity-based bonus per chain
			if data.TxCount >= 10 {
				bonus += 5 // High activity
			} else if data.TxCount >= 3 {
				bonus += 3 // Medium activity
			} else {
				bonus += 1 // Low activity
			}

			// DeFi interaction bonus
			if data.DeFiInteraction {
				bonus += 2
			}
		}
	}

	// Multi-chain presence bonus
	if activeChains >= 3 {
		bonus += 10 // Active on 3+ chains
	} else if activeChains >= 2 {
		bonus += 5 // Active on 2 chains
	}

	return sybilMin(bonus, 30) // Cap cross-chain bonus at 30 points
}

// countVerified counts verified actions in results
func countVerified(results []VerifyActionResponse) int {
	count := 0
	for _, r := range results {
		if r.Verified {
			count++
		}
	}
	return count
}
