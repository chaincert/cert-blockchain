package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
)

// Minimal EIP-191 auth (challenge -> signature -> JWT)
// This is designed for browser wallets (MetaMask) and cert-web in-app wallets.

type authChallengeEntry struct {
	Address   string
	Message   string
	ExpiresAt time.Time
	Used      bool
}

var (
	authMu         sync.Mutex
	authChallenges = map[string]*authChallengeEntry{} // nonce -> entry
)

type authChallengeResponse struct {
	Address   string `json:"address"`
	Nonce     string `json:"nonce"`
	Challenge string `json:"challenge"`
	ExpiresAt int64  `json:"expires_at"`
}

func (s *Server) handleAuthChallenge(w http.ResponseWriter, r *http.Request) {
	address := strings.TrimSpace(r.URL.Query().Get("address"))
	if address == "" {
		s.respondError(w, http.StatusBadRequest, "address is required")
		return
	}

	nonceBytes := make([]byte, 16)
	_, _ = rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)

	expiresAt := time.Now().Add(5 * time.Minute)
	challenge := strings.Join([]string{
		"CERT Authentication",
		"\n\nAddress: " + address,
		"\nNonce: " + nonce,
		"\nIssued At: " + time.Now().UTC().Format(time.RFC3339),
		"\n\nBy signing, you authorize this app to obtain a short-lived JWT for CERT APIs.",
	}, "")

	authMu.Lock()
	authChallenges[nonce] = &authChallengeEntry{Address: address, Message: challenge, ExpiresAt: expiresAt, Used: false}
	authMu.Unlock()

	s.respondJSON(w, http.StatusOK, authChallengeResponse{
		Address:   address,
		Nonce:     nonce,
		Challenge: challenge,
		ExpiresAt: expiresAt.Unix(),
	})
}

type authVerifyRequest struct {
	Address   string `json:"address"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

type authVerifyResponse struct {
	OK        bool   `json:"ok"`
	Address   string `json:"address,omitempty"`
	Token     string `json:"token,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *Server) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	var req authVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "Invalid request body"})
		return
	}

	req.Address = strings.TrimSpace(req.Address)
	req.Nonce = strings.TrimSpace(req.Nonce)
	req.Signature = strings.TrimSpace(req.Signature)
	if req.Address == "" || req.Nonce == "" || req.Signature == "" {
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "address, nonce, and signature are required"})
		return
	}

	authMu.Lock()
	entry := authChallenges[req.Nonce]
	if entry == nil {
		authMu.Unlock()
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "invalid nonce"})
		return
	}
	// One-time use
	if entry.Used {
		authMu.Unlock()
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "nonce already used"})
		return
	}
	if time.Now().After(entry.ExpiresAt) {
		authMu.Unlock()
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "nonce expired"})
		return
	}
	// Bind nonce to address
	if !strings.EqualFold(entry.Address, req.Address) {
		authMu.Unlock()
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "nonce not issued for this address"})
		return
	}
	entry.Used = true
	authMu.Unlock()

	// Verify EIP-191 signature for EVM-style addresses.
	// (Cosmos/bech32 wallet auth is not implemented here.)
	if !(strings.HasPrefix(req.Address, "0x") || strings.HasPrefix(req.Address, "0X")) {
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "only 0x... addresses supported for auth"})
		return
	}

	sigBytes, err := decodeAnySignature(req.Signature)
	if err != nil {
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "invalid signature"})
		return
	}
	if len(sigBytes) != 65 {
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "signature must be 65 bytes"})
		return
	}
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	hash := accounts.TextHash([]byte(entry.Message))
	pub, err := crypto.SigToPub(hash, sigBytes)
	if err != nil {
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signature verification failed"})
		return
	}
	recovered := crypto.PubkeyToAddress(*pub).Hex()
	if !strings.EqualFold(recovered, req.Address) {
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signer does not match address"})
		return
	}

	// Issue JWT (default 12 hours)
	exp := time.Now().Add(12 * time.Hour)
	claims := jwt.MapClaims{
		"address": req.Address,
		"nonce":   req.Nonce,
		"iat":     time.Now().Unix(),
		"exp":     exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.config.JWTSecret)
	if err != nil {
		s.respondJSON(w, http.StatusInternalServerError, authVerifyResponse{OK: false, Error: "failed to sign token"})
		return
	}

	s.respondJSON(w, http.StatusOK, authVerifyResponse{OK: true, Address: req.Address, Token: signed, ExpiresAt: exp.Unix()})
}
