package identitymesh

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIdentityMesh(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)
	assert.NotNil(t, m)
	assert.Equal(t, "SNISID-HTI", m.PlatformID)
	assert.Nil(t, m.neo4jDriver)
}

func TestFuseRecords_AllMatch_Success(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)

	oni := Record{Agency: "oni", ID: "NNU-001", Data: map[string]interface{}{
		"firstName":   "Jean",
		"lastName":    "Dupont",
		"dob":         "1990-01-15",
		"gender":      "M",
		"nationality": "HTI",
	}}
	dgi := Record{Agency: "dgi", ID: "DGI-123", Data: map[string]interface{}{
		"firstName": "Jean",
		"lastName":  "Dupont",
	}}
	anh := Record{Agency: "anh", ID: "ANH-456", Data: map[string]interface{}{
		"firstName":   "Jean",
		"lastName":    "Dupont",
		"dob":         "1990-01-15",
		"gender":      "M",
		"nationality": "HTI",
	}}
	dcpj := Record{Agency: "dcpj", ID: "DCPJ-789", Data: map[string]interface{}{
		"gender": "M",
	}}

	fused := m.FuseRecords(oni, dgi, anh, dcpj)
	assert.Equal(t, "NNU-001", fused.NNU)
	assert.Equal(t, "VERIFIED", fused.Status)
	assert.GreaterOrEqual(t, fused.Confidence.Overall, 0.85)
	assert.NotEmpty(t, fused.MatchedFields)
	assert.Empty(t, fused.ConflictFields)
}

func TestFuseRecords_Conflict_Detected(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)

	oni := Record{Agency: "oni", ID: "NNU-002", Data: map[string]interface{}{
		"firstName": "Jean",
		"lastName":  "Dupont",
		"gender":    "M",
	}}
	dgi := Record{Agency: "dgi", ID: "DGI-456", Data: map[string]interface{}{
		"firstName": "Jean",
		"lastName":  "Dupont",
	}}
	anh := Record{Agency: "anh", ID: "ANH-789", Data: map[string]interface{}{
		"firstName": "Jeanne", // Different!
		"lastName":  "Dupont",
		"gender":    "F", // Different!
	}}
	dcpj := Record{Agency: "dcpj", ID: "DCPJ-012", Data: map[string]interface{}{
		"gender": "F",
	}}

	fused := m.FuseRecords(oni, dgi, anh, dcpj)
	assert.Equal(t, "CONFLICT", fused.Status)
	assert.NotEmpty(t, fused.ConflictFields)
	assert.Contains(t, fused.ConflictFields, "firstName")
	assert.Contains(t, fused.ConflictFields, "gender")
}

func TestDetectInconsistency_ReturnsMessages(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)
	fused := FusedIdentity{
		NNU: "NNU-003",
		ConflictFields: map[string][]interface{}{
			"firstName": {
				map[string]interface{}{"agency": "oni", "value": "Jean"},
				map[string]interface{}{"agency": "anh", "value": "Jeanne"},
			},
		},
	}

	inconsistencies := m.DetectInconsistency(fused)
	assert.Len(t, inconsistencies, 2)
	assert.Contains(t, inconsistencies[0], "firstName")
}

func TestResolveConflict_RemovesConflict(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)
	fused := FusedIdentity{
		NNU: "NNU-004",
		ConflictFields: map[string][]interface{}{
			"firstName": {
				map[string]interface{}{"agency": "anh", "value": "Jeanne", "resolved": false},
			},
		},
		Confidence: ConfidenceScore{Overall: 0.85},
	}

	resolved := m.ResolveConflict(fused, "anh", "firstName", "Jean")
	assert.Empty(t, resolved.ConflictFields)
	assert.Equal(t, "VERIFIED", resolved.Status)
	assert.Greater(t, resolved.Confidence.Overall, 0.85)
}

func TestGetIdentityGraph_NoNeo4j_ReturnsError(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)
	_, err := m.GetIdentityGraph(context.Background(), "NNU-999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "neo4j not configured")
}

func TestFuseRecords_ConfidenceLow_ManualReview(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)

	oni := Record{Agency: "oni", ID: "NNU-005", Data: map[string]interface{}{
		"firstName": "A",
		"lastName":  "B",
		"dob":       "2000-01-01",
		"gender":    "M",
		"nationality": "HTI",
	}}
	dgi := Record{Agency: "dgi", ID: "DGI-001", Data: map[string]interface{}{
		"firstName": "X", // Conflict
		"lastName":  "B",
	}}
	anh := Record{Agency: "anh", ID: "ANH-001", Data: map[string]interface{}{
		"firstName":   "A",
		"lastName":    "C", // Conflict
		"dob":         "2000-01-01",
		"gender":      "M",
		"nationality": "HTI",
	}}
	dcpj := Record{Agency: "dcpj", ID: "DCPJ-001", Data: map[string]interface{}{
		"gender": "M",
	}}

	fused := m.FuseRecords(oni, dgi, anh, dcpj)
	assert.Equal(t, "MANUAL_REVIEW", fused.Status)
}

func TestPersistToGraph_NoNeo4j(t *testing.T) {
	m := NewIdentityMesh("SNISID-HTI", nil)
	fused := FusedIdentity{
		NNU:    "NNU-TEST",
		Status: "VERIFIED",
		AgencyRecords: map[string]string{"oni": "NNU-TEST"},
		Confidence: ConfidenceScore{Overall: 0.95},
		LastUpdated: time.Now().UTC(),
	}

	err := m.persistToGraph(context.Background(), fused)
	assert.NoError(t, err) // Should gracefully skip when no driver
}
