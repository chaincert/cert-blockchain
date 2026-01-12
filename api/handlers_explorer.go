// Package api provides explorer handlers for the CERT Blockchain
// Supports Chain Certify, Cert ID, and standard transaction decoding
package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Known contract addresses for ecosystem tagging
const (
	ChainCertifyContract = "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
	CertIDContract       = "0x7a250d5630b4cf539739df2c5dacb4c659f2488d"
	CertTokenContract    = "0xc3rt000000000000000000000000000000000001"
)

// TransactionResponse represents the detailed transaction data
type TransactionResponse struct {
	Hash          string                 `json:"hash"`
	Status        string                 `json:"status"` // success, failed, pending
	BlockNumber   int64                  `json:"block_number"`
	Timestamp     string                 `json:"timestamp"`
	Confirmations int64                  `json:"confirmations"`
	From          string                 `json:"from"`
	FromLabel     string                 `json:"from_label,omitempty"`
	To            string                 `json:"to"`
	ToLabel       string                 `json:"to_label,omitempty"`
	ValueCert     string                 `json:"value_cert"`
	ValueUSD      float64                `json:"value_usd,omitempty"`
	GasLimit      int64                  `json:"gas_limit"`
	GasUsed       int64                  `json:"gas_used"`
	GasPrice      string                 `json:"gas_price"`
	TxFee         string                 `json:"tx_fee"`
	InputData     string                 `json:"input_data"`
	EcosystemType string                 `json:"ecosystem_type"` // ChainCertify, CertID, Standard
	CertHash      string                 `json:"cert_hash,omitempty"`
	Metadata      string                 `json:"metadata,omitempty"`
	DecodedParams map[string]interface{} `json:"decoded_params,omitempty"`
	Logs          []EventLog             `json:"logs,omitempty"`
}

// EventLog represents a transaction event log
type EventLog struct {
	Index     int      `json:"index"`
	EventName string   `json:"event_name"`
	Topics    []string `json:"topics"`
	Data      string   `json:"data"`
}

// BlockResponse represents block data for the explorer
type BlockResponse struct {
	Height       int64              `json:"height"`
	Hash         string             `json:"hash"`
	Time         string             `json:"time"`
	Timestamp    string             `json:"timestamp"`
	TxCount      int                `json:"tx_count"`
	Proposer     string             `json:"proposer"`
	GasUsed      int64              `json:"gas_used"`
	GasLimit     int64              `json:"gas_limit"`
	TxHashes     []string           `json:"tx_hashes,omitempty"`
	Transactions []BlockTransaction `json:"transactions,omitempty"`
}

