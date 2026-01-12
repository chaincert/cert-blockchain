package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ProposalInfo represents a governance proposal
type ProposalInfo struct {
	ID               string          `json:"id"`
	Messages         json.RawMessage `json:"messages,omitempty"`
	Status           string          `json:"status"`
	FinalTallyResult TallyResult     `json:"final_tally_result"`
	SubmitTime       string          `json:"submit_time"`
	DepositEndTime   string          `json:"deposit_end_time"`
	TotalDeposit     []Coin          `json:"total_deposit"`
	VotingStartTime  string          `json:"voting_start_time"`
	VotingEndTime    string          `json:"voting_end_time"`
	Metadata         string          `json:"metadata"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary"`
	Proposer         string          `json:"proposer"`
}

type TallyResult struct {
	YesCount        string `json:"yes_count"`
	AbstainCount    string `json:"abstain_count"`
	NoCount         string `json:"no_count"`
	NoWithVetoCount string `json:"no_with_veto_count"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type ProposalsResponse struct {
	Proposals  []ProposalInfo `json:"proposals"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// handleGetProposals returns all governance proposals (replaces stub in handlers_common.go)
func (s *Server) handleGetAllProposals(w http.ResponseWriter, r *http.Request) {
	// Query proposals from REST API
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals", getRESTBaseURL())
	
	resp, err := restClient.Get(url)
	if err != nil {
		s.logger.Warn("proposals query failed", zap.Error(err))
		// Return empty list on error
		s.respondJSON(w, http.StatusOK, ProposalsResponse{Proposals: []ProposalInfo{}})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Warn("failed to read proposals response", zap.Error(err))
		s.respondJSON(w, http.StatusOK, ProposalsResponse{Proposals: []ProposalInfo{}})
		return
	}

	var result ProposalsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		s.logger.Warn("failed to parse proposals response", zap.Error(err), zap.String("body", string(body)))
		s.respondJSON(w, http.StatusOK, ProposalsResponse{Proposals: []ProposalInfo{}})
		return
	}

	s.respondJSON(w, http.StatusOK, result)
}

