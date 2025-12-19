package api

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

// CertID VC verification
//
// This endpoint intentionally supports the in-app wallet + MetaMask flow used by cert-web:
// - The VC is signed via EIP-191 (personal_sign / signMessage)
// - The proof includes both the signed message and signature
//
// POST /api/v1/certid/vc/verify
// Body: { "vc": { ... } }

type certIDVCVerifyRequest struct {
	VC json.RawMessage `json:"vc"`
}

type certIDVCVerifyResponse struct {
	OK              bool     `json:"ok"`
	Errors          []string `json:"errors,omitempty"`
	RecoveredSigner string   `json:"recovered_signer,omitempty"`
	SubjectAddress  string   `json:"subject_address,omitempty"`

	// For debugging / integrations
	CanonicalMessage string `json:"canonical_message,omitempty"`
}

func (s *Server) handleVerifyCertIDVC(w http.ResponseWriter, r *http.Request) {
	var req certIDVCVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.VC) == 0 {
		s.respondError(w, http.StatusBadRequest, "vc is required")
		return
	}

	var vc map[string]any
	if err := json.Unmarshal(req.VC, &vc); err != nil {
		s.respondError(w, http.StatusBadRequest, "vc must be valid JSON")
		return
	}

	resp := certIDVCVerifyResponse{OK: false}

	// Extract proof fields
	proof, _ := vc["proof"].(map[string]any)
	proofMsg, _ := proof["message"].(string)
	proofSig, _ := proof["signature"].(string)
	if proofMsg == "" {
		resp.Errors = append(resp.Errors, "missing proof.message")
	}
	if proofSig == "" {
		resp.Errors = append(resp.Errors, "missing proof.signature")
	}

	// Extract subject address (EVM only for now)
	subject, _ := vc["credentialSubject"].(map[string]any)
	subjectAddress, _ := subject["address"].(string)
	resp.SubjectAddress = subjectAddress
	if subjectAddress == "" {
		resp.Errors = append(resp.Errors, "missing credentialSubject.address")
	}

	// Recompute canonical message used by cert-web issuer.
	canonicalMsg, err := certIDVCCanonicalMessage(vc)
	if err != nil {
		resp.Errors = append(resp.Errors, "failed to canonicalize vc: "+err.Error())
	} else {
		resp.CanonicalMessage = canonicalMsg
		// If caller included a message, ensure it matches what we'd expect.
		if proofMsg != "" && proofMsg != canonicalMsg {
			resp.Errors = append(resp.Errors, "proof.message does not match canonical message")
		}
	}

	if len(resp.Errors) > 0 {
		s.respondJSON(w, http.StatusOK, resp)
		return
	}

	// Recover signer from EIP-191 signature.
	// EIP-191: hash = keccak256("\x19Ethereum Signed Message:\n" + len(message) + message)
	hash := accounts.TextHash([]byte(proofMsg))

	sigBytes, sigErr := decodeAnySignature(proofSig)
	if sigErr != nil {
		resp.Errors = append(resp.Errors, sigErr.Error())
		s.respondJSON(w, http.StatusOK, resp)
		return
	}
	if len(sigBytes) != 65 {
		resp.Errors = append(resp.Errors, "signature must be 65 bytes")
		s.respondJSON(w, http.StatusOK, resp)
		return
	}
	// Normalize V to {0,1}
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	pub, err := crypto.SigToPub(hash, sigBytes)
	if err != nil {
		resp.Errors = append(resp.Errors, "failed to recover signer: "+err.Error())
		s.respondJSON(w, http.StatusOK, resp)
		return
	}
	recovered := crypto.PubkeyToAddress(*pub).Hex()
	resp.RecoveredSigner = recovered

	resp.OK = strings.EqualFold(recovered, subjectAddress)
	if !resp.OK {
		resp.Errors = append(resp.Errors, "recovered signer does not match credentialSubject.address")
	}

	s.respondJSON(w, http.StatusOK, resp)
}

func certIDVCCanonicalMessage(vc map[string]any) (string, error) {
	unsigned := map[string]any{}
	for k, v := range vc {
		if k == "proof" {
			continue
		}
		unsigned[k] = v
	}

	stable, err := stableJSON(unsigned)
	if err != nil {
		return "", err
	}
	return "CERT CertID Verifiable Credential\n" + stable, nil
}

// stableJSON returns a canonical JSON encoding with:
// - object keys sorted lexicographically
// - arrays preserved in-order
func stableJSON(v any) (string, error) {
	buf, err := stableJSONBytes(v)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func stableJSONBytes(v any) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return []byte("null"), nil
	case bool:
		if t {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case float64:
		// Numbers from encoding/json decode to float64. Re-marshal to preserve JSON number formatting.
		return json.Marshal(t)
	case string:
		return json.Marshal(t)
	case []any:
		parts := make([]json.RawMessage, 0, len(t))
		for _, it := range t {
			b, err := stableJSONBytes(it)
			if err != nil {
				return nil, err
			}
			parts = append(parts, b)
		}
		// Build: [a,b,c]
		out := []byte{'['}
		for i, p := range parts {
			if i > 0 {
				out = append(out, ',')
			}
			out = append(out, p...)
		}
		out = append(out, ']')
		return out, nil
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := []byte{'{'}
		for i, k := range keys {
			if i > 0 {
				out = append(out, ',')
			}
			kb, _ := json.Marshal(k)
			out = append(out, kb...)
			out = append(out, ':')
			vb, err := stableJSONBytes(t[k])
			if err != nil {
				return nil, err
			}
			out = append(out, vb...)
		}
		out = append(out, '}')
		return out, nil
	default:
		// Fallback: marshal then unmarshal to supported primitives
		b, err := json.Marshal(t)
		if err != nil {
			return nil, err
		}
		var anyv any
		if err := json.Unmarshal(b, &anyv); err != nil {
			return nil, err
		}
		return stableJSONBytes(anyv)
	}
}

func decodeAnySignature(sig string) ([]byte, error) {
	s := strings.TrimSpace(sig)
	if s == "" {
		return nil, errors.New("missing signature")
	}

	// Hex (0x...) or raw hex
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		b, err := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X"))
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	if len(s) == 130 {
		if b, err := hex.DecodeString(s); err == nil {
			return b, nil
		}
	}

	// Base64
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}

	return nil, errors.New("unsupported signature encoding (expected 0x-hex or base64)")
}
