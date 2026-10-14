package identitymesh

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type Record struct {
	Agency string                 `json:"agency"`
	ID     string                 `json:"id"`
	Data   map[string]interface{} `json:"data"`
}

type ConfidenceScore struct {
	Overall  float64            `json:"overall"`
	ByAgency map[string]float64 `json:"byAgency"`
	Factors  []string           `json:"factors"`
}

type FusedIdentity struct {
	NNU            string                   `json:"nnu"`
	AgencyRecords  map[string]string        `json:"agencyRecords"`
	Confidence     ConfidenceScore          `json:"confidence"`
	Status         string                   `json:"status"`
	MatchedFields  map[string][]string      `json:"matchedFields"`
	ConflictFields map[string][]interface{} `json:"conflictFields"`
	GraphNeo4jID   string                   `json:"graphNeo4jId,omitempty"`
	LastUpdated    time.Time                `json:"lastUpdated"`
}

type IdentityMesh struct {
	PlatformID  string
	neo4jDriver neo4j.DriverWithContext
}

func NewIdentityMesh(platformID string, driver neo4j.DriverWithContext) *IdentityMesh {
	return &IdentityMesh{
		PlatformID:  platformID,
		neo4jDriver: driver,
	}
}

func (m *IdentityMesh) FuseRecords(oni, dgi, anh, dcpj Record) FusedIdentity {
	fields := map[string][]string{
		"firstName":   {"oni", "dgi", "anh"},
		"lastName":    {"oni", "dgi", "anh"},
		"dob":         {"oni", "anh"},
		"gender":      {"oni", "anh", "dcpj"},
		"nationality": {"oni", "anh"},
	}

	matchedFields := make(map[string][]string)
	conflictFields := make(map[string][]interface{})
	agencyConfidence := map[string]float64{
		"oni":  0.95,
		"dgi":  0.85,
		"anh":  0.90,
		"dcpj": 0.80,
	}

	for field, agencies := range fields {
		values := make(map[string]string)
		var firstVal string
		firstSet := false
		allMatch := true

		for _, agency := range agencies {
			var rec Record
			switch agency {
			case "oni":
				rec = oni
			case "dgi":
				rec = dgi
			case "anh":
				rec = anh
			case "dcpj":
				rec = dcpj
			}
			if v, ok := rec.Data[field]; ok {
				valStr := fmt.Sprintf("%v", v)
				values[agency] = valStr
				if !firstSet {
					firstVal = valStr
					firstSet = true
				} else if valStr != firstVal {
					allMatch = false
				}
			}
		}

		if allMatch && firstSet {
			matchedFields[field] = agencies
		} else if !allMatch {
			conflicts := make([]interface{}, 0)
			for agency, val := range values {
				conflicts = append(conflicts, map[string]interface{}{
					"agency": agency,
					"value":  val,
				})
			}
			conflictFields[field] = conflicts

			for _, agency := range agencies {
				if _, ok := values[agency]; ok {
					agencyConfidence[agency] *= 0.85
				}
			}
		}
	}

	overallConf := 1.0
	for _, conf := range agencyConfidence {
		overallConf *= conf
	}
	overallConf = math.Pow(overallConf, 1.0/float64(len(agencyConfidence)))

	status := "VERIFIED"
	factors := []string{"all_agencies_responded"}
	if len(conflictFields) > 0 {
		status = "CONFLICT"
		factors = append(factors, fmt.Sprintf("conflicts_in_%d_fields", len(conflictFields)))
	}
	if overallConf < 0.85 {
		status = "MANUAL_REVIEW"
		factors = append(factors, "low_confidence")
	}

	fused := FusedIdentity{
		NNU: oni.ID,
		AgencyRecords: map[string]string{
			"oni":  oni.ID,
			"dgi":  dgi.ID,
			"anh":  anh.ID,
			"dcpj": dcpj.ID,
		},
		Confidence: ConfidenceScore{
			Overall:  math.Round(overallConf*1000) / 1000,
			ByAgency: agencyConfidence,
			Factors:  factors,
		},
		Status:         status,
		MatchedFields:  matchedFields,
		ConflictFields: conflictFields,
		LastUpdated:    time.Now().UTC(),
	}

	if m.neo4jDriver != nil {
		if err := m.persistToGraph(context.Background(), fused); err != nil {
			logger.Warn(context.Background(), "failed to persist identity graph node", zap.Error(err))
		}
	}

	return fused
}

