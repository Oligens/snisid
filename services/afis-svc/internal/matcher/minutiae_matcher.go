package matcher

import (
	"context"
	"math"
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

type MinutiaPoint struct {
	X       int
	Y       int
	Angle   float64
	Type    string
	Quality int16
}

type MinutiaeMatcher struct {
	mu        sync.RWMutex
	threshold float64
}

func NewMinutiaeMatcher(threshold float64) *MinutiaeMatcher {
	return &MinutiaeMatcher{
		threshold: threshold,
	}
}

func (m *MinutiaeMatcher) CompareMinutiae(ctx context.Context, captures []domain.FingerprintCapture, subjectID uuid.UUID) (float64, error) {
	queryPoints := extractSimulatedMinutiae(captures)
	var candidatePoints []MinutiaPoint
	for range 40 {
		candidatePoints = append(candidatePoints, MinutiaPoint{
			X:       rand.Intn(500),
			Y:       rand.Intn(500),
			Angle:   rand.Float64() * 2 * math.Pi,
			Type:    "RIDGE_ENDING",
			Quality: int16(rand.Intn(40) + 60),
		})
	}

	score := computeMatchScore(queryPoints, candidatePoints)
	return score, nil
}

func extractSimulatedMinutiae(captures []domain.FingerprintCapture) []MinutiaPoint {
	points := make([]MinutiaPoint, 0)
	for range captures {
		for i := 0; i < 30+rand.Intn(20); i++ {
			points = append(points, MinutiaPoint{
				X:       rand.Intn(500),
				Y:       rand.Intn(500),
				Angle:   rand.Float64() * 2 * math.Pi,
				Type:    []string{"RIDGE_ENDING", "BIFURCATION", "DOT"}[rand.Intn(3)],
				Quality: int16(rand.Intn(40) + 60),
			})
		}
	}
	return points
}

func computeMatchScore(query, candidate []MinutiaPoint) float64 {
	if len(query) == 0 || len(candidate) == 0 {
		return 0
	}

	matches := 0
	for _, qp := range query {
		for _, cp := range candidate {
			dist := math.Sqrt(float64((qp.X-cp.X)*(qp.X-cp.X) + (qp.Y-cp.Y)*(qp.Y-cp.Y)))
			if dist < 20 && qp.Type == cp.Type {
				angleDiff := math.Abs(qp.Angle - cp.Angle)
				if angleDiff < math.Pi/6 || angleDiff > 2*math.Pi-math.Pi/6 {
					matches++
					break
				}
			}
		}
	}

	score := float64(matches) / float64(len(query))
	if score > 1.0 {
		score = 1.0
	}
	return score
}
