package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

type Locus struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
}

type STRLoci map[string]Locus

type MatchResult struct {
	Score       float64 `json:"score"`
	MatchType   string  `json:"match_type"`
	MatchedLoci int     `json:"matched_loci"`
	TotalLoci   int     `json:"total_loci"`
	AlertLevel  string  `json:"alert_level"`
	SampleID    string  `json:"sample_id"`
}

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) HashProfile(profile STRLoci) string {
	keys := make([]string, 0, len(profile))
	for k := range profile {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		l := profile[k]
		fmt.Fprintf(h, "%s:%s:%s|", k, l.Value1, l.Value2)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (m *Matcher) Compare(query, candidate STRLoci) *MatchResult {
	matched := 0
	total := len(query)
	if total == 0 {
		return nil
	}

	for locus, qv := range query {
		if cv, ok := candidate[locus]; ok {
			if qv.Value1 == cv.Value1 && qv.Value2 == cv.Value2 {
				matched++
			} else if qv.Value1 == cv.Value2 && qv.Value2 == cv.Value1 {
				matched++
			} else if qv.Value1 == cv.Value1 || qv.Value1 == cv.Value2 || qv.Value2 == cv.Value1 || qv.Value2 == cv.Value2 {
				matched++
			}
		}
	}

	score := float64(matched) / float64(total)
	matchType := m.classifyMatch(score, total)
	alertLevel := m.classifyAlert(score)

	return &MatchResult{
		Score:       math.Round(score*10000) / 10000,
		MatchType:   matchType,
		MatchedLoci: matched,
		TotalLoci:   total,
		AlertLevel:  alertLevel,
	}
}

func (m *Matcher) classifyMatch(score float64, totalLoci int) string {
	if score >= 0.95 && totalLoci >= 10 {
		return "FULL_MATCH"
	} else if score >= 0.70 {
		return "PARTIAL"
	}
	return "FAMILIAL"
}

func (m *Matcher) classifyAlert(score float64) string {
	switch {
	case score >= 0.95:
		return "CRITICAL"
	case score >= 0.85:
		return "HIGH"
	case score >= 0.70:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func SerializeLoci(data json.RawMessage) (STRLoci, error) {
	var loci STRLoci
	if err := json.Unmarshal(data, &loci); err != nil {
		return nil, fmt.Errorf("unmarshal loci: %w", err)
	}
	return loci, nil
}
