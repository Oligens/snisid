package engine

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/snisid/platform/services/bio-adn/pkg/models"
)

// ── EventPublisher interface (Kafka forward) ───────────────────────────────

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

// ── Cache interface (Redis hits) ───────────────────────────────────────────

type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

// ── DNAMatcher — high-level matching pipeline ──────────────────────────────

type DNAMatcher struct {
	db      models.Database
	matcher *Matcher
	cache   Cache
	events  EventPublisher
}

func NewDNAMatcher(db models.Database, cache Cache, events EventPublisher) *DNAMatcher {
	return &DNAMatcher{
		db:      db,
		matcher: NewMatcher(),
		cache:   cache,
		events:  events,
	}
}

// SearchProfile runs the full pipeline: cache → DB candidates → LR scoring → Kafka hits
func (m *DNAMatcher) SearchProfile(ctx context.Context, profile models.DNAProfile, threshold float64) ([]MatchResult, error) {
	if threshold == 0 {
		threshold = 0.85
	}

	cacheKey := m.buildCacheKey(profile)
	if cached, err := m.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if results, ok := cached.([]MatchResult); ok {
			return results, nil
		}
	}

	profiles, _, err := m.db.SearchDNAProfiles(ctx, profile.IndexType, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("candidate search: %w", err)
	}

	var results []MatchResult
	for _, candidate := range profiles {
		queryLoci := m.hashToLoci(profile.LociHash)
		candLoci := m.hashToLoci(candidate.LociHash)

		lrResult := m.matcher.CalculateLR(queryLoci, candLoci, nil)
		if lrResult == nil {
			continue
		}
		if lrResult.LikelihoodRatio < threshold {
			continue
		}

		alertLevel := classifyAlertLR(lrResult.LikelihoodRatio)
		matchType := classifyMatchLR(lrResult.LikelihoodRatio)

		results = append(results, MatchResult{
			Score:       math.Round(lrResult.LikelihoodRatio*100) / 100,
			MatchType:   matchType,
			MatchedLoci: lrResult.MatchedLoci,
			TotalLoci:   lrResult.TotalLoci,
			AlertLevel:  alertLevel,
			SampleID:    candidate.SampleID,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > 100 {
		results = results[:100]
	}

	if len(results) > 0 {
		_ = m.cache.Set(ctx, cacheKey, results, 5*time.Minute)
	}

	if len(results) > 0 {
		_ = m.events.Publish(ctx, "snisid.bio.hits", map[string]any{
			"query_sample_id": profile.SampleID,
			"hits":            results,
			"timestamp":       time.Now().UnixMilli(),
		})
	}

	return results, nil
}

func (m *DNAMatcher) buildCacheKey(profile models.DNAProfile) string {
	return fmt.Sprintf("match:%s:%s", profile.IndexType, profile.LociHash)
}

func (m *DNAMatcher) hashToLoci(hash string) STRLoci {
	if hash == "" {
		return nil
	}
	loci := make(STRLoci)
	if len(hash) >= 4 {
		loci["D3S1358"] = Locus{Value1: hash[:2], Value2: hash[2:4]}
	}
	return loci
}

// ── AuditableMatchResult — signed hit for immutable audit ledger ───────────

type AuditableMatchResult struct {
	HitID         string    `json:"hit_id"`
	QuerySampleID string    `json:"query_sample_id"`
	MatchSampleID string    `json:"match_sample_id"`
	MatchType     string    `json:"match_type"`
	Confidence    float64   `json:"confidence"`
	MatchedLoci   int       `json:"matched_loci"`
	TotalLoci     int       `json:"total_loci"`
	MatchedAt     time.Time `json:"matched_at"`
	OfficerNIU    string    `json:"officer_niu"`
	Signature     string    `json:"signature"` // hex(ECDSA-P256)
}

func SignAuditEntry(entry *AuditableMatchResult, privateKeyPEM string) error {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse EC key: %w", err)
	}

	data := fmt.Sprintf("%s|%s|%s|%s|%s|%d|%d",
		entry.HitID, entry.QuerySampleID, entry.MatchSampleID,
		entry.MatchType, entry.OfficerNIU,
		entry.MatchedLoci, entry.TotalLoci,
	)
	digest := sha256.Sum256([]byte(data))

	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		return fmt.Errorf("ecdsa sign: %w", err)
	}
	sig := append(r.Bytes(), s.Bytes()...)
	entry.Signature = hex.EncodeToString(sig)
	return nil
}

func VerifyAuditSignature(entry *AuditableMatchResult, publicKeyPEM string) (bool, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, fmt.Errorf("failed to decode PEM block")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := pubAny.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an ECDSA public key")
	}

	data := fmt.Sprintf("%s|%s|%s|%s|%s|%d|%d",
		entry.HitID, entry.QuerySampleID, entry.MatchSampleID,
		entry.MatchType, entry.OfficerNIU,
		entry.MatchedLoci, entry.TotalLoci,
	)
	digest := sha256.Sum256([]byte(data))

	sigBytes, err := hex.DecodeString(entry.Signature)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}
	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])

	return ecdsa.Verify(pub, digest[:], r, s), nil
}

func classifyMatchLR(lr float64) string {
	logLR := math.Log10(lr)
	switch {
	case logLR >= 6:
		return "FULL_MATCH"
	case logLR >= 3:
		return "PARTIAL"
	case logLR >= 1:
		return "FAMILIAL"
	default:
		return "NO_MATCH"
	}
}

func classifyAlertLR(lr float64) string {
	logLR := math.Log10(lr)
	switch {
	case logLR >= 10:
		return "CRITICAL"
	case logLR >= 6:
		return "HIGH"
	case logLR >= 3:
		return "MEDIUM"
	default:
		return "LOW"
	}
}
