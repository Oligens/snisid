package interpol

import (
	"context"
	"strings"
	"testing"
	"time"
)

type mockInterpolClient struct {
	connectErr    error
	toI247XMLData []byte
	toI247XMLErr  error
	submitErr     error
	closeErr      error

	connectCalled int
	submitCalled  int
	closeCalled   int
}

func (m *mockInterpolClient) Connect(_ context.Context) error {
	m.connectCalled++
	return m.connectErr
}

func (m *mockInterpolClient) ToI247XML(_ *Submission) ([]byte, error) {
	return m.toI247XMLData, m.toI247XMLErr
}

func (m *mockInterpolClient) Submit(_ context.Context, _ []byte) error {
	m.submitCalled++
	return m.submitErr
}

func (m *mockInterpolClient) Close() error {
	m.closeCalled++
	return m.closeErr
}

var _ InterpolClient = (*mockInterpolClient)(nil)

func TestMockInterpolClient_ImplementsInterface(t *testing.T) {
	m := &mockInterpolClient{}
	var iface InterpolClient = m
	_ = iface
}

func TestFormatDNAProfile(t *testing.T) {
	if FormatDNAProfile != "DNA_PROFILE" {
		t.Errorf("FormatDNAProfile = %q, want %q", FormatDNAProfile, "DNA_PROFILE")
	}
}

func TestSubmissionCreation(t *testing.T) {
	now := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
	s := &Submission{
		ID:        "SUB-001",
		SampleIDs: []string{"SAMP-A01", "SAMP-B02"},
		Reason:    "Crime scene match",
		CaseRef:   "HTI-2026-0042",
		CreatedAt: now,
	}
	if s.ID != "SUB-001" {
		t.Errorf("s.ID = %q, want %q", s.ID, "SUB-001")
	}
	if len(s.SampleIDs) != 2 {
		t.Errorf("len(s.SampleIDs) = %d, want 2", len(s.SampleIDs))
	}
	if s.SampleIDs[1] != "SAMP-B02" {
		t.Errorf("s.SampleIDs[1] = %q, want %q", s.SampleIDs[1], "SAMP-B02")
	}
	if s.Reason != "Crime scene match" {
		t.Errorf("s.Reason = %q, want %q", s.Reason, "Crime scene match")
	}
	if s.CaseRef != "HTI-2026-0042" {
		t.Errorf("s.CaseRef = %q, want %q", s.CaseRef, "HTI-2026-0042")
	}
	if !s.CreatedAt.Equal(now) {
		t.Errorf("s.CreatedAt mismatch")
	}
}

func TestToI247XML_ValidFormat(t *testing.T) {
	g := &Gateway{}
	s := &Submission{
		ID:        "MSG-XYZ",
		SampleIDs: []string{"S1", "S2"},
		Reason:    "Familial match",
		CaseRef:   "CASE-99",
		CreatedAt: time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC),
	}
	data, err := g.ToI247XML(s)
	if err != nil {
		t.Fatalf("ToI247XML returned error: %v", err)
	}
	xml := string(data)

	if !strings.HasPrefix(xml, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("XML missing declaration")
	}
	if !strings.Contains(xml, "<I247Message>") {
		t.Error("XML missing <I247Message>")
	}
	if !strings.Contains(xml, "<MessageID>MSG-XYZ</MessageID>") {
		t.Error("XML missing MessageID")
	}
	if !strings.Contains(xml, "<Format>DNA_PROFILE</Format>") {
		t.Error("XML missing or wrong Format")
	}
	if !strings.Contains(xml, "<SampleID>S1</SampleID>") {
		t.Error("XML missing SampleID S1")
	}
	if !strings.Contains(xml, "<SampleID>S2</SampleID>") {
		t.Error("XML missing SampleID S2")
	}
	if !strings.Contains(xml, "<Reason>Familial match</Reason>") {
		t.Error("XML missing Reason")
	}
	if !strings.Contains(xml, "<CaseRef>CASE-99</CaseRef>") {
		t.Error("XML missing CaseRef")
	}
	if !strings.Contains(xml, "<Origin>HTI-BCN-DCPJ</Origin>") {
		t.Error("XML missing Origin")
	}
	if !strings.Contains(xml, "<Timestamp>2026-01-15T14:30:00Z</Timestamp>") {
		t.Error("XML missing or wrong Timestamp")
	}
	if !strings.Contains(xml, "</I247Message>") {
		t.Error("XML missing closing </I247Message>")
	}
}

func TestConnect_EmptyCertPaths_ReturnsError(t *testing.T) {
	g := NewGateway("", "", "", "")
	err := g.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for empty cert paths, got nil")
	}
	if !strings.Contains(err.Error(), "SNISID-PKI") {
		t.Errorf("error %q should mention SNISID-PKI", err.Error())
	}
}

func TestMockInterpolClient_Connect(t *testing.T) {
	m := &mockInterpolClient{connectErr: nil}
	err := m.Connect(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if m.connectCalled != 1 {
		t.Errorf("connectCalled = %d, want 1", m.connectCalled)
	}
}

func TestMockInterpolClient_ConnectError(t *testing.T) {
	m := &mockInterpolClient{connectErr: context.DeadlineExceeded}
	err := m.Connect(context.Background())
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestMockInterpolClient_Submit(t *testing.T) {
	m := &mockInterpolClient{}
	err := m.Submit(context.Background(), []byte("data"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if m.submitCalled != 1 {
		t.Errorf("submitCalled = %d, want 1", m.submitCalled)
	}
}

func TestMockInterpolClient_Close(t *testing.T) {
	m := &mockInterpolClient{}
	err := m.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if m.closeCalled != 1 {
		t.Errorf("closeCalled = %d, want 1", m.closeCalled)
	}
}
