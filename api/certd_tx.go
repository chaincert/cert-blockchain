package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type certdTxResponse struct {
	TxHash    string `json:"txhash"`
	Code      int    `json:"code"`
	Codespace string `json:"codespace"`
	RawLog    string `json:"raw_log"`
	Log       string `json:"log"`
	Timestamp string `json:"timestamp"`
	Logs      []struct {
		Events []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
		} `json:"events"`
	} `json:"logs"`
}

type certdTxExecError struct {
	Err    error
	Output string
	Tx     certdTxResponse
}

func (e *certdTxExecError) Error() string {
	if e == nil {
		return "certd tx error"
	}
	if e.Tx.RawLog != "" {
		return fmt.Sprintf("certd tx failed: %v: %s", e.Err, e.Tx.RawLog)
	}
	return fmt.Sprintf("certd tx failed: %v: %s", e.Err, e.Output)
}

func (e *certdTxExecError) Unwrap() error { return e.Err }

func (s *Server) execCertdTxJSON(ctx context.Context, out any, txArgs ...string) ([]byte, error) {
	// Run inside docker: `certd tx ... --node tcp://localhost:26657 --chain-id ... --from ... --output json`
	args := []string{"tx"}
	args = append(args, txArgs...)
	args = append(args,
		"--node", s.config.TxNode,
		"--chain-id", s.config.ChainID,
		"--from", s.config.TxFrom,
		"--keyring-backend", s.config.TxKeyringBackend,
		"--home", s.config.TxHome,
		"--broadcast-mode", s.config.TxBroadcastMode,
		"--output", "json",
		"--yes",
	)

	if strings.TrimSpace(s.config.TxGas) != "" {
		args = append(args, "--gas", strings.TrimSpace(s.config.TxGas))
	}
	if strings.TrimSpace(s.config.TxFees) != "" {
		args = append(args, "--fees", strings.TrimSpace(s.config.TxFees))
	} else if strings.TrimSpace(s.config.TxGasPrices) != "" {
		args = append(args, "--gas-prices", strings.TrimSpace(s.config.TxGasPrices))
	}

	cmd := exec.CommandContext(ctx, "docker", append([]string{"exec", "certd", "certd"}, args...)...)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		// Best-effort: still try to decode JSON from output so callers can surface raw_log.
		var raw json.RawMessage
		if v, ok := extractFirstJSONObject[json.RawMessage](string(buf)); ok {
			raw = v
		} else {
			raw = json.RawMessage(strings.TrimSpace(string(buf)))
		}

		var tx certdTxResponse
		if uerr := json.Unmarshal(raw, &tx); uerr == nil {
			_ = json.Unmarshal(raw, out)
			return buf, &certdTxExecError{Err: err, Output: string(buf), Tx: tx}
		}
		return buf, fmt.Errorf("certd tx failed: %w: %s", err, string(buf))
	}

	// certd can print extra lines; extract first JSON object.
	var raw json.RawMessage
	if v, ok := extractFirstJSONObject[json.RawMessage](string(buf)); ok {
		raw = v
	} else {
		raw = json.RawMessage(strings.TrimSpace(string(buf)))
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return buf, fmt.Errorf("failed to decode certd tx json: %w", err)
	}

	return buf, nil
}

func certdTxTimestampUnix(ts string) int64 {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return time.Now().Unix()
	}
	// Cosmos SDK uses RFC3339 or RFC3339Nano depending on build.
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t.Unix()
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.Unix()
	}
	return time.Now().Unix()
}

func findTxEventAttribute(tx certdTxResponse, wantKey string) (string, bool) {
	for _, l := range tx.Logs {
		for _, e := range l.Events {
			for _, a := range e.Attributes {
				k := normalizeEventField(a.Key)
				if k != wantKey {
					continue
				}
				v := normalizeEventField(a.Value)
				return v, true
			}
		}
	}
	return "", false
}

func normalizeEventField(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Some tendermint JSON renderings base64-encode event keys/values.
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		if isMostlyASCII(decoded) {
			return string(decoded)
		}
	}
	return s
}

func isMostlyASCII(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	// Heuristic: treat as printable ASCII if all bytes are within [9..13] or [32..126].
	for _, c := range b {
		if c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}