// handleGetProposal returns a specific proposal by ID
func (s *Server) handleGetProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID := vars["proposal_id"]
	if proposalID == "" {
		s.respondError(w, http.StatusBadRequest, "proposal_id is required")
		return
	}

	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", getRESTBaseURL(), proposalID)
	
	resp, err := restClient.Get(url)
	if err != nil {
		s.logger.Warn("proposal query failed", zap.String("id", proposalID), zap.Error(err))
		s.respondError(w, http.StatusBadGateway, "Failed to query proposal")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		s.respondError(w, http.StatusNotFound, "Proposal not found")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.respondError(w, http.StatusBadGateway, "Failed to read response")
		return
	}

	var result struct {
		Proposal ProposalInfo `json:"proposal"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		s.respondError(w, http.StatusBadGateway, "Failed to parse response")
		return
	}

	s.respondJSON(w, http.StatusOK, result)
}

// handleGetProposalTally returns current tally for a proposal
func (s *Server) handleGetProposalTally(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID := vars["proposal_id"]
	if proposalID == "" {
		s.respondError(w, http.StatusBadRequest, "proposal_id is required")
		return
	}

	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/tally", getRESTBaseURL(), proposalID)
	
	resp, err := restClient.Get(url)
	if err != nil {
		s.respondError(w, http.StatusBadGateway, "Failed to query tally")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// handleGetGovParams returns governance module parameters
func (s *Server) handleGetGovParams(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/params/voting", getRESTBaseURL())

	resp, err := restClient.Get(url)
	if err != nil {
		// Return default params on error
		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"voting_params": map[string]interface{}{
				"voting_period": "172800s", // 2 days
			},
			"deposit_params": map[string]interface{}{
				"min_deposit": []map[string]string{
					{"denom": "ucert", "amount": "10000000"},
				},
				"max_deposit_period": "172800s",
			},
			"tally_params": map[string]interface{}{
				"quorum":         "0.334000000000000000",
				"threshold":      "0.500000000000000000",
				"veto_threshold": "0.334000000000000000",
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

// VoteRequest represents a vote request
type VoteRequest struct {
	Voter      string `json:"voter"`
	ProposalID string `json:"proposal_id"`
	Option     string `json:"option"` // VOTE_OPTION_YES, VOTE_OPTION_NO, VOTE_OPTION_ABSTAIN, VOTE_OPTION_NO_WITH_VETO
}

// handleVoteOnProposal creates an unsigned vote transaction
func (s *Server) handleVoteOnProposal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID := vars["proposal_id"]

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Voter == "" {
		s.respondError(w, http.StatusBadRequest, "voter address is required")
		return
	}

	// Convert option string to vote option number
	optionNum := "1" // Default YES
	switch req.Option {
	case "VOTE_OPTION_YES", "yes", "1":
		optionNum = "1"
	case "VOTE_OPTION_ABSTAIN", "abstain", "2":
		optionNum = "2"
	case "VOTE_OPTION_NO", "no", "3":
		optionNum = "3"
	case "VOTE_OPTION_NO_WITH_VETO", "no_with_veto", "4":
		optionNum = "4"
	}

	unsignedTx := map[string]interface{}{
		"body": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"@type":       "/cosmos.gov.v1.MsgVote",
					"proposal_id": proposalID,
					"voter":       req.Voter,
					"option":      optionNum,
					"metadata":    "",
				},
			},
			"memo":           "",
			"timeout_height": "0",
		},
		"auth_info": map[string]interface{}{
			"signer_infos": []interface{}{},
			"fee": map[string]interface{}{
				"amount":    []map[string]string{{"denom": "ucert", "amount": "5000"}},
				"gas_limit": "150000",
			},
		},
		"signatures": []string{},
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"unsigned_tx": unsignedTx,
		"message":     "Sign this transaction with your wallet and broadcast it",
	})
}

// CreateProposalRequest represents a proposal creation request
type CreateProposalRequest struct {
	Proposer    string   `json:"proposer"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Deposit     string   `json:"deposit"` // in ucert
	Messages    []string `json:"messages,omitempty"` // Optional: execution messages
}

// handleCreateProposal creates an unsigned proposal transaction
func (s *Server) handleCreateProposal(w http.ResponseWriter, r *http.Request) {
	var req CreateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Proposer == "" || req.Title == "" || req.Summary == "" {
		s.respondError(w, http.StatusBadRequest, "proposer, title, and summary are required")
		return
	}

	deposit := req.Deposit
	if deposit == "" {
		deposit = "10000000" // Default 10 CERT minimum deposit
	}

	// Create a text proposal (simplest type)
	unsignedTx := map[string]interface{}{
		"body": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"@type":            "/cosmos.gov.v1.MsgSubmitProposal",
					"messages":         []interface{}{},
					"initial_deposit":  []map[string]string{{"denom": "ucert", "amount": deposit}},
					"proposer":         req.Proposer,
					"metadata":         req.Description,
					"title":            req.Title,
					"summary":          req.Summary,
					"expedited":        false,
				},
			},
			"memo":           "",
			"timeout_height": "0",
		},
		"auth_info": map[string]interface{}{
			"signer_infos": []interface{}{},
			"fee": map[string]interface{}{
				"amount":    []map[string]string{{"denom": "ucert", "amount": "25000"}},
				"gas_limit": "300000",
			},
		},
		"signatures": []string{},
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"unsigned_tx": unsignedTx,
		"message":     "Sign this transaction with your wallet and broadcast it",
	})
}

// handleGetVotes returns votes for a proposal
func (s *Server) handleGetVotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proposalID := vars["proposal_id"]
	if proposalID == "" {
		s.respondError(w, http.StatusBadRequest, "proposal_id is required")
		return
	}

	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/votes", getRESTBaseURL(), proposalID)

	resp, err := restClient.Get(url)
	if err != nil {
		s.respondJSON(w, http.StatusOK, map[string]interface{}{"votes": []interface{}{}})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

