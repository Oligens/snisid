package models

import (
	"time"

	"gorm.io/gorm"
)

type Identity struct {
	ID            string         `gorm:"primaryKey" json:"id"`
	NNU           string         `gorm:"uniqueIndex;size:12" json:"nnu"`
	FirstName     string         `gorm:"size:100;not null" json:"first_name"`
	LastName      string         `gorm:"size:100;not null" json:"last_name"`
	DateOfBirth   string         `gorm:"size:10" json:"date_of_birth"`
	Gender        string         `gorm:"size:1" json:"gender"`
	Nationality   string         `gorm:"size:3;default:'HTI'" json:"nationality"`
	Status        string         `gorm:"size:20;default:'pending'" json:"status"`
	BiometricHash string         `gorm:"size:64" json:"biometric_hash"`
	Email         string         `gorm:"size:255" json:"email"`
	Phone         string         `gorm:"size:20" json:"phone"`
	Address       string         `gorm:"type:text" json:"address"`
	Version       int            `gorm:"default:1" json:"version"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type IdentityHistory struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	IdentityID string    `gorm:"index;not null" json:"identity_id"`
	FieldName  string    `gorm:"size:50" json:"field_name"`
	OldValue   string    `gorm:"type:text" json:"old_value"`
	NewValue   string    `gorm:"type:text" json:"new_value"`
	ChangedBy  string    `gorm:"size:100" json:"changed_by"`
	ChangedAt  time.Time `json:"changed_at"`
}

type BiometricReference struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	IdentityID   string         `gorm:"index;not null" json:"identity_id"`
	Type         string         `gorm:"size:50" json:"type"`
	ReferenceURI string         `gorm:"type:text" json:"reference_uri"`
	QualityScore float64        `json:"quality_score"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type DocumentAssociation struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	IdentityID   string         `gorm:"index;not null" json:"identity_id"`
	DocumentType string         `gorm:"size:50" json:"document_type"`
	DocumentURI  string         `gorm:"type:text" json:"document_uri"`
	Verified     bool           `json:"verified"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
