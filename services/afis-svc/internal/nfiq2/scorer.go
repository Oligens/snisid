package nfiq2

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"sync"
)

type NFIQ2Scorer struct {
	mu        sync.Mutex
	modelPath string
}

func NewScorer(modelPath string) *NFIQ2Scorer {
	return &NFIQ2Scorer{
		modelPath: modelPath,
	}
}

func (s *NFIQ2Scorer) ScoreImage(imgData []byte) (int16, error) {
	_, _, err := image.Decode(bytesToReader(imgData))
	if err != nil {
		return 0, fmt.Errorf("decode fingerprint image: %w", err)
	}

	score := computeSimulatedScore(imgData)
	return int16(math.Round(score)), nil
}

func computeSimulatedScore(data []byte) float64 {
	quality := 80.0

	if len(data) < 10240 {
		quality -= 15.0
	}

	entropy := estimateEntropy(data)
	if entropy < 6.0 {
		quality -= 10.0
	} else if entropy > 7.5 {
		quality += 5.0
	}

	if quality < 0 {
		return 0
	}
	if quality > 100 {
		return 100
	}
	return quality
}

func estimateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	var entropy float64
	length := float64(len(data))
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

type reader struct {
	data []byte
	pos  int
}

func bytesToReader(data []byte) *reader {
	return &reader{data: data, pos: 0}
}

func (r *reader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
