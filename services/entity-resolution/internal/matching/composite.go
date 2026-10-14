package matching

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/entity-resolution/internal/models"
	"gorm.io/gorm"
)

type MatchMethod string

const (
	MethodExact     MatchMethod = "exact"
	MethodFuzzy     MatchMethod = "fuzzy"
	MethodPhonetic  MatchMethod = "phonetic"
	MethodBiometric MatchMethod = "biometric"
	MethodLSH       MatchMethod = "lsh"
)

type CompositeEngine struct {
	db  *gorm.DB
	lsh *LSHIndex
}

func NewCompositeEngine(db *gorm.DB, lsh *LSHIndex) *CompositeEngine {
	return &CompositeEngine{db: db, lsh: lsh}
}

func (e *CompositeEngine) Match(req models.MatchRequest) ([]models.MatchCandidate, error) {
	var allCandidates []models.MatchCandidate

	candidateIDs := e.getCandidates(req)
	if len(candidateIDs) == 0 {
		return nil, nil
	}

	var candidates []models.Identity
	e.db.Where("id IN ?", candidateIDs).Find(&candidates)

	for _, c := range candidates {
		methods := make([]string, 0)
		var totalScore float64
		var weightSum float64

		exactScore := e.matchExact(req, c)
		if exactScore > 0 {
			totalScore += exactScore * 1.0
			weightSum += 1.0
			methods = append(methods, "exact")
		}

		fuzzyScore := e.matchFuzzy(req, c)
		if fuzzyScore > 0.5 {
			totalScore += fuzzyScore * 0.6
			weightSum += 0.6
			methods = append(methods, "fuzzy")
		}

		phoneticScore := e.matchPhonetic(req, c)
		if phoneticScore > 0.5 {
			totalScore += phoneticScore * 0.4
			weightSum += 0.4
			methods = append(methods, "phonetic")
		}

		biometricScore := e.matchBiometric(req, c)
		if biometricScore > 0 {
			totalScore += biometricScore * 0.8
			weightSum += 0.8
			methods = append(methods, "biometric")
		}

		if weightSum == 0 {
			continue
		}

		finalScore := totalScore / weightSum
		if finalScore < 0.3 {
			continue
		}

		allCandidates = append(allCandidates, models.MatchCandidate{
			IdentityID: c.ID,
			Score:      math.Round(finalScore*100) / 100,
			Methods:    methods,
		})
	}

	return allCandidates, nil
}

func (e *CompositeEngine) Reconcile(id1, id2 string) (*models.ReconciliationResult, error) {
	var primary, secondary models.Identity
	if err := e.db.First(&primary, "id = ?", id1).Error; err != nil {
		return nil, fmt.Errorf("primary identity not found: %w", err)
	}
	if err := e.db.First(&secondary, "id = ?", id2).Error; err != nil {
		return nil, fmt.Errorf("secondary identity not found: %w", err)
	}

	scores := make(map[string]float64)
	evidence := make([]string, 0)

	req := models.MatchRequest{
		FirstName:     primary.FirstName,
		LastName:      primary.LastName,
		FullName:      primary.FullName,
		DOB:           primary.DOB,
		TaxID:         primary.TaxID,
		NNU:           primary.NNU,
		BiometricHash: primary.BiometricHash,
	}

	sim := models.Identity{
		ID:            secondary.ID,
		FirstName:     secondary.FirstName,
		LastName:      secondary.LastName,
		FullName:      secondary.FullName,
		DOB:           secondary.DOB,
		TaxID:         secondary.TaxID,
		NNU:           secondary.NNU,
		BiometricHash: secondary.BiometricHash,
	}

	exactScore := e.matchExact(req, sim)
	scores["exact"] = exactScore
	if exactScore > 0 {
		evidence = append(evidence, fmt.Sprintf("Exact match on NNU/TaxID (score: %.2f)", exactScore))
	}

	fuzzyScore := e.matchFuzzy(req, sim)
	scores["fuzzy"] = fuzzyScore
	if fuzzyScore > 0.7 {
		evidence = append(evidence, fmt.Sprintf("Fuzzy name match (Jaro-Winkler: %.2f)", fuzzyScore))
	}

	phoneticScore := e.matchPhonetic(req, sim)
	scores["phonetic"] = phoneticScore
	if phoneticScore > 0.7 {
		evidence = append(evidence, fmt.Sprintf("Phonetic name match (Metaphone/Soundex: %.2f)", phoneticScore))
	}

	biometricScore := e.matchBiometric(req, sim)
	scores["biometric"] = biometricScore
	if biometricScore > 0.7 {
		evidence = append(evidence, fmt.Sprintf("Biometric hash similarity (%.2f)", biometricScore))
	}

	overall := (exactScore*1.0 + fuzzyScore*0.6 + phoneticScore*0.4 + biometricScore*0.8) / 2.8
	decision := "rejected"
	if overall >= 0.85 {
		decision = "confirmed_match"
	} else if overall >= 0.65 {
		decision = "pending_review"
	}

	return &models.ReconciliationResult{
		PrimaryID:       id1,
		SecondaryID:     id2,
		OverallScore:    math.Round(overall*100) / 100,
		AttributeScores: scores,
		Decision:        decision,
		Evidence:        evidence,
	}, nil
}

