package models

import "time"

type Identity struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	NNU           string    `gorm:"uniqueIndex;size:50" json:"nnu"`
	NationalID    string    `gorm:"index;size:50" json:"national_id"`
	FirstName     string    `gorm:"size:100" json:"first_name"`
	LastName      string    `gorm:"size:100" json:"last_name"`
	FullName      string    `gorm:"size:200" json:"full_name"`
	DOB           string    `gorm:"size:20" json:"dob"`
	BirthPlace    string    `gorm:"size:200" json:"birth_place"`
	MotherName    string    `gorm:"size:200" json:"mother_name"`
	TaxID         string    `gorm:"index;size:50" json:"tax_id"`
	BiometricHash string    `gorm:"size:255" json:"biometric_hash"`
	Status        string    `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (i Identity) GetID() string        { return i.ID }
func (i Identity) GetNNU() string       { return i.NNU }
func (i Identity) GetFullName() string  { return i.FullName }
func (i Identity) GetFirstName() string { return i.FirstName }
func (i Identity) GetLastName() string  { return i.LastName }
func (i Identity) GetDOB() string       { return i.DOB }
func (i Identity) GetTaxID() string     { return i.TaxID }

func (Identity) TableName() string {
	return "resolution_identities"
}

type ResolvedIdentity struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	PrimaryID   string     `gorm:"index;not null" json:"primary_id"`
	SecondaryID string     `gorm:"index;not null" json:"secondary_id"`
	MatchScore  float64    `gorm:"type:decimal(5,2)" json:"match_score"`
	MatchMethod string     `gorm:"size:50" json:"match_method"`
	Status      string     `gorm:"size:20;default:'pending'" json:"status"`
	ResolvedBy  string     `gorm:"size:100" json:"resolved_by"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (ResolvedIdentity) TableName() string {
	return "resolution_resolved_identities"
}

type MatchCandidate struct {
	IdentityID string   `json:"identity_id"`
	Score      float64  `json:"score"`
	Methods    []string `json:"methods"`
}

type MatchRequest struct {
	NNU           string `json:"nnu"`
	NationalID    string `json:"national_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	FullName      string `json:"full_name"`
	DOB           string `json:"dob"`
	BirthPlace    string `json:"birth_place"`
	MotherName    string `json:"mother_name"`
	TaxID         string `json:"tax_id"`
	BiometricHash string `json:"biometric_hash"`
}

type ReconciliationResult struct {
	PrimaryID       string             `json:"primary_id"`
	SecondaryID     string             `json:"secondary_id"`
	OverallScore    float64            `json:"overall_score"`
	AttributeScores map[string]float64 `json:"attribute_scores"`
	Decision        string             `json:"decision"`
	Evidence        []string           `json:"evidence"`
}

type MergeRequest struct {
	PrimaryID   string `json:"primary_id" binding:"required"`
	SecondaryID string `json:"secondary_id" binding:"required"`
	ResolvedBy  string `json:"resolved_by"`
}

type SplitRequest struct {
	IdentityID string `json:"identity_id" binding:"required"`
	ResolvedBy string `json:"resolved_by"`
}

type StatsResponse struct {
	TotalIdentities  int64 `json:"total_identities"`
	TotalResolved    int64 `json:"total_resolved"`
	PendingReview    int64 `json:"pending_review"`
	ConfirmedMatches int64 `json:"confirmed_matches"`
	RejectedMatches  int64 `json:"rejected_matches"`
	MergedCount      int64 `json:"merged_count"`
}
