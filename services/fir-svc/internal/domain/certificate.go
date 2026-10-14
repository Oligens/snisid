package domain

import (
	"time"

	"github.com/google/uuid"
)

type Certificate struct {
	CertID            uuid.UUID         `json:"cert_id"`
	RecordID          *uuid.UUID        `json:"record_id,omitempty"`
	SNISIDPersonID    uuid.UUID         `json:"snisid_person_id"`
	CertificateNumber string            `json:"certificate_number"`
	IssuedFor         string            `json:"issued_for"`
	Result            CertificateResult `json:"result"`
	IssuedBy          uuid.UUID         `json:"issued_by"`
	IssuingOffice     string            `json:"issuing_office"`
	IssuedAt          time.Time         `json:"issued_at"`
	ExpiresAt         *time.Time        `json:"expires_at,omitempty"`
	QRCodeRef         string            `json:"qr_code_ref,omitempty"`
}

type CertificateRequest struct {
	PersonID uuid.UUID `json:"person_id" binding:"required"`
	Purpose  string    `json:"purpose" binding:"required"`
	Office   string    `json:"office" binding:"required"`
	IssuedBy uuid.UUID `json:"issued_by" binding:"required"`
}