func (m *IdentityMesh) DetectInconsistency(fused FusedIdentity) []string {
	inconsistencies := make([]string, 0)
	for field, conflicts := range fused.ConflictFields {
		for _, c := range conflicts {
			entry, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			inconsistencies = append(inconsistencies,
				fmt.Sprintf("field '%s' mismatch for agency '%s': value '%v'", field, entry["agency"], entry["value"]))
		}
	}
	return inconsistencies
}

func (m *IdentityMesh) ResolveConflict(fused FusedIdentity, agency string, field string, resolvedValue interface{}) FusedIdentity {
	if conflicts, ok := fused.ConflictFields[field]; ok {
		newConflicts := make([]interface{}, 0)
		for _, c := range conflicts {
			entry, ok := c.(map[string]interface{})
			if !ok {
				newConflicts = append(newConflicts, c)
				continue
			}
			if entry["agency"] == agency {
				entry["value"] = resolvedValue
				entry["resolved"] = true
			}
			newConflicts = append(newConflicts, entry)
		}
		if len(newConflicts) == 0 {
			delete(fused.ConflictFields, field)
		} else {
			fused.ConflictFields[field] = newConflicts
		}

		allResolved := true
		for _, conflicts := range fused.ConflictFields {
			for _, c := range conflicts {
				entry, ok := c.(map[string]interface{})
				if ok && entry["resolved"] != true {
					allResolved = false
				}
			}
		}
		if allResolved && len(fused.ConflictFields) > 0 {
			fused.ConflictFields = make(map[string][]interface{})
			fused.Status = "VERIFIED"
		}
	}
	fused.Confidence.Overall = math.Min(1.0, fused.Confidence.Overall+0.05)
	return fused
}

func (m *IdentityMesh) persistToGraph(ctx context.Context, fused FusedIdentity) error {
	session := m.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	data, _ := json.Marshal(fused)
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx,
			`MERGE (i:Identity {nnu: $nnu})
			 SET i.fusedData = $data, i.status = $status, i.updatedAt = datetime()
			 WITH i
			 UNWIND keys($agencyRecords) AS agency
			 MERGE (a:Agency {name: agency})
			 MERGE (i)-[:REGISTERED_AT]->(a)
			 RETURN i.nnu`,
			map[string]interface{}{
				"nnu":           fused.NNU,
				"data":          string(data),
				"status":        fused.Status,
				"agencyRecords": fused.AgencyRecords,
			})
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})
	return err
}

func (m *IdentityMesh) GetIdentityGraph(ctx context.Context, nnu string) (*FusedIdentity, error) {
	if m.neo4jDriver == nil {
		return nil, fmt.Errorf("neo4j not configured")
	}

	session := m.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		record, err := tx.Run(ctx,
			`MATCH (i:Identity {nnu: $nnu})
			 OPTIONAL MATCH (i)-[r:REGISTERED_AT]->(a:Agency)
			 RETURN i.nnu AS nnu, i.fusedData AS fusedData, i.status AS status,
			        collect(a.name) AS agencies`,
			map[string]interface{}{"nnu": nnu})
		if err != nil {
			return nil, err
		}
		return record.Collect(ctx)
	})
	if err != nil {
		return nil, err
	}

	records := result.([]*neo4j.Record)
	if len(records) == 0 {
		return nil, fmt.Errorf("identity %s not found in graph", nnu)
	}

	fusedDataJSON, _ := records[0].Get("fusedData")
	var fused FusedIdentity
	if err := json.Unmarshal([]byte(fusedDataJSON.(string)), &fused); err != nil {
		return nil, err
	}
	fused.GraphNeo4jID = nnu
	return &fused, nil
}