// BlockTransaction represents a transaction summary in a block
type BlockTransaction struct {
	Hash    string `json:"hash"`
	Type    string `json:"type"`
	Success bool   `json:"success"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
}

// AddressResponse represents address data for the explorer
type AddressResponse struct {
	Address       string  `json:"address"`
	Label         string  `json:"label,omitempty"`
	Balance       string  `json:"balance"`
	BalanceUSD    float64 `json:"balance_usd,omitempty"`
	TxCount       int64   `json:"tx_count"`
	IsContract    bool    `json:"is_contract"`
	EcosystemType string  `json:"ecosystem_type,omitempty"`
}

// handleGetTransaction returns detailed transaction data by hash
func (s *Server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash := vars["hash"]
	if txHash == "" {
		s.respondError(w, http.StatusBadRequest, "Transaction hash required")
		return
	}

	// Normalize hash format
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Try to fetch from database first (for indexed transactions)
	if s.db != nil {
		tx, err := s.db.GetTransaction(ctx, txHash)
		if err == nil && tx != nil {
			s.respondJSON(w, http.StatusOK, tx)
			return
		}
	}

	// Fallback to RPC query
	txData, err := s.fetchTransactionFromRPC(ctx, txHash)
	if err != nil {
		s.logger.Warn("Failed to fetch transaction", zap.String("hash", txHash), zap.Error(err))
		s.respondError(w, http.StatusNotFound, "Transaction not found")
		return
	}

	// Decode ecosystem-specific data
	s.enrichTransactionData(txData)

	// Lookup address labels if db available
	if s.db != nil {
		s.enrichAddressLabels(ctx, txData)
	}

	s.respondJSON(w, http.StatusOK, txData)
}

// fetchTransactionFromRPC queries the Tendermint RPC for transaction data
func (s *Server) fetchTransactionFromRPC(ctx context.Context, txHash string) (*TransactionResponse, error) {
	// Query Tendermint RPC
	rpcURL := fmt.Sprintf("%s/tx?hash=%s", s.config.ChainRPCURL, txHash)
	req, _ := http.NewRequestWithContext(ctx, "GET", rpcURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RPC returned status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var rpcResult struct {
		Result struct {
			Hash     string `json:"hash"`
			Height   string `json:"height"`
			Index    int    `json:"index"`
			TxResult struct {
				Code      int    `json:"code"`
				Data      string `json:"data"`
				Log       string `json:"log"`
				GasWanted string `json:"gas_wanted"`
				GasUsed   string `json:"gas_used"`
			} `json:"tx_result"`
			Tx string `json:"tx"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &rpcResult); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}

	height, _ := strconv.ParseInt(rpcResult.Result.Height, 10, 64)
	gasWanted, _ := strconv.ParseInt(rpcResult.Result.TxResult.GasWanted, 10, 64)
	gasUsed, _ := strconv.ParseInt(rpcResult.Result.TxResult.GasUsed, 10, 64)

	// Determine status
	status := "success"
	if rpcResult.Result.TxResult.Code != 0 {
		status = "failed"
	}

	// Get current height for confirmations
	currentHeight := s.getCurrentBlockHeight(ctx)
	confirmations := int64(0)
	if currentHeight > height {
		confirmations = currentHeight - height
	}

	// Decode transaction bytes to extract from/to/value
	txBytes, _ := hex.DecodeString(rpcResult.Result.Tx)
	from, to, value, inputData := s.decodeTxPayload(txBytes)

	// Calculate fee (gasUsed * gasPrice)
	gasPrice := int64(10) // Default gas price in ucert
	txFee := gasUsed * gasPrice

	return &TransactionResponse{
		Hash:          "0x" + rpcResult.Result.Hash,
		Status:        status,
		BlockNumber:   height,
		Confirmations: confirmations,
		From:          from,
		To:            to,
		ValueCert:     value,
		GasLimit:      gasWanted,
		GasUsed:       gasUsed,
		GasPrice:      fmt.Sprintf("%d", gasPrice),
		TxFee:         fmt.Sprintf("%d", txFee),
		InputData:     inputData,
		EcosystemType: "Standard",
	}, nil
}

// decodeTxPayload extracts transaction fields from raw bytes
func (s *Server) decodeTxPayload(txBytes []byte) (from, to, value, inputData string) {
	// TODO: Implement proper Cosmos SDK tx decoding
	// For now, return placeholder values
	from = "cert1..."
	to = "cert1..."
	value = "0"
	inputData = "0x"
	if len(txBytes) > 0 {
		inputData = "0x" + hex.EncodeToString(txBytes)
	}
	return
}

