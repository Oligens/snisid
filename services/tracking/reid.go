package tracking

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type TrackResult struct {
	SubjectID  string  `json:"subject_id"`
	MatchScore float64 `json:"match_score"`
	Database   string  `json:"database"`
	CameraID   string  `json:"camera_id,omitempty"`
	Timestamp  int64   `json:"timestamp,omitempty"`
	Anomalous  bool    `json:"anomalous"`
}

type CrossImageTracker struct {
	Registry       map[string]string    `json:"registry"`
	SubjectDB      map[string][]float64 `json:"-"` // subjectID -> feature vector
	mu             sync.RWMutex
	matchThreshold float64
	minSimilarity  float64
}

func NewCrossImageTracker() *CrossImageTracker {
	return &CrossImageTracker{
		Registry:       make(map[string]string),
		SubjectDB:      make(map[string][]float64),
		matchThreshold: 0.85,
		minSimilarity:  0.6,
	}
}

func (t *CrossImageTracker) ReIdentify(vector []float64) []TrackResult {
	logger.Info(context.Background(), "NSIM: running cross-image re-identification", zap.Int("vector_dim", len(vector)))

	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.SubjectDB) == 0 {
		return nil
	}

	type scored struct {
		subjectID string
		database  string
		score     float64
	}

	var results []scored

	for subjectID, dbVector := range t.SubjectDB {
		if len(dbVector) != len(vector) {
			continue
		}
		similarity := cosineSimilarity(vector, dbVector)
		if similarity >= t.minSimilarity {
			database := t.inferDatabase(subjectID)
			results = append(results, scored{
				subjectID: subjectID,
				database:  database,
				score:     similarity,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	topK := 5
	if len(results) > topK {
		results = results[:topK]
	}

	trackResults := make([]TrackResult, 0, len(results))
	for _, r := range results {
		anomalous := t.detectAnomaly(r, results)
		trackResults = append(trackResults, TrackResult{
			SubjectID:  r.subjectID,
			MatchScore: math.Round(r.score*100) / 100,
			Database:   r.database,
			Anomalous:  anomalous,
		})
	}

	if len(trackResults) > 0 {
		logger.Info(context.Background(), "NSIM: re-identification results",
			zap.Int("matches", len(trackResults)),
			zap.Float64("top_score", trackResults[0].MatchScore),
		)
	}

	return trackResults
}

func (t *CrossImageTracker) LinkInconsistencies(results []TrackResult) {
	if len(results) < 2 {
		return
	}

	bestScore := results[0].MatchScore
	for _, r := range results[1:] {
		if r.MatchScore > 0.9 && r.Database != results[0].Database {
			scoreDiff := bestScore - r.MatchScore
			if scoreDiff < 0.05 {
				logger.Warn(context.Background(), "NSIM_TRACK: high-confidence cross-database link",
					zap.String("subject", r.SubjectID),
					zap.String("db1", results[0].Database),
					zap.String("db2", r.Database),
					zap.Any("scores", []float64{bestScore, r.MatchScore}),
				)
			}
		}
	}
}

func (t *CrossImageTracker) detectAnomaly(result scored, allResults []scored) bool {
	if result.score >= t.matchThreshold {
		return false
	}

	avgScore := 0.0
	for _, r := range allResults {
		avgScore += r.score
	}
	avgScore /= float64(len(allResults))

	return result.score < avgScore*0.8
}

func (t *CrossImageTracker) RegisterSubject(subjectID string, vector []float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.SubjectDB[subjectID] = vector

	prefix := t.inferDatabase(subjectID)
	t.Registry[subjectID] = prefix
}

func (t *CrossImageTracker) RemoveSubject(subjectID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.SubjectDB, subjectID)
	delete(t.Registry, subjectID)
}

func (t *CrossImageTracker) inferDatabase(subjectID string) string {
	if len(subjectID) < 4 {
		return "UNKNOWN"
	}
	prefix := subjectID[:4]
	dbMap := map[string]string{
		"CIT_": "ONI",
		"TAX_": "DGI",
		"POL_": "PNH",
		"BIO_": "BIOMETRICS",
		"FPR_": "FPR",
	}
	if db, ok := dbMap[prefix]; ok {
		return db
	}
	return "UNKNOWN"
}

func (t *CrossImageTracker) AddFromDatabase(database string, subjects map[string][]float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for subjectID, vector := range subjects {
		t.SubjectDB[subjectID] = vector
		t.Registry[subjectID] = database
	}

	logger.Info(context.Background(), "NSIM: loaded subjects from database",
		zap.String("database", database),
		zap.Int("count", len(subjects)),
	)
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
