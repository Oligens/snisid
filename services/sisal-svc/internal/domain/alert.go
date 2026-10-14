package domain

import (
	"time"

	"github.com/google/uuid"
)

type HazardType string

const (
	Earthquake           HazardType = "EARTHQUAKE"
	Hurricane            HazardType = "HURRICANE"
	Flood                HazardType = "FLOOD"
	Tsunami              HazardType = "TSUNAMI"
	Landslide            HazardType = "LANDSLIDE"
	SecurityGang         HazardType = "SECURITY_GANG"
	SecurityMassCasualty HazardType = "SECURITY_MASS_CASUALTY"
	Epidemic             HazardType = "EPIDEMIC"
	Industrial           HazardType = "INDUSTRIAL"
	Composite            HazardType = "COMPOSITE"
)

type Severity string

const (
	Advisory    Severity = "ADVISORY"
	Watch       Severity = "WATCH"
	Warning     Severity = "WARNING"
	Emergency   Severity = "EMERGENCY"
	Catastrophe Severity = "CATASTROPHE"
)

type SISALAlert struct {
	ID              uuid.UUID  `json:"alert_id" db:"alert_id"`
	NationalSisalID string     `json:"national_sisal_id" db:"national_sisal_id"`
	HazardType      HazardType `json:"hazard_type" db:"hazard_type"`
	Severity        Severity   `json:"severity" db:"severity"`
	Title           string     `json:"title" db:"title"`
	MessageFR       string     `json:"message_fr" db:"message_fr"`
	MessageHT       string     `json:"message_ht" db:"message_ht"`
	AffectedDepts   []string   `json:"affected_depts" db:"affected_depts"`
	AffectedPopEst  *int       `json:"affected_pop_est,omitempty" db:"affected_pop_est"`
	IssuedAt        time.Time  `json:"issued_at" db:"issued_at"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	SourceAgency    string     `json:"source_agency" db:"source_agency"`
	SourceEventID   *uuid.UUID `json:"source_event_id,omitempty" db:"source_event_id"`
	IsCancelled     *bool      `json:"is_cancelled,omitempty" db:"is_cancelled"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelReason    *string    `json:"cancel_reason,omitempty" db:"cancel_reason"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type Subscription struct {
	ID             uuid.UUID  `json:"sub_id" db:"sub_id"`
	SNISIDPersonID *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	PhoneNumber    *string    `json:"phone_number,omitempty" db:"phone_number"`
	Email          *string    `json:"email,omitempty" db:"email"`
	DeptCode       *string    `json:"dept_code,omitempty" db:"dept_code"`
	Commune        *string    `json:"commune,omitempty" db:"commune"`
	MinSeverity    Severity   `json:"min_severity" db:"min_severity"`
	IsActive       *bool      `json:"is_active,omitempty" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type IssueAlertRequest struct {
	HazardType     string   `json:"hazard_type" binding:"required"`
	Severity       string   `json:"severity" binding:"required"`
	Title          string   `json:"title" binding:"required"`
	MessageFR      string   `json:"message_fr" binding:"required"`
	MessageHT      string   `json:"message_ht" binding:"required"`
	AffectedDepts  []string `json:"affected_depts"`
	SourceAgency   string   `json:"source_agency" binding:"required"`
	AffectedPopEst *int     `json:"affected_pop_est"`
}

type SubscribeRequest struct {
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	DeptCode    string `json:"dept_code"`
	Commune     string `json:"commune"`
	MinSeverity string `json:"min_severity"`
}

type AlertRepository interface {
	Create(alert *SISALAlert) (*SISALAlert, error)
	FindActive() ([]SISALAlert, error)
	FindHistory() ([]SISALAlert, error)
	Cancel(id uuid.UUID, reason string) error
	CreateSubscription(sub *Subscription) (*Subscription, error)
}
