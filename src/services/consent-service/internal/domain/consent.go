package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAgency    = errors.New("invalid target agency")
	ErrExpirationTooFar = errors.New("consent expiration exceeds maximum allowed duration")
)

type ConsentAgreement struct {
	ID               uuid.UUID
	CitizenNIU       string
	TargetAgency     string
	ExpirationDate   time.Time
	DigitalSignature string
	CreatedAt        time.Time
}

func NewConsentAgreement(niu, agency string, daysValid int, signature string) (*ConsentAgreement, error) {
	if agency == "" {
		return nil, ErrInvalidAgency
	}
	if daysValid > 365 {
		return nil, ErrExpirationTooFar
	}

	return &ConsentAgreement{
		ID:               uuid.New(),
		CitizenNIU:       niu,
		TargetAgency:     agency,
		ExpirationDate:   time.Now().UTC().AddDate(0, 0, daysValid),
		DigitalSignature: signature,
		CreatedAt:        time.Now().UTC(),
	}, nil
}