func (e *CompositeEngine) Merge(id1, id2, resolvedBy string) error {
	var primary, secondary models.Identity
	if err := e.db.First(&primary, "id = ?", id1).Error; err != nil {
		return err
	}
	if err := e.db.First(&secondary, "id = ?", id2).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	resolved := models.ResolvedIdentity{
		ID:          uuid.New().String(),
		PrimaryID:   id1,
		SecondaryID: id2,
		MatchScore:  1.0,
		MatchMethod: "manual_merge",
		Status:      "merged",
		ResolvedBy:  resolvedBy,
		ResolvedAt:  &now,
	}
	if err := e.db.Create(&resolved).Error; err != nil {
		return err
	}

	return e.db.Model(&models.Identity{}).Where("id = ?", id2).Update("status", "merged").Error
}

func (e *CompositeEngine) Split(identityID, resolvedBy string) error {
	e.db.Model(&models.ResolvedIdentity{}).
		Where("secondary_id = ? AND status = ?", identityID, "merged").
		Update("status", "split")

	return e.db.Model(&models.Identity{}).Where("id = ?", identityID).Update("status", "active").Error
}

func (e *CompositeEngine) GetStats() (*models.StatsResponse, error) {
	stats := &models.StatsResponse{}
	e.db.Model(&models.Identity{}).Count(&stats.TotalIdentities)
	e.db.Model(&models.ResolvedIdentity{}).Count(&stats.TotalResolved)
	e.db.Model(&models.ResolvedIdentity{}).Where("status = ?", "pending").Count(&stats.PendingReview)
	e.db.Model(&models.ResolvedIdentity{}).Where("status = ?", "confirmed").Count(&stats.ConfirmedMatches)
	e.db.Model(&models.ResolvedIdentity{}).Where("status = ?", "rejected").Count(&stats.RejectedMatches)
	e.db.Model(&models.ResolvedIdentity{}).Where("status = ?", "merged").Count(&stats.MergedCount)
	return stats, nil
}

func (e *CompositeEngine) getCandidates(req models.MatchRequest) []string {
	features := extractFeatures(req)
	candidates := e.lsh.Query(features)

	if len(candidates) == 0 {
		if req.TaxID != "" {
			var byTaxID []models.Identity
			e.db.Where("tax_id = ?", req.TaxID).Select("id").Find(&byTaxID)
			for _, id := range byTaxID {
				candidates = append(candidates, id.ID)
			}
		}
		if req.NNU != "" {
			var byNNU []models.Identity
			e.db.Where("nnu = ?", req.NNU).Select("id").Find(&byNNU)
			for _, id := range byNNU {
				candidates = append(candidates, id.ID)
			}
		}
	}

	return candidates
}

