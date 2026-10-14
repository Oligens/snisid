package domain

import (
	"time"

	"github.com/google/uuid"
)

type Classification string

const (
	ClassificationUnclassified Classification = "UNCLASSIFIED"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationSecret       Classification = "SECRET"
	ClassificationTopSecret    Classification = "TOP_SECRET"
)

type ConfidenceLevel string

const (
	ConfidenceLow      ConfidenceLevel = "LOW"
	ConfidenceModerate ConfidenceLevel = "MODERATE"
	ConfidenceHigh     ConfidenceLevel = "HIGH"
	ConfidenceCertain  ConfidenceLevel = "CERTAIN"
)

type ActorType string

const (
	ActorTypeState     ActorType = "STATE"
	ActorTypeNonState  ActorType = "NON_STATE"
	ActorTypeCriminal  ActorType = "CRIMINAL"
	ActorTypeTerrorist ActorType = "TERRORIST"
	ActorTypeInsurgent ActorType = "INSURGENT"
)

type CorrelationType string

const (
	CorrelationSupports    CorrelationType = "SUPPORTS"
	CorrelationContradicts CorrelationType = "CONTRADICTS"
	CorrelationIndependent CorrelationType = "INDEPENDENT"
)

type IntelProduct struct {
	ProductID           uuid.UUID       `json:"product_id"`
	Title               string          `json:"title"`
	Classification      Classification  `json:"classification"`
	SourceDisciplines   []string        `json:"source_disciplines"`
	SigintRefs          []string        `json:"sigint_refs"`
	HumintRefs          []string        `json:"humint_refs"`
	GeointRefs          []string        `json:"geoint_refs"`
	OsintRefs           []string        `json:"osint_refs"`
	AnalystAssessment   string          `json:"analyst_assessment"`
	ConfidenceLevel     ConfidenceLevel `json:"confidence_level"`
	RelatedThreatActors []string        `json:"related_threat_actors"`
	RelatedRegions      []string        `json:"related_regions"`
	NIERef              *string         `json:"nie_ref,omitempty"`
	CreatedBy           uuid.UUID       `json:"created_by"`
	ApprovedBy          *uuid.UUID      `json:"approved_by,omitempty"`
	ValidFrom           time.Time       `json:"valid_from"`
	ValidUntil          *time.Time      `json:"valid_until,omitempty"`
}

type ThreatActor struct {
	ActorID          uuid.UUID  `json:"actor_id"`
	Name             string     `json:"name"`
	Aliases          []string   `json:"aliases"`
	Type             ActorType  `json:"type"`
	CapLevel         int        `json:"cap_level"`
	IntentLevel      int        `json:"intent_level"`
	OpportunityLevel int        `json:"opportunity_level"`
	OverallRisk      int        `json:"overall_risk"`
	LastActivityAt   *time.Time `json:"last_activity_at,omitempty"`
	PrimaryRegion    string     `json:"primary_region"`
	AssociatedGroups []string   `json:"associated_groups"`
	OFACDesignated   bool       `json:"ofac_designated"`
	Notes            string     `json:"notes"`
}

type CrossDisciplineCorrelation struct {
	CorrelationID   uuid.UUID       `json:"correlation_id"`
	DisciplineA     string          `json:"discipline_a"`
	ReferenceA      string          `json:"reference_a"`
	DisciplineB     string          `json:"discipline_b"`
	ReferenceB      string          `json:"reference_b"`
	CorrelationType CorrelationType `json:"correlation_type"`
	AnalystNotes    string          `json:"analyst_notes"`
	Score           float64         `json:"score"`
	CreatedAt       time.Time       `json:"created_at"`
}
