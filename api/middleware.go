package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Context keys
type contextKey string

const (
	UserAddressKey contextKey = "user_address"
	APIKeyInfoKey  contextKey = "api_key_info"
)

// loggingMiddleware logs all incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		s.logger.Info("Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates JWT tokens for protected endpoints
// Per CertID Section 2.2: Extract UserAddressKey from JWT
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return s.config.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract user address from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		address, ok := claims["address"].(string)
		if !ok || address == "" {
			http.Error(w, "Address not found in token", http.StatusUnauthorized)
			return
		}

		// Add address to context for use in handlers
		ctx := context.WithValue(r.Context(), UserAddressKey, address)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireAuth wraps a handler with authentication
func (s *Server) requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.authMiddleware(http.HandlerFunc(handler)).ServeHTTP(w, r)
	}
}

// getAuthenticatedAddress extracts the authenticated user address from context
func getAuthenticatedAddress(r *http.Request) string {
	if addr, ok := r.Context().Value(UserAddressKey).(string); ok {
		return addr
	}
	return ""
}

// apiKeyMiddleware validates API keys from X-API-Key header
// This provides higher rate limits and tracks usage per key
func (s *Server) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// No API key - continue with default rate limits
			next.ServeHTTP(w, r)
			return
		}

		// Validate API key format (must start with cert_)
		if !strings.HasPrefix(apiKey, "cert_") {
			http.Error(w, "Invalid API key format", http.StatusUnauthorized)
			return
		}

		// Hash the key to look up in database
		hash := sha256.Sum256([]byte(apiKey))
		keyHash := hex.EncodeToString(hash[:])

		if s.db == nil {
			// No database - can't validate keys
			next.ServeHTTP(w, r)
			return
		}

		// Validate the key
		keyInfo, err := s.db.ValidateAPIKey(r.Context(), keyHash)
		if err != nil {
			s.logger.Error("Failed to validate API key", zap.Error(err))
			http.Error(w, "Failed to validate API key", http.StatusInternalServerError)
			return
		}

		if keyInfo == nil {
			http.Error(w, "Invalid or expired API key", http.StatusUnauthorized)
			return
		}

		// Increment usage counter (async to not block request)
		go func() {
			if err := s.db.IncrementAPIKeyUsage(context.Background(), keyInfo.ID); err != nil {
				s.logger.Error("Failed to increment API key usage", zap.Error(err))
			}
		}()

		// Add key info to context
		ctx := context.WithValue(r.Context(), APIKeyInfoKey, keyInfo)
		// Also set the owner address for authorization
		ctx = context.WithValue(ctx, UserAddressKey, keyInfo.OwnerAddress)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
