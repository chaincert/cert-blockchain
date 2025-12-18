package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NOTE: These endpoints are primarily for the web UX. They intentionally:
// - accept both `cert1...` (bech32) and `0x...` (EVM hex) addresses
// - aggregate wallet + staking + attestation stats for the User Dashboard
// - default to TESTNET semantics (tokens have no real-world value)

type walletBalanceResult struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
}

type stakingDelegationsResult struct {
	DelegationResponses []struct {
		Delegation struct {
			DelegatorAddress string `json:"delegator_address"`
			ValidatorAddress string `json:"validator_address"`
		} `json:"delegation"`
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balance"`
	} `json:"delegation_responses"`
}

type attestationListResult struct {
	Attestations []struct {
		UID             string `json:"uid"`
		SchemaUID       string `json:"schema_uid"`
		Attester        string `json:"attester"`
		Recipient       string `json:"recipient"`
		Time            string `json:"time"`
		AttestationType string `json:"attestation_type"`
	} `json:"attestations"`
}

type DashboardSummaryResponse struct {
	Address       string `json:"address"`
	Bech32Address string `json:"bech32_address"`

	Wallet struct {
		BalanceUcert string `json:"balance_ucert"`
	} `json:"wallet"`

	Staking struct {
		StakedUcert string  `json:"staked_ucert"`
		ApyPercent  float64 `json:"apy_percent"`
	} `json:"staking"`

	Attestations struct {
		ReceivedCount int `json:"received_count"`
		IssuedCount   int `json:"issued_count"`
	} `json:"attestations"`

	Network struct {
		Name       string `json:"name"`
		IsTestnet  bool   `json:"is_testnet"`
		Disclaimer string `json:"disclaimer"`
	} `json:"network"`
}

type WalletBalanceResponse struct {
	Address       string `json:"address"`
	Bech32Address string `json:"bech32_address"`
	Denom         string `json:"denom"`
	BalanceUcert  string `json:"balance_ucert"`
}

type StakingDelegationsResponse struct {
	Address          string `json:"address"`
	Bech32Address    string `json:"bech32_address"`
	BondDenom        string `json:"bond_denom"`
	TotalStakedUcert string `json:"total_staked_ucert"`
	Delegations      []struct {
		ValidatorAddress string `json:"validator_address"`
		AmountUcert      string `json:"amount_ucert"`
	} `json:"delegations"`
}

type StakingSummaryResponse struct {
	Address       string  `json:"address"`
	Bech32Address string  `json:"bech32_address"`
	StakedUcert   string  `json:"staked_ucert"`
	ApyPercent    float64 `json:"apy_percent"`
}

func (s *Server) handleGetWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	bech32Addr, err := toBech32Address(address)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	balanceUcert, err := s.queryWalletBalanceUcert(bech32Addr)
	if err != nil {
		s.logger.Warn("wallet balance query failed", zap.String("address", bech32Addr), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query wallet balance")
		return
	}

	s.respondJSON(w, http.StatusOK, WalletBalanceResponse{
		Address:       address,
		Bech32Address: bech32Addr,
		Denom:         "ucert",
		BalanceUcert:  balanceUcert,
	})
}

func (s *Server) handleGetStakingDelegations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	bech32Addr, err := toBech32Address(address)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := s.queryStakingDelegations(bech32Addr)
	if err != nil {
		s.logger.Warn("staking delegations query failed", zap.String("address", bech32Addr), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query staking delegations")
		return
	}

	var out StakingDelegationsResponse
	out.Address = address
	out.Bech32Address = bech32Addr
	out.BondDenom = "ucert"
	out.TotalStakedUcert = "0"

	total := int64(0)
	for _, dr := range res.DelegationResponses {
		if dr.Balance.Denom != "ucert" {
			continue
		}
		amt, _ := strconv.ParseInt(dr.Balance.Amount, 10, 64)
		total += amt
		out.Delegations = append(out.Delegations, struct {
			ValidatorAddress string `json:"validator_address"`
			AmountUcert      string `json:"amount_ucert"`
		}{
			ValidatorAddress: dr.Delegation.ValidatorAddress,
			AmountUcert:      dr.Balance.Amount,
		})
	}
	out.TotalStakedUcert = strconv.FormatInt(total, 10)

	s.respondJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetStakingSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	bech32Addr, err := toBech32Address(address)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stakedUcert, err := s.queryTotalStakedUcert(bech32Addr)
	if err != nil {
		s.logger.Warn("staking summary query failed", zap.String("address", bech32Addr), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query staking summary")
		return
	}

	s.respondJSON(w, http.StatusOK, StakingSummaryResponse{
		Address:       address,
		Bech32Address: bech32Addr,
		StakedUcert:   stakedUcert,
		ApyPercent:    s.getTestnetAPYPercent(),
	})
}

func (s *Server) handleGetDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	bech32Addr, err := toBech32Address(address)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	balanceUcert, balErr := s.queryWalletBalanceUcert(bech32Addr)
	stakedUcert, stakeErr := s.queryTotalStakedUcert(bech32Addr)
	received, recvErr := s.queryAttestationsByRecipient(bech32Addr)
	issued, issErr := s.queryAttestationsByAttester(bech32Addr)

	// If *everything* fails, treat as upstream failure.
	if balErr != nil && stakeErr != nil && recvErr != nil && issErr != nil {
		s.logger.Warn("dashboard aggregate query failed",
			zap.String("address", bech32Addr),
			zap.Error(balErr),
			zap.Error(stakeErr),
			zap.Error(recvErr),
			zap.Error(issErr),
		)
		s.respondError(w, http.StatusBadGateway, "Failed to query dashboard data")
		return
	}

	var resp DashboardSummaryResponse
	resp.Address = address
	resp.Bech32Address = bech32Addr
	resp.Wallet.BalanceUcert = safeString(balanceUcert)
	resp.Staking.StakedUcert = safeString(stakedUcert)
	resp.Staking.ApyPercent = s.getTestnetAPYPercent()
	resp.Attestations.ReceivedCount = len(received)
	resp.Attestations.IssuedCount = len(issued)
	resp.Network.Name = "CERT Testnet"
	resp.Network.IsTestnet = true
	resp.Network.Disclaimer = "CERT Testnet only. Tokens have no real-world value. Participation may make you eligible for a future mainnet airdrop."

	s.respondJSON(w, http.StatusOK, resp)
}