func (e *CompositeEngine) matchExact(req models.MatchRequest, candidate models.Identity) float64 {
	if req.NNU != "" && strings.EqualFold(req.NNU, candidate.NNU) {
		return 1.0
	}
	if req.TaxID != "" && strings.EqualFold(req.TaxID, candidate.TaxID) {
		return 1.0
	}
	if req.NationalID != "" && strings.EqualFold(req.NationalID, candidate.NationalID) {
		return 1.0
	}
	return 0
}

func (e *CompositeEngine) matchFuzzy(req models.MatchRequest, candidate models.Identity) float64 {
	fnSim := JaroWinkler(req.FirstName, candidate.FirstName)
	lnSim := JaroWinkler(req.LastName, candidate.LastName)
	fnLev := NormalizedLevenshtein(req.FirstName, candidate.FirstName)
	lnLev := NormalizedLevenshtein(req.LastName, candidate.LastName)

	jaroScore := 0.4*fnSim + 0.6*lnSim
	levScore := 0.4*fnLev + 0.6*lnLev

	return 0.6*jaroScore + 0.4*levScore
}

func (e *CompositeEngine) matchPhonetic(req models.MatchRequest, candidate models.Identity) float64 {
	fnSoundex1 := Soundex(req.FirstName)
	fnSoundex2 := Soundex(candidate.FirstName)
	lnSoundex1 := Soundex(req.LastName)
	lnSoundex2 := Soundex(candidate.LastName)

	fnMeta1 := Metaphone(req.FirstName)
	fnMeta2 := Metaphone(candidate.FirstName)
	lnMeta1 := Metaphone(req.LastName)
	lnMeta2 := Metaphone(candidate.LastName)

	soundexScore := 0.0
	if fnSoundex1 != "" && fnSoundex2 != "" && fnSoundex1 == fnSoundex2 {
		soundexScore += 0.4
	}
	if lnSoundex1 != "" && lnSoundex2 != "" && lnSoundex1 == lnSoundex2 {
		soundexScore += 0.6
	}

	metaScore := 0.0
	if fnMeta1 != "" && fnMeta2 != "" && fnMeta1 == fnMeta2 {
		metaScore += 0.4
	}
	if lnMeta1 != "" && lnMeta2 != "" && lnMeta1 == lnMeta2 {
		metaScore += 0.6
	}

	return 0.5*soundexScore + 0.5*metaScore
}

func (e *CompositeEngine) matchBiometric(req models.MatchRequest, candidate models.Identity) float64 {
	if req.BiometricHash == "" || candidate.BiometricHash == "" {
		return 0
	}
	if strings.EqualFold(req.BiometricHash, candidate.BiometricHash) {
		return 1.0
	}
	return 0
}

func extractFeatures(req models.MatchRequest) []float64 {
	features := make([]float64, 256)
	fullName := strings.ToUpper(strings.TrimSpace(req.FullName))
	if fullName == "" {
		fullName = strings.ToUpper(strings.TrimSpace(req.FirstName + " " + req.LastName))
	}

	for i := 0; i < len(fullName) && i < 256; i++ {
		features[i%256] += float64(fullName[i]) / 255.0
	}
	if req.DOB != "" {
		for i := 0; i < len(req.DOB) && i < 256; i++ {
			features[(i+64)%256] += float64(req.DOB[i]) / 255.0
		}
	}
	if req.TaxID != "" {
		for i := 0; i < len(req.TaxID) && i < 256; i++ {
			features[(i+128)%256] += float64(req.TaxID[i]) / 255.0
		}
	}
	if req.NNU != "" {
		for i := 0; i < len(req.NNU) && i < 256; i++ {
			features[(i+192)%256] += float64(req.NNU[i]) / 255.0
		}
	}

	for i := range features {
		if features[i] > 1.0 {
			features[i] = 1.0
		}
	}
	return features
}
