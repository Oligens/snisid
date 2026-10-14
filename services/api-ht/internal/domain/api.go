package domain

import (
	"time"

	"github.com/google/uuid"
)

type Sensitivity string

const (
	SensitivityPublic       Sensitivity = "PUBLIC"
	SensitivityRestricted   Sensitivity = "RESTRICTED"
	SensitivityConfidential Sensitivity = "CONFIDENTIAL"
	SensitivitySecret       Sensitivity = "SECRET"
)

type APIEndpoint struct {
	ID           uuid.UUID   `json:"id"`
	Path         string      `json:"path"`
	Method       string      `json:"method"`
	Description  *string     `json:"description,omitempty"`
	Sensitivity  Sensitivity `json:"sensitivity"`
	ModuleSource *string     `json:"module_source,omitempty"`
	BasePath     string      `json:"base_path"`
	IsActive     bool        `json:"is_active"`
	Version      string      `json:"version"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type DeveloperAccount struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	OrgName      *string   `json:"org_name,omitempty"`
	ContactName  string    `json:"contact_name"`
	ContactPhone *string   `json:"contact_phone,omitempty"`
	IsApproved   bool      `json:"is_approved"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type APIKey struct {
	ID          uuid.UUID  `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	KeyValue    string     `json:"key_value"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

type UsageLog struct {
	ID        uuid.UUID `json:"id"`
	KeyID     uuid.UUID `json:"key_id"`
	Endpoint  string    `json:"endpoint"`
	Method    string    `json:"method"`
	Status    int       `json:"status"`
	LatencyMs int       `json:"latency_ms"`
	IPAddress *string   `json:"ip_address,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
