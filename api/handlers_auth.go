package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/ripemd160"
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
	Signature json.RawMessage `json:"signature"`
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
		s.logger.Warn("auth verify failed: invalid json", zap.Error(err))
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "Invalid request body"})
		return
	}

	req.Address = strings.TrimSpace(req.Address)
	req.Address = strings.TrimSpace(req.Address)
	req.Nonce = strings.TrimSpace(req.Nonce)
	
	// Check if signature is present
	if len(req.Signature) == 0 {
		s.logger.Warn("auth verify failed: missing signature")
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "signature is required"})
		return
	}
	// Address and Nonce check
	if req.Address == "" || req.Nonce == "" {
		s.logger.Warn("auth verify failed: missing fields", zap.Any("req", req))
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

	// Normalize address: support both 0x... (EVM) and cert1... (bech32) formats
	// For bech32 addresses, convert to 0x format for signature verification
	evmAddress := req.Address
	originalAddress := req.Address // Preserve original format for JWT

	if strings.HasPrefix(req.Address, "cert1") {
		// Decode bech32 to get raw address bytes
		_, addrBytes, err := bech32.DecodeAndConvert(req.Address)
		if err != nil {
			s.logger.Warn("auth verify failed: invalid bech32", zap.Error(err), zap.String("addr", req.Address))
			s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "invalid bech32 address"})
			return
		}
		if len(addrBytes) != 20 {
			s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "invalid address length"})
			return
		}
		evmAddress = fmt.Sprintf("0x%x", addrBytes)
	} else if !(strings.HasPrefix(req.Address, "0x") || strings.HasPrefix(req.Address, "0X")) {
		s.logger.Warn("auth verify failed: invalid address format", zap.String("addr", req.Address))
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "address must be 0x... or cert1... format"})
		return
	}

	// Try to parse as Keplr JSON signature with pubkey
	type keplrPubKey struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	type keplrSignature struct {
		Signature string      `json:"signature"`
		PubKey    keplrPubKey `json:"pub_key"`
	}

	var keplrSig keplrSignature
	var sigBytes []byte
	var pubKeyBytes []byte

	// Parse Signature: it can be a JSON string (EVM/standard) or a JSON object (Keplr)
	var sigString string
	
	// First, try to see if it's a JSON object (Keplr struct)
	if err := json.Unmarshal(req.Signature, &keplrSig); err == nil && keplrSig.Signature != "" {
		// Keplr JSON format with pubkey
		s.logger.Debug("parsed Keplr signature object", zap.String("pubkey_type", keplrSig.PubKey.Type))
		sigBytes, err = base64.StdEncoding.DecodeString(keplrSig.Signature)
		if err != nil {
			s.logger.Warn("auth verify failed: invalid signature encoding (keplr)", zap.Error(err))
			s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "invalid signature encoding"})
			return
		}
		if keplrSig.PubKey.Value != "" {
			pubKeyBytes, _ = base64.StdEncoding.DecodeString(keplrSig.PubKey.Value)
		}
	} else {
		// Try to unmarshal as a string (EVM case)
		if err := json.Unmarshal(req.Signature, &sigString); err != nil {
			// If it's not a valid JSON string, maybe it's the raw string? (Unlikely if Decode succeeded, but safe fallback)
			sigString = string(req.Signature)
		}
		sigString = strings.TrimSpace(sigString)
		
		// Standard signature format
		sigBytes, err = decodeAnySignature(sigString)
		if err != nil {
			s.logger.Warn("signature decode failed", zap.Error(err), zap.String("sig_prefix", sigString[:min(20, len(sigString))]))
			s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: "invalid signature"})
			return
		}
	}
	s.logger.Debug("signature decoded", zap.Int("len", len(sigBytes)), zap.Int("pubkey_len", len(pubKeyBytes)))

	// Handle different signature lengths:
	// - 65 bytes: EIP-191 with recovery byte (MetaMask, ethers.js)
	// - 64 bytes: Cosmos-style without recovery byte (Keplr signArbitrary)
	if len(sigBytes) != 65 && len(sigBytes) != 64 {
		s.logger.Warn("signature wrong length", zap.Int("got", len(sigBytes)))
		s.respondJSON(w, http.StatusBadRequest, authVerifyResponse{OK: false, Error: fmt.Sprintf("signature must be 64 or 65 bytes, got %d", len(sigBytes))})
		return
	}

	hash := accounts.TextHash([]byte(entry.Message))
	var recovered string

	// If we have a pubkey, verify using direct verification instead of recovery
	if len(pubKeyBytes) == 33 {
		// Compressed secp256k1 pubkey from Keplr
		// Verify signature against ADR-036 hash
		adr036Hash := adr036SignDocHash(originalAddress, []byte(entry.Message))

		// Verify signature directly
		sigValid := crypto.VerifySignature(pubKeyBytes, adr036Hash, sigBytes[:64])
		if !sigValid {
			s.logger.Warn("direct signature verification failed")
			s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signature verification failed"})
			return
		}

		// Keplr uses Cosmos address derivation: SHA256 + RIPEMD160 of pubkey
		// (not Ethermint's Keccak256)
		cosmosAddrBytes := cosmosAddressFromPubkey(pubKeyBytes)
		cosmosAddr, err := bech32.ConvertAndEncode("cert", cosmosAddrBytes)
		if err != nil {
			s.logger.Warn("failed to encode cosmos address", zap.Error(err))
			s.respondJSON(w, http.StatusInternalServerError, authVerifyResponse{OK: false, Error: "internal error"})
			return
		}

		s.logger.Debug("verified with pubkey", zap.String("cosmosAddr", cosmosAddr), zap.String("expected", originalAddress))

		// Compare cosmos addresses (bech32)
		if !strings.EqualFold(cosmosAddr, originalAddress) {
			s.logger.Warn("pubkey address mismatch", zap.String("recovered", cosmosAddr), zap.String("expected", originalAddress))
			s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signer does not match address"})
			return
		}

		// Set recovered for JWT generation (use bech32 address)
		recovered = evmAddress // Token will use EVM address for compatibility
	} else if len(sigBytes) == 65 {
		// Standard EIP-191 signature with recovery byte
		if sigBytes[64] >= 27 {
			sigBytes[64] -= 27
		}
		pub, err := crypto.SigToPub(hash, sigBytes)
		if err != nil {
			s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signature verification failed"})
			return
		}
		recovered = crypto.PubkeyToAddress(*pub).Hex()
	} else {
		// 64-byte signature (Cosmos/Keplr) - try ADR-036 hash first, then EIP-191
		var matched bool
		var triedAddrs []string

		// Try ADR-036 hash (Keplr signArbitrary) - requires original bech32 address
		adr036Hash := adr036SignDocHash(originalAddress, []byte(entry.Message))
		for _, v := range []byte{0, 1} {
			sig65 := make([]byte, 65)
			copy(sig65, sigBytes)
			sig65[64] = v
			pub, err := crypto.SigToPub(adr036Hash, sig65)
			if err != nil {
				continue
			}
			addr := crypto.PubkeyToAddress(*pub).Hex()
			triedAddrs = append(triedAddrs, fmt.Sprintf("adr036-v%d:%s", v, addr))
			if strings.EqualFold(addr, evmAddress) {
				recovered = addr
				matched = true
				s.logger.Debug("ADR-036 signature matched", zap.String("addr", addr))
				break
			}
		}

		// Fallback: try EIP-191 hash
		if !matched {
			for _, v := range []byte{0, 1} {
				sig65 := make([]byte, 65)
				copy(sig65, sigBytes)
				sig65[64] = v
				pub, err := crypto.SigToPub(hash, sig65)
				if err != nil {
					continue
				}
				addr := crypto.PubkeyToAddress(*pub).Hex()
				triedAddrs = append(triedAddrs, fmt.Sprintf("eip191-v%d:%s", v, addr))
				if strings.EqualFold(addr, evmAddress) {
					recovered = addr
					matched = true
					break
				}
			}
		}

		if !matched {
			s.logger.Warn("signature verification failed - no address match",
				zap.String("expected", evmAddress),
				zap.Strings("recovered", triedAddrs))
			s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signature verification failed"})
			return
		}
	}

	// Compare against the normalized EVM address
	if !strings.EqualFold(recovered, evmAddress) {
		s.respondJSON(w, http.StatusUnauthorized, authVerifyResponse{OK: false, Error: "signer does not match address"})
		return
	}

	// Issue JWT (default 12 hours)
	// Store the original address format (could be bech32 or EVM) for consistency
	exp := time.Now().Add(12 * time.Hour)
	claims := jwt.MapClaims{
		"address": originalAddress,
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

	s.respondJSON(w, http.StatusOK, authVerifyResponse{OK: true, Address: originalAddress, Token: signed, ExpiresAt: exp.Unix()})
}