func safeString(v string) string {
	if v == "" {
		return "0"
	}
	return v
}

func (s *Server) getTestnetAPYPercent() float64 {
	// Fixed value for testnet UX. Can be overridden at deploy time.
	// Example: TESTNET_STAKING_APY_PERCENT=10
	if v := strings.TrimSpace(os.Getenv("TESTNET_STAKING_APY_PERCENT")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 10.0
}

func (s *Server) queryWalletBalanceUcert(bech32Addr string) (string, error) {
	var res walletBalanceResult
	if err := s.execCertdQueryJSON(&res, "bank", "balances", bech32Addr); err != nil {
		return "", err
	}
	for _, b := range res.Balances {
		if b.Denom == "ucert" {
			return b.Amount, nil
		}
	}
	return "0", nil
}

func (s *Server) queryStakingDelegations(bech32Addr string) (stakingDelegationsResult, error) {
	var res stakingDelegationsResult
	if err := s.execCertdQueryJSON(&res, "staking", "delegations", bech32Addr); err != nil {
		return stakingDelegationsResult{}, err
	}
	return res, nil
}

func (s *Server) queryTotalStakedUcert(bech32Addr string) (string, error) {
	res, err := s.queryStakingDelegations(bech32Addr)
	if err != nil {
		return "", err
	}
	total := int64(0)
	for _, dr := range res.DelegationResponses {
		if dr.Balance.Denom != "ucert" {
			continue
		}
		amt, _ := strconv.ParseInt(dr.Balance.Amount, 10, 64)
		total += amt
	}
	return strconv.FormatInt(total, 10), nil
}

func (s *Server) queryAttestationsByRecipient(bech32Addr string) ([]map[string]any, error) {
	var res attestationListResult
	if err := s.execCertdQueryJSON(&res, "attestation", "by-recipient", bech32Addr); err != nil {
		return nil, err
	}
	return normalizeAttestations(res.Attestations), nil
}

func (s *Server) queryAttestationsByAttester(bech32Addr string) ([]map[string]any, error) {
	var res attestationListResult
	if err := s.execCertdQueryJSON(&res, "attestation", "by-attester", bech32Addr); err != nil {
		return nil, err
	}
	return normalizeAttestations(res.Attestations), nil
}

func normalizeAttestations(items []struct {
	UID             string `json:"uid"`
	SchemaUID       string `json:"schema_uid"`
	Attester        string `json:"attester"`
	Recipient       string `json:"recipient"`
	Time            string `json:"time"`
	AttestationType string `json:"attestation_type"`
}) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, a := range items {
		attester := normalizeMaybeAddress(a.Attester)
		recipient := normalizeMaybeAddress(a.Recipient)
		out = append(out, map[string]any{
			"uid":       a.UID,
			"schema":    a.SchemaUID,
			"issuer":    attester,
			"recipient": recipient,
			"time":      a.Time,
			"encrypted": a.AttestationType != "public" && a.AttestationType != "",
			"type":      a.AttestationType,
		})
	}
	return out
}

func normalizeMaybeAddress(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "cert1") || strings.HasPrefix(s, "0x") {
		return s
	}
	// Many Cosmos CLI JSON outputs encode address bytes as base64. Best-effort decode.
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	if len(decoded) != 20 {
		return s
	}
	bech, err := bech32.ConvertAndEncode("cert", decoded)
	if err != nil {
		return s
	}
	return bech
}

func (s *Server) execCertdQueryJSON(out any, queryArgs ...string) error {
	// Run inside docker: `certd query ... --node tcp://localhost:26657 --output json`
	args := []string{"query"}
	args = append(args, queryArgs...)
	args = append(args, "--node", "tcp://localhost:26657", "--output", "json")

	cmd := exec.Command("docker", append([]string{"exec", "certd", "certd"}, args...)...)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certd query failed: %w: %s", err, string(buf))
	}

	// certd can print extra lines; extract first JSON object.
	var raw json.RawMessage
	if v, ok := extractFirstJSONObject[json.RawMessage](string(buf)); ok {
		raw = v
	} else {
		raw = json.RawMessage(strings.TrimSpace(string(buf)))
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("failed to decode certd json: %w", err)
	}

	return nil
}
