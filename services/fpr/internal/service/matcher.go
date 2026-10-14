package service

import (
	"math"
	"math/rand"
)

type Minutia struct {
	X, Y  int
	Angle float64
	Type  string
}

type FPRMatcher struct {
	Threshold float64
}

func NewFPRMatcher(threshold float64) *FPRMatcher {
	return &FPRMatcher{Threshold: threshold}
}

func (m *FPRMatcher) extractMinutiae(imageData string) []Minutia {
	n := 10 + rand.Intn(20)
	minutiae := make([]Minutia, n)
	for i := range minutiae {
		minutiae[i] = Minutia{
			X:     rand.Intn(512),
			Y:     rand.Intn(512),
			Angle: rand.Float64() * 2 * math.Pi,
			Type:  "ridge_ending",
		}
	}
	return minutiae
}

func (m *FPRMatcher) compareMinutiae(a, b []Minutia) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	matches := 0
	for _, ma := range a {
		for _, mb := range b {
			dx := float64(ma.X - mb.X)
			dy := float64(ma.Y - mb.Y)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 40 && math.Abs(ma.Angle-mb.Angle) < 0.5 {
				matches++
				break
			}
		}
	}
	return float64(matches) / float64(len(a)) * 100
}

func (m *FPRMatcher) Verify(imageData string, enrolledData []byte) float64 {
	probe := m.extractMinutiae(imageData)
	reference := m.extractMinutiae(string(enrolledData))
	return m.compareMinutiae(probe, reference)
}

type IdentificationResult struct {
	UserID string  `json:"user_id"`
	Score  float64 `json:"score"`
}

func (m *FPRMatcher) Identify(imageData string, templates []*Template) []IdentificationResult {
	probe := m.extractMinutiae(imageData)
	var results []IdentificationResult
	for _, t := range templates {
		reference := m.extractMinutiae(string(t.Data))
		score := m.compareMinutiae(probe, reference)
		if score >= m.Threshold {
			results = append(results, IdentificationResult{
				UserID: t.UserID,
				Score:  score,
			})
		}
	}
	return results
}
