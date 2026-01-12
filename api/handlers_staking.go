package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ValidatorInfo represents validator information for the API
type ValidatorInfo struct {
	OperatorAddress   string `json:"operator_address"`
	ConsensusPubkey   string `json:"consensus_pubkey,omitempty"`
	Jailed            bool   `json:"jailed"`
	Status            string `json:"status"`
	Tokens            string `json:"tokens"`
	DelegatorShares   string `json:"delegator_shares"`
	Description       ValidatorDescription `json:"description"`
	UnbondingHeight   string `json:"unbonding_height"`
	UnbondingTime     string `json:"unbonding_time"`
	Commission        ValidatorCommission  `json:"commission"`
	MinSelfDelegation string `json:"min_self_delegation"`
}

type ValidatorDescription struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

type ValidatorCommission struct {
	CommissionRates CommissionRates `json:"commission_rates"`
	UpdateTime      string          `json:"update_time"`
}

type CommissionRates struct {
	Rate          string `json:"rate"`
	MaxRate       string `json:"max_rate"`
	MaxChangeRate string `json:"max_change_rate"`
}

type ValidatorsResponse struct {
	Validators []ValidatorInfo `json:"validators"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// handleGetValidators returns all validators with their staking info
func (s *Server) handleGetValidators(w http.ResponseWriter, r *http.Request) {
	// Try REST API first
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED", getRESTBaseURL())

	resp, err := restClient.Get(url)
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result ValidatorsResponse
		if json.Unmarshal(body, &result) == nil && len(result.Validators) > 0 {
			s.respondJSON(w, http.StatusOK, result)
			return
		}
	}

	// Fallback to CometBFT RPC for consensus validators
	rpcValidators, err := s.getValidatorsFromRPC()
	if err != nil {
		s.logger.Warn("validators RPC query failed", zap.Error(err))
		s.respondJSON(w, http.StatusOK, ValidatorsResponse{Validators: []ValidatorInfo{}})
		return
	}

	s.respondJSON(w, http.StatusOK, rpcValidators)
}

// getValidatorsFromRPC queries CometBFT RPC for validator info
func (s *Server) getValidatorsFromRPC() (ValidatorsResponse, error) {
	url := fmt.Sprintf("%s/validators", getRPCBaseURL())

	resp, err := restClient.Get(url)
	if err != nil {
		return ValidatorsResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ValidatorsResponse{}, err
	}

	var rpcResult struct {
		Result struct {
			Validators []struct {
				Address  string `json:"address"`
				PubKey   struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"pub_key"`
				VotingPower      string `json:"voting_power"`
				ProposerPriority string `json:"proposer_priority"`
			} `json:"validators"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &rpcResult); err != nil {
		return ValidatorsResponse{}, err
	}

	// Convert RPC validators to our format
	validators := make([]ValidatorInfo, 0, len(rpcResult.Result.Validators))
	for i, v := range rpcResult.Result.Validators {
		validators = append(validators, ValidatorInfo{
			OperatorAddress: v.Address,
			ConsensusPubkey: v.PubKey.Value,
			Jailed:          false,
			Status:          "BOND_STATUS_BONDED",
			Tokens:          v.VotingPower + "000000", // Convert to ucert
			DelegatorShares: v.VotingPower + "000000000000000000",
			Description: ValidatorDescription{
				Moniker: fmt.Sprintf("Validator %d", i+1),
			},
			Commission: ValidatorCommission{
				CommissionRates: CommissionRates{
					Rate:          "0.050000000000000000",
					MaxRate:       "0.200000000000000000",
					MaxChangeRate: "0.010000000000000000",
				},
			},
			MinSelfDelegation: "1",
		})
	}

	return ValidatorsResponse{
		Validators: validators,
		Pagination: struct {
			NextKey string `json:"next_key"`
			Total   string `json:"total"`
		}{
			Total: fmt.Sprintf("%d", len(validators)),
		},
	}, nil
}

// handleGetValidator returns a specific validator by operator address
func (s *Server) handleGetValidator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	validatorAddr := vars["validator_address"]
	if validatorAddr == "" {
		s.respondError(w, http.StatusBadRequest, "validator_address is required")
		return
	}

	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", getRESTBaseURL(), validatorAddr)
	
	resp, err := restClient.Get(url)
	if err != nil {
		s.logger.Warn("validator query failed", zap.String("validator", validatorAddr), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query validator")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		s.respondError(w, http.StatusNotFound, "Validator not found")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.respondError(w, http.StatusBadGateway, "Failed to read response")
		return
	}

	var result struct {
		Validator ValidatorInfo `json:"validator"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		s.respondError(w, http.StatusBadGateway, "Failed to parse response")
		return
	}

	s.respondJSON(w, http.StatusOK, result)
}

