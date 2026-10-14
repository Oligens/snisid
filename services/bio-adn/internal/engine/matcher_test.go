package engine

import (
	"testing"
)

func TestHashProfile(t *testing.T) {
	m := NewMatcher()
	loci := STRLoci{
		"CSF1PO":  {Value1: "10", Value2: "12"},
		"D3S1358": {Value1: "15", Value2: "18"},
	}
	h := m.HashProfile(loci)
	if h == "" {
		t.Fatal("expected non-empty hash")
	}
	if len(h) != 64 {
		t.Fatalf("expected SHA-256 length 64, got %d", len(h))
	}
}

func TestCompareFullMatch(t *testing.T) {
	m := NewMatcher()
	query := STRLoci{
		"CSF1PO":  {Value1: "10", Value2: "12"},
		"D3S1358": {Value1: "15", Value2: "18"},
		"D5S818":  {Value1: "11", Value2: "14"},
	}
	candidate := STRLoci{
		"CSF1PO":  {Value1: "10", Value2: "12"},
		"D3S1358": {Value1: "15", Value2: "18"},
		"D5S818":  {Value1: "11", Value2: "14"},
	}
	r := m.Compare(query, candidate)
	if r == nil {
		t.Fatal("expected result")
	}
	if r.Score != 1.0 {
		t.Fatalf("expected score 1.0, got %f", r.Score)
	}
	if r.MatchType != "FULL_MATCH" {
		t.Fatalf("expected FULL_MATCH, got %s", r.MatchType)
	}
}

func TestComparePartialMatch(t *testing.T) {
	m := NewMatcher()
	query := STRLoci{
		"CSF1PO":  {Value1: "10", Value2: "12"},
		"D3S1358": {Value1: "15", Value2: "18"},
		"D5S818":  {Value1: "11", Value2: "14"},
		"D8S1179": {Value1: "13", Value2: "16"},
	}
	candidate := STRLoci{
		"CSF1PO":  {Value1: "10", Value2: "12"},
		"D3S1358": {Value1: "15", Value2: "18"},
		"D5S818":  {Value1: "11", Value2: "13"},
		"D8S1179": {Value1: "13", Value2: "17"},
	}
	r := m.Compare(query, candidate)
	if r.Score < 0.7 || r.Score >= 0.95 {
		t.Fatalf("expected partial match score ~0.75, got %f", r.Score)
	}
}

func TestCompareNoMatch(t *testing.T) {
	m := NewMatcher()
	query := STRLoci{
		"CSF1PO":  {Value1: "10", Value2: "12"},
		"D3S1358": {Value1: "15", Value2: "18"},
	}
	candidate := STRLoci{
		"CSF1PO":  {Value1: "11", Value2: "13"},
		"D3S1358": {Value1: "16", Value2: "19"},
	}
	r := m.Compare(query, candidate)
	if r.MatchType != "FAMILIAL" {
		t.Fatalf("expected FAMILIAL for low score, got %s", r.MatchType)
	}
}

func TestClassifyAlert(t *testing.T) {
	m := NewMatcher()
	tests := []struct {
		score float64
		want  string
	}{
		{0.98, "CRITICAL"},
		{0.90, "HIGH"},
		{0.80, "MEDIUM"},
		{0.50, "LOW"},
	}
	for _, tc := range tests {
		got := m.classifyAlert(tc.score)
		if got != tc.want {
			t.Errorf("classifyAlert(%f) = %s, want %s", tc.score, got, tc.want)
		}
	}
}
