// Package api provides CertID identity resolution handlers
// Per Cert ID Evolution spec: Universal Resolver Integration
package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// FullIdentity represents a complete CertID identity with badges and trust metrics
type FullIdentity struct {
	Address        string   `json:"address"`
	Handle         string   `json:"handle"`
	Name           string   `json:"name,omitempty"`
	Bio            string   `json:"bio,omitempty"`
	AvatarURL      string   `json:"avatar_url,omitempty"`
	MetadataURI    string   `json:"metadata_uri,omitempty"`
	IsVerified     bool     `json:"is_verified"`
	IsInstitutional bool    `json:"is_institutional"`
	TrustScore     int      `json:"trust_score"`
	EntityType     int      `json:"entity_type"`
	Badges         []Badge  `json:"badges"`
	IsKYC          bool     `json:"is_kyc"`
	IsAcademic     bool     `json:"is_academic"`
	IsCreator      bool     `json:"is_creator"`
	CreatedAt      int64    `json:"created_at,omitempty"`
}

// Badge represents a Soulbound Token badge
type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon"`
	AwardedAt   int64  `json:"awarded_at,omitempty"`
}

// Standard badge definitions
var standardBadges = map[string]Badge{
	"KYC_L1":           {ID: "KYC_L1", Name: "KYC Level 1", Icon: "🪪", Description: "Basic identity verification"},
	"KYC_L2":           {ID: "KYC_L2", Name: "KYC Level 2", Icon: "🛡️", Description: "Advanced identity verification"},
	"ACADEMIC_ISSUER":  {ID: "ACADEMIC_ISSUER", Name: "Academic Issuer", Icon: "🎓", Description: "Verified educational institution"},
	"VERIFIED_CREATOR": {ID: "VERIFIED_CREATOR", Name: "Verified Creator", Icon: "✨", Description: "Verified content creator"},
	"GOV_AGENCY":       {ID: "GOV_AGENCY", Name: "Government Agency", Icon: "🏛️", Description: "Verified government entity"},
	"LEGAL_ENTITY":     {ID: "LEGAL_ENTITY", Name: "Legal Entity", Icon: "⚖️", Description: "Verified legal organization"},
	"ISO_9001_CERTIFIED": {ID: "ISO_9001_CERTIFIED", Name: "ISO 9001", Icon: "📋", Description: "ISO 9001 quality certified"},
}

// handleGetFullIdentity returns complete identity with badges and trust score
// GET /api/v1/identity/{address}
func (s *Server) handleGetFullIdentity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := strings.ToLower(vars["address"])

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	identity := FullIdentity{
		Address:    address,
		Handle:     "Anonymous",
		TrustScore: 0,
		Badges:     []Badge{},
		EntityType: 0,
	}

	// Get profile from database if available
	if s.db != nil {
		if prof, err := s.db.GetProfile(ctx, address); err == nil && prof != nil {
			identity.Name = prof.Name
			identity.Handle = prof.Name
			if identity.Handle == "" {
				identity.Handle = truncateAddress(address)
			}
			identity.Bio = prof.Bio
			identity.AvatarURL = prof.AvatarURL
			identity.CreatedAt = prof.CreatedAt.Unix()
		}

		// Get credentials and map to badges
		if creds, err := s.db.GetCredentialsByUser(ctx, address); err == nil {
			for _, c := range creds {
				if c.Verified {
					identity.IsVerified = true
					// Map credential types to badges
					if badge, ok := mapCredentialToBadge(c.CredentialType); ok {
						identity.Badges = append(identity.Badges, badge)
					}
				}
			}
		}
	}

	// Derive convenience flags from badges
	for _, b := range identity.Badges {
		switch b.ID {
		case "KYC_L1", "KYC_L2":
			identity.IsKYC = true
		case "ACADEMIC_ISSUER":
			identity.IsAcademic = true
		case "VERIFIED_CREATOR":
			identity.IsCreator = true
		}
	}

	// Calculate trust score (after KYC flag is set)
	if s.db != nil {
		socialCount := 0
		if count, err := s.db.CountVerifiedSocialAccounts(ctx, address); err == nil {
			socialCount = count
		}
		if prof, err := s.db.GetProfile(ctx, address); err == nil && prof != nil {
			identity.TrustScore = calculateTrustScore(prof.CreatedAt, 0, identity.IsKYC, socialCount)
		}
	}

	s.respondJSON(w, http.StatusOK, identity)
}