// handleGetStakingParams returns staking module parameters
func (s *Server) handleGetStakingParams(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/params", getRESTBaseURL())
	
	resp, err := restClient.Get(url)
	if err != nil {
		// Return default params on error
		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"params": map[string]interface{}{
				"unbonding_time":     "1814400s", // 21 days
				"max_validators":     80,
				"max_entries":        7,
				"historical_entries": 10000,
				"bond_denom":         "ucert",
				"min_commission_rate": "0.050000000000000000",
			},
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// DelegateRequest represents a delegation request
type DelegateRequest struct {
	DelegatorAddress string `json:"delegator_address"`
	ValidatorAddress string `json:"validator_address"`
	Amount           string `json:"amount"` // in ucert
}

// handleDelegate handles POST /api/v1/staking/delegate
// For testnet, this creates an unsigned transaction that can be signed client-side
func (s *Server) handleDelegate(w http.ResponseWriter, r *http.Request) {
	var req DelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DelegatorAddress == "" || req.ValidatorAddress == "" || req.Amount == "" {
		s.respondError(w, http.StatusBadRequest, "Missing required fields: delegator_address, validator_address, amount")
		return
	}

	// Validate amount is positive integer
	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil || amount <= 0 {
		s.respondError(w, http.StatusBadRequest, "Invalid amount: must be positive integer")
		return
	}

	// For now, return the unsigned transaction message for client-side signing
	// In a production system, this would integrate with Keplr or other signing
	unsignedTx := map[string]interface{}{
		"body": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"@type":             "/cosmos.staking.v1beta1.MsgDelegate",
					"delegator_address": req.DelegatorAddress,
					"validator_address": req.ValidatorAddress,
					"amount": map[string]interface{}{
						"denom":  "ucert",
						"amount": req.Amount,
					},
				},
			},
			"memo":           "",
			"timeout_height": "0",
			"extension_options": []interface{}{},
			"non_critical_extension_options": []interface{}{},
		},
		"auth_info": map[string]interface{}{
			"signer_infos": []interface{}{},
			"fee": map[string]interface{}{
				"amount": []map[string]string{
					{"denom": "ucert", "amount": "10000"},
				},
				"gas_limit": "200000",
				"payer":     "",
				"granter":   "",
			},
		},
		"signatures": []string{},
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"unsigned_tx": unsignedTx,
		"message":     "Sign this transaction with your wallet and broadcast it",
	})
}

// handleUndelegate handles POST /api/v1/staking/undelegate
func (s *Server) handleUndelegate(w http.ResponseWriter, r *http.Request) {
	var req DelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DelegatorAddress == "" || req.ValidatorAddress == "" || req.Amount == "" {
		s.respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	unsignedTx := map[string]interface{}{
		"body": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"@type":             "/cosmos.staking.v1beta1.MsgUndelegate",
					"delegator_address": req.DelegatorAddress,
					"validator_address": req.ValidatorAddress,
					"amount": map[string]interface{}{
						"denom":  "ucert",
						"amount": req.Amount,
					},
				},
			},
			"memo": "",
		},
		"auth_info": map[string]interface{}{
			"fee": map[string]interface{}{
				"amount":    []map[string]string{{"denom": "ucert", "amount": "10000"}},
				"gas_limit": "200000",
			},
		},
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"unsigned_tx": unsignedTx,
		"message":     "Sign this transaction with your wallet and broadcast it",
	})
}