// adr036SignDocHash computes the hash for ADR-036 arbitrary message signing.
// This matches what Keplr's signArbitrary produces.
// Format: amino-encoded SignDoc with a MsgSignData containing the signer and data.
func adr036SignDocHash(signer string, data []byte) []byte {
	// ADR-036 uses amino encoding for SignDoc
	// The structure is:
	// SignDoc { chain_id: "", account_number: "0", sequence: "0", fee: {gas: "0", amount: []}, msgs: [MsgSignData], memo: "" }
	// MsgSignData { signer: <bech32>, data: <base64> }

	// Build the canonical JSON (sorted keys, no spaces)
	dataB64 := base64.StdEncoding.EncodeToString(data)

	// MsgSignData amino type
	msg := map[string]interface{}{
		"type": "sign/MsgSignData",
		"value": map[string]interface{}{
			"data":   dataB64,
			"signer": signer,
		},
	}

	signDoc := map[string]interface{}{
		"account_number": "0",
		"chain_id":       "",
		"fee": map[string]interface{}{
			"amount": []interface{}{},
			"gas":    "0",
		},
		"memo": "",
		"msgs": []interface{}{msg},
		"sequence": "0",
	}

	// Marshal to canonical JSON (sorted keys)
	jsonBytes := marshalCanonicalJSON(signDoc)

	// SHA256 hash
	hash := sha256.Sum256(jsonBytes)
	return hash[:]
}

// marshalCanonicalJSON produces canonical JSON with sorted keys
func marshalCanonicalJSON(v interface{}) []byte {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString("{")
		for i, k := range keys {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(`"`)
			b.WriteString(k)
			b.WriteString(`":`)
			b.Write(marshalCanonicalJSON(val[k]))
		}
		b.WriteString("}")
		return []byte(b.String())
	case []interface{}:
		var b strings.Builder
		b.WriteString("[")
		for i, item := range val {
			if i > 0 {
				b.WriteString(",")
			}
			b.Write(marshalCanonicalJSON(item))
		}
		b.WriteString("]")
		return []byte(b.String())
	case string:
		escaped, _ := json.Marshal(val)
		return escaped
	default:
		result, _ := json.Marshal(val)
		return result
	}
}


// cosmosAddressFromPubkey computes Cosmos-style address from compressed secp256k1 pubkey
// Uses SHA256 + RIPEMD160 (not Keccak256 like Ethereum)
func cosmosAddressFromPubkey(pubkey []byte) []byte {
	sha := sha256.Sum256(pubkey)
	rip := ripemd160.New()
	rip.Write(sha[:])
	return rip.Sum(nil)
}