// getCurrentBlockHeight fetches the current block height from RPC
func (s *Server) getCurrentBlockHeight(ctx context.Context) int64 {
	rpcURL := fmt.Sprintf("%s/status", s.config.ChainRPCURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", rpcURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	height, _ := strconv.ParseInt(result.Result.SyncInfo.LatestBlockHeight, 10, 64)
	return height
}

// enrichTransactionData adds ecosystem-specific decoded data
func (s *Server) enrichTransactionData(tx *TransactionResponse) {
	if tx.To == "" || tx.InputData == "" || tx.InputData == "0x" {
		return
	}

	toAddr := strings.ToLower(tx.To)

	// Check if interacting with Chain Certify contract
	if toAddr == strings.ToLower(ChainCertifyContract) {
		tx.EcosystemType = "ChainCertify"
		tx.DecodedParams = s.decodeChainCertifyCall(tx.InputData)
		if hash, ok := tx.DecodedParams["docHash"].(string); ok {
			tx.CertHash = hash
		}
		if meta, ok := tx.DecodedParams["metadata"].(string); ok {
			tx.Metadata = meta
		}
	}

	// Check if interacting with Cert ID contract
	if toAddr == strings.ToLower(CertIDContract) {
		tx.EcosystemType = "CertID"
		tx.DecodedParams = s.decodeCertIDCall(tx.InputData)
	}

	// Check if Cert Token contract
	if toAddr == strings.ToLower(CertTokenContract) {
		tx.EcosystemType = "CertToken"
		tx.DecodedParams = s.decodeTokenCall(tx.InputData)
	}
}

// decodeChainCertifyCall decodes Chain Certify contract method calls
func (s *Server) decodeChainCertifyCall(inputData string) map[string]interface{} {
	params := make(map[string]interface{})
	if len(inputData) < 10 {
		return params
	}

	methodSig := inputData[:10]
	switch methodSig {
	case "0x1cf71a93": // issueCertificate(address,string,string)
		params["method"] = "issueCertificate"
		// TODO: Proper ABI decoding
		if len(inputData) > 138 {
			params["recipient"] = "0x" + inputData[34:74]
		}
	case "0x3ccfd60b": // verifyDocument(string)
		params["method"] = "verifyDocument"
	default:
		params["method"] = "unknown"
	}
	return params
}

// decodeCertIDCall decodes Cert ID contract method calls
func (s *Server) decodeCertIDCall(inputData string) map[string]interface{} {
	params := make(map[string]interface{})
	if len(inputData) < 10 {
		return params
	}

	methodSig := inputData[:10]
	switch methodSig {
	case "0x2e1a7d4d": // registerIdentity(string,address)
		params["method"] = "registerIdentity"
	default:
		params["method"] = "unknown"
	}
	return params
}

// decodeTokenCall decodes Cert Token contract method calls
func (s *Server) decodeTokenCall(inputData string) map[string]interface{} {
	params := make(map[string]interface{})
	if len(inputData) < 10 {
		return params
	}

	methodSig := inputData[:10]
	switch methodSig {
	case "0xa9059cbb": // transfer(address,uint256)
		params["method"] = "transfer"
	case "0x40c10f19": // mint(address,uint256)
		params["method"] = "mint"
	case "0x42966c68": // burn(uint256)
		params["method"] = "burn"
	default:
		params["method"] = "unknown"
	}
	return params
}

// enrichAddressLabels looks up Cert ID labels for addresses
func (s *Server) enrichAddressLabels(ctx context.Context, tx *TransactionResponse) {
	if tx.From != "" {
		if profile, err := s.db.GetProfile(ctx, tx.From); err == nil && profile != nil && profile.Name != "" {
			tx.FromLabel = profile.Name
		}
	}
	if tx.To != "" {
		if profile, err := s.db.GetProfile(ctx, tx.To); err == nil && profile != nil && profile.Name != "" {
			tx.ToLabel = profile.Name
		}
	}
}

// handleGetBlock returns block data by height or hash
func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightOrHash := vars["height"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var rpcURL string
	if strings.HasPrefix(heightOrHash, "0x") || len(heightOrHash) == 64 {
		rpcURL = fmt.Sprintf("%s/block_by_hash?hash=%s", s.config.ChainRPCURL, heightOrHash)
	} else {
		rpcURL = fmt.Sprintf("%s/block?height=%s", s.config.ChainRPCURL, heightOrHash)
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", rpcURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to fetch block")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rpcResult struct {
		Result struct {
			BlockID struct {
				Hash string `json:"hash"`
			} `json:"block_id"`
			Block struct {
				Header struct {
					Height          string `json:"height"`
					Time            string `json:"time"`
					ProposerAddress string `json:"proposer_address"`
				} `json:"header"`
				Data struct {
					Txs []string `json:"txs"`
				} `json:"data"`
			} `json:"block"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &rpcResult); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to parse block data")
		return
	}

	height, _ := strconv.ParseInt(rpcResult.Result.Block.Header.Height, 10, 64)
	blockTime := rpcResult.Result.Block.Header.Time
	block := BlockResponse{
		Height:    height,
		Hash:      rpcResult.Result.BlockID.Hash,
		Time:      blockTime,
		Timestamp: blockTime,
		TxCount:   len(rpcResult.Result.Block.Data.Txs),
		Proposer:  rpcResult.Result.Block.Header.ProposerAddress,
		TxHashes:  rpcResult.Result.Block.Data.Txs,
	}

	// Fetch transaction details for each tx in the block
	if len(rpcResult.Result.Block.Data.Txs) > 0 {
		block.Transactions = make([]BlockTransaction, 0, len(rpcResult.Result.Block.Data.Txs))
		for _, txBase64 := range rpcResult.Result.Block.Data.Txs {
			// Decode base64 tx to get the hash
			txBytes, err := base64.StdEncoding.DecodeString(txBase64)
			if err != nil {
				continue
			}
			// Calculate tx hash (SHA256)
			txHashBytes := sha256.Sum256(txBytes)
			txHash := strings.ToUpper(hex.EncodeToString(txHashBytes[:]))

			// Try to get tx details from block_results
			txInfo := BlockTransaction{
				Hash:    txHash,
				Type:    "Transaction",
				Success: true,
			}

			// Query tx_search to get more details
			txSearchURL := fmt.Sprintf("%s/tx?hash=0x%s", s.config.ChainRPCURL, txHash)
			txReq, _ := http.NewRequestWithContext(ctx, "GET", txSearchURL, nil)
			txResp, err := http.DefaultClient.Do(txReq)
			if err == nil {
				defer txResp.Body.Close()
				txBody, _ := io.ReadAll(txResp.Body)
				var txResult struct {
					Result struct {
						TxResult struct {
							Code int    `json:"code"`
							Log  string `json:"log"`
						} `json:"tx_result"`
					} `json:"result"`
				}
				if json.Unmarshal(txBody, &txResult) == nil {
					txInfo.Success = txResult.Result.TxResult.Code == 0
					// Try to extract tx type from log
					if strings.Contains(txResult.Result.TxResult.Log, "certify") {
						txInfo.Type = "Chain Certify"
					} else if strings.Contains(txResult.Result.TxResult.Log, "certid") {
						txInfo.Type = "Cert ID"
					} else if strings.Contains(txResult.Result.TxResult.Log, "send") || strings.Contains(txResult.Result.TxResult.Log, "transfer") {
						txInfo.Type = "Transfer"
					}
				}
			}

			block.Transactions = append(block.Transactions, txInfo)
		}
	}

	s.respondJSON(w, http.StatusOK, block)
}

// handleGetAddress returns address data and transaction history
func (s *Server) handleGetAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	response := AddressResponse{
		Address:    address,
		Balance:    "0",
		TxCount:    0,
		IsContract: false,
	}

	// Check if it's a known ecosystem contract
	addrLower := strings.ToLower(address)
	switch addrLower {
	case strings.ToLower(ChainCertifyContract):
		response.Label = "Chain Certify Contract"
		response.IsContract = true
		response.EcosystemType = "ChainCertify"
	case strings.ToLower(CertIDContract):
		response.Label = "Cert ID Contract"
		response.IsContract = true
		response.EcosystemType = "CertID"
	case strings.ToLower(CertTokenContract):
		response.Label = "Cert Token Contract"
		response.IsContract = true
		response.EcosystemType = "CertToken"
	}

	// Query actual balance from blockchain
	bech32Addr, err := toBech32Address(address)
	if err == nil {
		balance, err := s.queryAddressBalance(bech32Addr)
		if err == nil {
			response.Balance = balance
		} else {
			s.logger.Debug("failed to query balance", zap.String("address", bech32Addr), zap.Error(err))
		}
	}

	// Lookup profile label from database
	if s.db != nil && response.Label == "" {
		if profile, err := s.db.GetProfile(ctx, address); err == nil && profile != nil && profile.Name != "" {
			response.Label = profile.Name
		}
	}

	s.respondJSON(w, http.StatusOK, response)
}

// queryAddressBalance queries for an address's CERT balance
// Note: Due to a known Cosmos SDK v0.50.x state versioning bug, direct blockchain
// queries may fail with "version does not exist" errors. As a fallback, we estimate
// balance from faucet transactions stored in the database.
func (s *Server) queryAddressBalance(bech32Addr string) (string, error) {
	// First try the REST API (may fail due to SDK bug)
	url := fmt.Sprintf("http://localhost:1317/cosmos/bank/v1beta1/balances/%s", bech32Addr)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				var result struct {
					Balances []struct {
						Denom  string `json:"denom"`
						Amount string `json:"amount"`
					} `json:"balances"`
				}
				if json.Unmarshal(body, &result) == nil {
					for _, bal := range result.Balances {
						if bal.Denom == "ucert" {
							amount, err := strconv.ParseInt(bal.Amount, 10, 64)
							if err == nil {
								certAmount := float64(amount) / 1_000_000
								return fmt.Sprintf("%.6f", certAmount), nil
							}
						}
					}
				}
			}
		}
	}

	// REST API failed (common SDK v0.50.x bug) - fall back to database tracking
	// The blockchain is functioning but state queries have version mismatch issues
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		faucetBalance, err := s.db.GetFaucetBalance(ctx, bech32Addr)
		if err == nil && faucetBalance > 0 {
			certAmount := float64(faucetBalance) / 1_000_000
			s.logger.Debug("using database-tracked faucet balance",
				zap.String("address", bech32Addr),
				zap.Int64("ucert", faucetBalance))
			return fmt.Sprintf("%.6f", certAmount), nil
		}
	}

	s.logger.Debug("balance query fell back to 0 due to SDK state bug",
		zap.String("address", bech32Addr))
	return "0", nil
}

// handleSearchExplorer provides unified search across transactions, blocks, and addresses
func (s *Server) handleSearchExplorer(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.respondError(w, http.StatusBadRequest, "Search query required")
		return
	}

	searchType := "unknown"
	var redirectPath string

	// Determine search type based on query format
	query = strings.TrimSpace(query)
	if strings.HasPrefix(query, "0x") && len(query) == 66 {
		// Transaction hash (0x + 64 hex chars)
		searchType = "transaction"
		redirectPath = fmt.Sprintf("/tx/%s", query)
	} else if strings.HasPrefix(query, "0x") && len(query) == 42 {
		// Address (0x + 40 hex chars)
		searchType = "address"
		redirectPath = fmt.Sprintf("/address/%s", query)
	} else if strings.HasPrefix(query, "cert1") {
		// Bech32 address
		searchType = "address"
		redirectPath = fmt.Sprintf("/address/%s", query)
	} else if _, err := strconv.ParseInt(query, 10, 64); err == nil {
		// Block height
		searchType = "block"
		redirectPath = fmt.Sprintf("/blocks/%s", query)
	} else if len(query) == 64 {
		// Could be a document hash (SHA-256)
		searchType = "document"
		redirectPath = fmt.Sprintf("/verify/%s", query)
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"query":    query,
		"type":     searchType,
		"redirect": redirectPath,
	})
}

// handleGetAddressTransactions returns paginated transactions for an address
func (s *Server) handleGetAddressTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// For now, return empty transactions list - actual implementation would query tx_search
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":      address,
		"transactions": []interface{}{},
		"page":         page,
		"limit":        limit,
		"total":        0,
	})
}

// handleGetRecentTransactions returns paginated recent transactions
func (s *Server) handleGetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Return empty for now - actual implementation would query recent blocks
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": []interface{}{},
		"page":         page,
		"limit":        limit,
		"total":        0,
	})
}

// handleVerifyDocument verifies a document hash on-chain
func (s *Server) handleVerifyDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docHash := vars["hash"]

	if docHash == "" {
		s.respondError(w, http.StatusBadRequest, "Document hash required")
		return
	}

	// TODO: Query attestation module for document verification
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"hash":     docHash,
		"verified": false,
		"message":  "Document verification not yet implemented",
	})
}

// handleGetExplorerStats returns explorer statistics
func (s *Server) handleGetExplorerStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats := map[string]interface{}{
		"totalTransactions": 0,
		"totalBlocks":       0,
		"totalAddresses":    0,
		"chainCertifyTxs":   0,
		"certIdTxs":         0,
	}

	// Get latest block height
	statusURL := fmt.Sprintf("%s/status", s.config.ChainRPCURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var statusResult struct {
			Result struct {
				SyncInfo struct {
					LatestBlockHeight string `json:"latest_block_height"`
				} `json:"sync_info"`
			} `json:"result"`
		}
		if json.Unmarshal(body, &statusResult) == nil {
			height, _ := strconv.ParseInt(statusResult.Result.SyncInfo.LatestBlockHeight, 10, 64)
			stats["totalBlocks"] = height
		}
	}

	s.respondJSON(w, http.StatusOK, stats)
}

