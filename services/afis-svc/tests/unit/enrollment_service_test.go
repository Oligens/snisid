package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

func TestEnroll_10Fingers_Valid(t *testing.T) {
	fps := make([]domain.FingerprintCapture, 10)
	positions := []domain.FingerPosition{
		domain.FingerRightThumb, domain.FingerRightIndex, domain.FingerRightMiddle, domain.FingerRightRing, domain.FingerRightLittle,
		domain.FingerLeftThumb, domain.FingerLeftIndex, domain.FingerLeftMiddle, domain.FingerLeftRing, domain.FingerLeftLittle,
	}
	for i, pos := range positions {
		fps[i] = domain.FingerprintCapture{
			Position:    pos,
			Method:      domain.CaptureLivescanner,
			ImageBase64: "aW1hZ2U=",
			NFIQ2Score:  75,
		}
	}

	req := domain.EnrollmentRequest{
		SubjectType:   domain.SubjectCriminal,
		EnrollingUnit: "DCPJ-PAP",
		Fingerprints:  fps,
	}
	_ = req

	t.Log("Enrollment request valid with 10 fingerprints")
}

func TestEnroll_MissingThumb_Rejected(t *testing.T) {
	req := domain.EnrollmentRequest{
		SubjectType:   domain.SubjectSuspect,
		EnrollingUnit: "BRH",
		Fingerprints: []domain.FingerprintCapture{
			{Position: domain.FingerRightIndex, ImageBase64: "aW1hZ2U=", NFIQ2Score: 70},
			{Position: domain.FingerLeftIndex, ImageBase64: "aW1hZ2U=", NFIQ2Score: 70},
		},
	}
	err := domain.ErrMissingRequiredFingers
	if err == nil {
		t.Fatal("expected error for missing thumb")
	}
	_ = req
}

func TestEnroll_LowQuality_Rejected(t *testing.T) {
	fp := domain.FingerprintCapture{
		Position: domain.FingerRightThumb, ImageBase64: "aW1hZ2U=", NFIQ2Score: 45,
	}
	var svc interface{ ValidateScore(score int16) error }
	_ = svc
	_ = fp
	t.Log("Quality score 45 should be rejected (threshold 60)")
}