// handleGetRewards returns staking rewards for a delegator
func (s *Server) handleGetRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	// Convert to bech32 if needed
	bech32Addr := address
	if strings.HasPrefix(address, "0x") {
		var err error
		bech32Addr, err = toBech32Address(address)
		if err != nil {
			s.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	url := fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards", getRESTBaseURL(), bech32Addr)

	resp, err := restClient.Get(url)
	if err != nil {
		// Return empty rewards on error
		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"rewards": []interface{}{},
			"total":   []interface{}{},
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// RedelegateRequest represents a redelegation request
type RedelegateRequest struct {
	DelegatorAddress    string `json:"delegator_address"`
	SrcValidatorAddress string `json:"src_validator_address"`
	DstValidatorAddress string `json:"dst_validator_address"`
	Amount              string `json:"amount"` // in ucert
}

// handleRedelegate handles POST /api/v1/staking/redelegate
func (s *Server) handleRedelegate(w http.ResponseWriter, r *http.Request) {
	var req RedelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DelegatorAddress == "" || req.SrcValidatorAddress == "" || req.DstValidatorAddress == "" || req.Amount == "" {
		s.respondError(w, http.StatusBadRequest, "Missing required fields: delegator_address, src_validator_address, dst_validator_address, amount")
		return
	}

	amount, err := strconv.ParseInt(req.Amount, 10, 64)
	if err != nil || amount <= 0 {
		s.respondError(w, http.StatusBadRequest, "Invalid amount: must be positive integer")
		return
	}

	unsignedTx := map[string]interface{}{
		"body": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"@type":                 "/cosmos.staking.v1beta1.MsgBeginRedelegate",
					"delegator_address":     req.DelegatorAddress,
					"validator_src_address": req.SrcValidatorAddress,
					"validator_dst_address": req.DstValidatorAddress,
					"amount": map[string]interface{}{
						"denom":  "ucert",
						"amount": req.Amount,
					},
				},
			},
			"memo":           "",
			"timeout_height": "0",
		},
		"auth_info": map[string]interface{}{
			"signer_infos": []interface{}{},
			"fee": map[string]interface{}{
				"amount":    []map[string]string{{"denom": "ucert", "amount": "15000"}},
				"gas_limit": "250000",
			},
		},
		"signatures": []string{},
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"unsigned_tx": unsignedTx,
		"message":     "Sign this transaction with your wallet and broadcast it",
	})
}

// ClaimRewardsRequest represents a claim rewards request
type ClaimRewardsRequest struct {
	DelegatorAddress  string   `json:"delegator_address"`
	ValidatorAddresses []string `json:"validator_addresses,omitempty"` // If empty, claim from all
}

// handleClaimRewards handles POST /api/v1/staking/claim-rewards
func (s *Server) handleClaimRewards(w http.ResponseWriter, r *http.Request) {
	var req ClaimRewardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DelegatorAddress == "" {
		s.respondError(w, http.StatusBadRequest, "delegator_address is required")
		return
	}

	// Convert to bech32 if needed
	bech32Addr := req.DelegatorAddress
	if strings.HasPrefix(req.DelegatorAddress, "0x") {
		var err error
		bech32Addr, err = toBech32Address(req.DelegatorAddress)
		if err != nil {
			s.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// If no validators specified, get all validators for this delegator
	validators := req.ValidatorAddresses
	if len(validators) == 0 {
		// Query delegations to get validator list
		url := fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", getRESTBaseURL(), bech32Addr)
		resp, err := restClient.Get(url)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var delResp struct {
				DelegationResponses []struct {
					Delegation struct {
						ValidatorAddress string `json:"validator_address"`
					} `json:"delegation"`
				} `json:"delegation_responses"`
			}
			if json.Unmarshal(body, &delResp) == nil {
				for _, dr := range delResp.DelegationResponses {
					validators = append(validators, dr.Delegation.ValidatorAddress)
				}
			}
		}
	}

	if len(validators) == 0 {
		s.respondError(w, http.StatusBadRequest, "No delegations found to claim rewards from")
		return
	}

	// Create messages for each validator
	messages := make([]map[string]interface{}, 0, len(validators))
	for _, valAddr := range validators {
		messages = append(messages, map[string]interface{}{
			"@type":             "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
			"delegator_address": bech32Addr,
			"validator_address": valAddr,
		})
	}

	unsignedTx := map[string]interface{}{
		"body": map[string]interface{}{
			"messages":       messages,
			"memo":           "",
			"timeout_height": "0",
		},
		"auth_info": map[string]interface{}{
			"signer_infos": []interface{}{},
			"fee": map[string]interface{}{
				"amount":    []map[string]string{{"denom": "ucert", "amount": fmt.Sprintf("%d", 5000*len(validators))}},
				"gas_limit": fmt.Sprintf("%d", 100000*len(validators)),
			},
		},
		"signatures": []string{},
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"unsigned_tx":          unsignedTx,
		"validators_count":     len(validators),
		"message":              "Sign this transaction with your wallet and broadcast it",
	})
}

