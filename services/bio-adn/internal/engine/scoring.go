package engine

import (
	"math"
	"sort"
)

type LRMatchResult struct {
	SampleID        string  `json:"sample_id"`
	LikelihoodRatio float64 `json:"likelihood_ratio"`
	MatchType       string  `json:"match_type"`
	MatchedLoci     int     `json:"matched_loci"`
	TotalLoci       int     `json:"total_loci"`
	AlertLevel      string  `json:"alert_level"`
}

type AlleleFreq map[string]map[string]float64

var DefaultAlleleFrequencies = AlleleFreq{
	"CSF1PO":   {"10": 0.220, "11": 0.310, "12": 0.350, "13": 0.080, "14": 0.035, "15": 0.005},
	"D3S1358":  {"14": 0.110, "15": 0.250, "16": 0.280, "17": 0.210, "18": 0.120, "19": 0.030},
	"D5S818":   {"10": 0.080, "11": 0.350, "12": 0.320, "13": 0.180, "14": 0.050, "15": 0.020},
	"D7S820":   {"8": 0.150, "9": 0.120, "10": 0.250, "11": 0.220, "12": 0.180, "13": 0.060, "14": 0.020},
	"D8S1179":  {"10": 0.080, "11": 0.120, "12": 0.150, "13": 0.300, "14": 0.180, "15": 0.120, "16": 0.050},
	"D13S317":  {"8": 0.110, "9": 0.080, "10": 0.080, "11": 0.320, "12": 0.280, "13": 0.100, "14": 0.030},
	"D16S539":  {"9": 0.100, "10": 0.080, "11": 0.310, "12": 0.280, "13": 0.180, "14": 0.050},
	"D18S51":   {"12": 0.050, "13": 0.120, "14": 0.150, "15": 0.160, "16": 0.120, "17": 0.100, "18": 0.080, "19": 0.050, "20": 0.030, "21": 0.020},
	"D21S11":   {"28": 0.020, "29": 0.200, "30": 0.280, "31": 0.080, "31.2": 0.060, "32": 0.050, "32.2": 0.100, "33.2": 0.070, "34.2": 0.020},
	"FGA":      {"19": 0.060, "20": 0.080, "21": 0.150, "22": 0.180, "23": 0.140, "24": 0.120, "25": 0.080, "26": 0.050, "27": 0.020},
	"TH01":     {"6": 0.210, "7": 0.150, "8": 0.080, "9": 0.180, "9.3": 0.320, "10": 0.050},
	"TPOX":     {"8": 0.520, "9": 0.100, "10": 0.080, "11": 0.280, "12": 0.020},
	"vWA":      {"14": 0.080, "15": 0.080, "16": 0.200, "17": 0.280, "18": 0.210, "19": 0.100, "20": 0.050},
	"D1S1656":  {"11": 0.050, "12": 0.120, "13": 0.200, "14": 0.150, "15": 0.250, "16": 0.120, "17": 0.080, "18": 0.030},
	"D2S441":   {"10": 0.080, "11": 0.250, "12": 0.080, "13": 0.050, "14": 0.320, "15": 0.180, "16": 0.040},
	"D2S1338":  {"17": 0.080, "18": 0.120, "19": 0.180, "20": 0.220, "21": 0.080, "22": 0.060, "23": 0.050, "24": 0.040, "25": 0.020},
	"D10S1248": {"13": 0.120, "14": 0.280, "15": 0.350, "16": 0.200, "17": 0.050},
	"D12S391":  {"17": 0.020, "18": 0.080, "19": 0.120, "20": 0.150, "21": 0.180, "22": 0.120, "23": 0.050, "24": 0.020},
	"D19S433":  {"12": 0.080, "13": 0.250, "14": 0.320, "15": 0.200, "16": 0.080, "17": 0.020},
	"D22S1045": {"11": 0.150, "12": 0.080, "13": 0.050, "14": 0.020, "15": 0.350, "16": 0.280, "17": 0.070},
}

func (m *Matcher) CalculateLR(query, candidate STRLoci, freqs AlleleFreq) *LRMatchResult {
	if freqs == nil {
		freqs = DefaultAlleleFrequencies
	}

	matchedLoci := 0
	totalLoci := 0
	var lr float64 = 1.0

	for locus, qv := range query {
		cv, ok := candidate[locus]
		if !ok {
			continue
		}
		totalLoci++

		locusFreqs, hasFreq := freqs[locus]
		if !hasFreq {
			continue
		}

		matchCount := 0
		if qv.Value1 == cv.Value1 || qv.Value1 == cv.Value2 {
			matchCount++
		}
		if qv.Value2 == cv.Value1 || qv.Value2 == cv.Value2 {
			matchCount++
		}

		if matchCount == 2 {
			matchedLoci += 2
			f1 := alleleFreq(locusFreqs, qv.Value1)
			f2 := alleleFreq(locusFreqs, qv.Value2)
			lr *= 1.0 / (2 * f1 * f2)
		} else if matchCount == 1 {
			matchedLoci++
			f1 := alleleFreq(locusFreqs, qv.Value1)
			f2 := alleleFreq(locusFreqs, qv.Value2)
			lr *= 1.0 / (f1 + f2)
		}
	}

	if totalLoci == 0 {
		return nil
	}

	lr = math.Max(lr, 1.0)
	logLR := math.Log10(lr)

	matchType := "NO_MATCH"
	alertLevel := "LOW"
	switch {
	case logLR >= 10:
		matchType = "FULL_MATCH"
		alertLevel = "CRITICAL"
	case logLR >= 6:
		matchType = "FULL_MATCH"
		alertLevel = "HIGH"
	case logLR >= 3:
		matchType = "PARTIAL"
		alertLevel = "MEDIUM"
	case logLR >= 1:
		matchType = "FAMILIAL"
		alertLevel = "LOW"
	}

	return &LRMatchResult{
		LikelihoodRatio: math.Round(lr*100) / 100,
		MatchType:       matchType,
		MatchedLoci:     matchedLoci,
		TotalLoci:       totalLoci * 2,
		AlertLevel:      alertLevel,
	}
}

func alleleFreq(freqs map[string]float64, allele string) float64 {
	f, ok := freqs[allele]
	if !ok {
		return 0.001
	}
	return f
}

func (m *Matcher) CalculateLRFromProfiles(query, candidate STRLoci, freqs AlleleFreq) float64 {
	result := m.CalculateLR(query, candidate, freqs)
	if result == nil {
		return 1.0
	}
	return result.LikelihoodRatio
}

func (m *Matcher) RankByLR(query STRLoci, candidates []STRLoci, candidateIDs []string, freqs AlleleFreq) []LRMatchResult {
	var results []LRMatchResult
	for i, candidate := range candidates {
		lr := m.CalculateLR(query, candidate, freqs)
		if lr != nil {
			lr.SampleID = candidateIDs[i]
			results = append(results, *lr)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].LikelihoodRatio > results[j].LikelihoodRatio
	})
	if len(results) > 100 {
		results = results[:100]
	}
	return results
}
