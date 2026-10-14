package domain

import (
	"time"

	"github.com/google/uuid"
)

type Person struct {
	SnisidID     uuid.UUID  `json:"snisid_id" db:"snisid_id"`
	Name         string     `json:"name" db:"name"`
	Aliases      []string   `json:"aliases" db:"aliases"`
	Nationality  *string    `json:"nationality,omitempty" db:"nationality"`
	DOB          *time.Time `json:"dob,omitempty" db:"dob"`
	RiskScore    float64    `json:"risk_score" db:"risk_score"`
	IsGangMember bool       `json:"is_gang_member" db:"is_gang_member"`
	IsSanctioned bool       `json:"is_sanctioned" db:"is_sanctioned"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type Gang struct {
	GangID          uuid.UUID     `json:"gang_id" db:"gang_id"`
	Name            string        `json:"name" db:"name"`
	PrimaryActivity *string       `json:"primary_activity,omitempty" db:"primary_activity"`
	TerritoryDept   *string       `json:"territory_dept,omitempty" db:"territory_dept"`
	ActivityLevel   ActivityLevel `json:"activity_level" db:"activity_level"`
	MemberCount     int           `json:"member_count" db:"member_count"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
}

type PersonGang struct {
	PersonID   uuid.UUID  `json:"person_id" db:"person_id"`
	GangID     uuid.UUID  `json:"gang_id" db:"gang_id"`
	Role       GangRole   `json:"role" db:"role"`
	Since      *time.Time `json:"since,omitempty" db:"since"`
	Confidence float64    `json:"confidence" db:"confidence"`
}

type GangRelation struct {
	GangID1      uuid.UUID    `json:"gang_id_1" db:"gang_id_1"`
	GangID2      uuid.UUID    `json:"gang_id_2" db:"gang_id_2"`
	RelationType RelationType `json:"relation_type" db:"relation_type"`
	Since        *time.Time   `json:"since,omitempty" db:"since"`
	Confidence   float64      `json:"confidence" db:"confidence"`
}

type PersonAssociation struct {
	PersonID1       uuid.UUID       `json:"person_id_1" db:"person_id_1"`
	PersonID2       uuid.UUID       `json:"person_id_2" db:"person_id_2"`
	AssociationType AssociationType `json:"association_type" db:"association_type"`
	Confidence      float64         `json:"confidence" db:"confidence"`
	Source          *string         `json:"source,omitempty" db:"source"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type CriminalNetwork struct {
	NetworkID       uuid.UUID   `json:"network_id" db:"network_id"`
	NetworkName     *string     `json:"network_name,omitempty" db:"network_name"`
	MemberIDs       []uuid.UUID `json:"member_ids" db:"member_ids"`
	CommunitySize   int         `json:"community_size" db:"community_size"`
	ModularityScore *float64    `json:"modularity_score,omitempty" db:"modularity_score"`
	DetectedAt      time.Time   `json:"detected_at" db:"detected_at"`
	AnalysisVersion string      `json:"analysis_version" db:"analysis_version"`
}

type NetworkAnalysisResult struct {
	Networks     []CriminalNetwork `json:"networks"`
	TotalPersons int               `json:"total_persons"`
	TotalGangs   int               `json:"total_gangs"`
	TotalEdges   int               `json:"total_edges"`
	Communities  int               `json:"communities_detected"`
	AnalyzedAt   time.Time         `json:"analyzed_at"`
}

type KeyActor struct {
	PersonID         uuid.UUID  `json:"person_id"`
	Name             string     `json:"name"`
	PageRankScore    float64    `json:"pagerank_score"`
	BetweennessScore float64    `json:"betweenness_score"`
	DegreeScore      float64    `json:"degree_score"`
	CompositeScore   float64    `json:"composite_score"`
	GangID           *uuid.UUID `json:"gang_id,omitempty"`
	GangName         *string    `json:"gang_name,omitempty"`
}

type ShortestPathResult struct {
	FromID     uuid.UUID   `json:"from_id"`
	ToID       uuid.UUID   `json:"to_id"`
	Path       []uuid.UUID `json:"path"`
	PathLength int         `json:"path_length"`
	PathNames  []string    `json:"path_names"`
	Exists     bool        `json:"exists"`
}

type GangOverlapResult struct {
	GangID1       uuid.UUID     `json:"gang_id_1"`
	GangID2       uuid.UUID     `json:"gang_id_2"`
	Gang1Name     string        `json:"gang_1_name"`
	Gang2Name     string        `json:"gang_2_name"`
	CommonMembers []uuid.UUID   `json:"common_members"`
	OverlapCount  int           `json:"overlap_count"`
	RelationType  *RelationType `json:"relation_type,omitempty"`
}

type CommunityDetectionResult struct {
	Communities   []Community `json:"communities"`
	Modularity    float64     `json:"modularity"`
	TotalClusters int         `json:"total_clusters"`
}

type Community struct {
	ID      int         `json:"id"`
	Members []uuid.UUID `json:"members"`
	Size    int         `json:"size"`
	GangIDs []uuid.UUID `json:"gang_ids"`
}

type NetworkRepository interface {
	GetAllPersons() ([]Person, error)
	GetAllGangs() ([]Gang, error)
	GetPersonGangMemberships() ([]PersonGang, error)
	GetGangRelations() ([]GangRelation, error)
	GetPersonAssociations() ([]PersonAssociation, error)
	GetNetworks() ([]CriminalNetwork, error)
	SaveNetwork(network *CriminalNetwork) error
	GetPersonNetwork(personID uuid.UUID) ([]PersonAssociation, error)
	GetGangOverlap(gangID1, gangID2 uuid.UUID) ([]uuid.UUID, error)
}
