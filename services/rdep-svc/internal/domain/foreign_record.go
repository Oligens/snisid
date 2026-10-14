package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ForeignRecord struct {
	ForeignRecordID    uuid.UUID          `json:"foreign_record_id"`
	DeporteeID         uuid.UUID          `json:"deportee_id"`
	Country            DeportationCountry `json:"country"`
	CourtName          string             `json:"court_name"`
	OffenseDescription string             `json:"offense_description"`
	OffenseDate        *time.Time         `json:"offense_date,omitempty"`
	ConvictionDate     *time.Time         `json:"conviction_date,omitempty"`
	Sentence           string             `json:"sentence"`
	PrisonServed       string             `json:"prison_served"`
	FBINumber          string             `json:"fbi_number"`
	InterpolRef        string             `json:"interpol_ref"`
	SourceDocument     string             `json:"source_document"`
	CreatedAt          time.Time          `json:"created_at"`
}

func (r *ForeignRecord) HasViolentOffenses() bool {
	violentKeywords := []string{"murder", "homicide", "assault", "robbery", "kidnapping", "rape", "sexual"}
	for _, keyword := range violentKeywords {
		for _, offense := range []string{r.OffenseDescription, r.Sentence} {
			if containsIgnoreCase(offense, keyword) {
				return true
			}
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && len(substr) > 0 &&
			(s[0:len(substr)] == substr || containsIgnoreCase(s[1:], substr)))
}

type ForeignRecordRepository interface {
	Create(ctx context.Context, record *ForeignRecord) error
	FindByID(ctx context.Context, id uuid.UUID) (*ForeignRecord, error)
	FindByDeporteeID(ctx context.Context, deporteeID uuid.UUID) ([]*ForeignRecord, error)
}