// handleGetBadges returns badges for an address
// GET /api/v1/identity/{address}/badges
func (s *Server) handleGetBadges(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := strings.ToLower(vars["address"])

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	badges := []Badge{}

	if s.db != nil {
		if creds, err := s.db.GetCredentialsByUser(ctx, address); err == nil {
			for _, c := range creds {
				if c.Verified {
					if badge, ok := mapCredentialToBadge(c.CredentialType); ok {
						badges = append(badges, badge)
					}
				}
			}
		}
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address": address,
		"badges":  badges,
		"count":   len(badges),
	})
}

// handleGetTrustScore returns trust score for an address
// GET /api/v1/identity/{address}/trust-score
func (s *Server) handleGetTrustScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := strings.ToLower(vars["address"])

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	score := 0
	credentialCount := 0
	attestationCount := 0
	hasKYC := false

	socialCount := 0
	if s.db != nil {
		// Check for KYC credentials first
		if creds, err := s.db.GetCredentialsByUser(ctx, address); err == nil {
			credentialCount = len(creds)
			for _, c := range creds {
				if c.Verified {
					credType := strings.ToUpper(c.CredentialType)
					if credType == "KYC" || credType == "KYC_L1" || credType == "KYC_L2" || credType == "IDENTITY" {
						hasKYC = true
					}
				}
			}
		}

		// Count verified social accounts
		if count, err := s.db.CountVerifiedSocialAccounts(ctx, address); err == nil {
			socialCount = count
		}

		// Calculate trust score with KYC flag and social count
		if prof, err := s.db.GetProfile(ctx, address); err == nil && prof != nil {
			score = calculateTrustScore(prof.CreatedAt, attestationCount, hasKYC, socialCount)
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":           address,
		"trust_score":       score,
		"credential_count":  credentialCount,
		"attestation_count": attestationCount,
		"factors": map[string]int{
			"profile_age":  score / 2,
			"credentials":  credentialCount * 5,
			"attestations": attestationCount * 2,
		},
	})
}

// handleResolveHandle resolves a .cert handle to an address
// GET /api/v1/identity/resolve/{handle}
func (s *Server) handleResolveHandle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	handle := strings.ToLower(vars["handle"])

	// Remove .cert suffix if present
	handle = strings.TrimSuffix(handle, ".cert")

	s.logger.Info("Resolving handle", zap.String("handle", handle))

	// TODO: Implement handle registry lookup
	// For now, return not found
	s.respondError(w, http.StatusNotFound, "Handle not found")
}

// Helper functions

func truncateAddress(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

func calculateTrustScore(createdAt time.Time, attestationCount int, hasKYC bool, socialCount int) int {
	score := 0

	// KYC bonus: 50 points for verified identity (50% of max score)
	if hasKYC {
		score += 50
	}

	// Social verification bonus: 8 points per verified social account (up to 24 for 3 platforms)
	socialBonus := socialCount * 8
	if socialBonus > 24 {
		socialBonus = 24
	}
	score += socialBonus

	// Age bonus: up to 16 points for account age
	daysOld := int(time.Since(createdAt).Hours() / 24)
	if daysOld > 365 {
		score += 16
	} else if daysOld > 180 {
		score += 12
	} else if daysOld > 30 {
		score += 8
	} else if daysOld > 7 {
		score += 4
	}

	// Attestation bonus: 2 points per attestation, up to 10
	attBonus := attestationCount * 2
	if attBonus > 10 {
		attBonus = 10
	}
	score += attBonus

	return score
}

func mapCredentialToBadge(credType string) (Badge, bool) {
	switch strings.ToUpper(credType) {
	case "KYC", "KYC_L1", "IDENTITY":
		return standardBadges["KYC_L1"], true
	case "KYC_L2", "IDENTITY_ADVANCED":
		return standardBadges["KYC_L2"], true
	case "EDUCATION", "ACADEMIC", "UNIVERSITY":
		return standardBadges["ACADEMIC_ISSUER"], true
	case "CREATOR", "ARTIST", "DEVELOPER":
		return standardBadges["VERIFIED_CREATOR"], true
	case "GOVERNMENT", "GOV":
		return standardBadges["GOV_AGENCY"], true
	case "LEGAL", "BUSINESS", "COMPANY":
		return standardBadges["LEGAL_ENTITY"], true
	case "ISO", "ISO_9001", "QUALITY":
		return standardBadges["ISO_9001_CERTIFIED"], true
	default:
		return Badge{}, false
	}
}

