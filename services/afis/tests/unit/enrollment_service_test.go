package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis/internal/domain"
	"github.com/snisid/platform/services/afis/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnroll_10Fingers_Valid(t *testing.T) {
	quality := service.NewQualityService(60)
	enrollment := service.NewEnrollmentService(quality)

	req := domain.EnrollmentRequest{
		SubjectType:   domain.SubjectCriminal,
		EnrollingUnit: "DCPJ-PAP",
		Fingerprints: []domain.FingerprintCapture{
			{Position: domain.FingerRightThumb, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,AAA=", NFIQ2Score: 85},
			{Position: domain.FingerRightIndex, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,BBB=", NFIQ2Score: 82},
			{Position: domain.FingerRightMiddle, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,CCC=", NFIQ2Score: 78},
			{Position: domain.FingerRightRing, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,DDD=", NFIQ2Score: 90},
			{Position: domain.FingerRightLittle, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,EEE=", NFIQ2Score: 75},
			{Position: domain.FingerLeftThumb, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,FFF=", NFIQ2Score: 88},
			{Position: domain.FingerLeftIndex, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,GGG=", NFIQ2Score: 80},
			{Position: domain.FingerLeftMiddle, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,HHH=", NFIQ2Score: 76},
			{Position: domain.FingerLeftRing, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,III=", NFIQ2Score: 81},
			{Position: domain.FingerLeftLittle, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,JJJ=", NFIQ2Score: 79},
		},
	}

	officerID := uuid.New()
	subject, fps, err := enrollment.Enroll(context.Background(), req, officerID)

	require.NoError(t, err)
	require.NotNil(t, subject)
	assert.NotNil(t, subject.SubjectID)
	assert.Equal(t, domain.SubjectCriminal, subject.SubjectType)
	assert.Equal(t, "DCPJ-PAP", subject.EnrollingUnit)
	assert.Len(t, fps, 10)
	assert.Contains(t, *subject.NationalAFISID, "AFIS-")
}

func TestEnroll_Invalid_MissingRequiredFingers(t *testing.T) {
	quality := service.NewQualityService(60)
	enrollment := service.NewEnrollmentService(quality)

	req := domain.EnrollmentRequest{
		SubjectType:   domain.SubjectSuspect,
		EnrollingUnit: "BRH",
		Fingerprints: []domain.FingerprintCapture{
			{Position: domain.FingerRightMiddle, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,AAA=", NFIQ2Score: 85},
		},
	}

	_, _, err := enrollment.Enroll(context.Background(), req, uuid.New())
	assert.ErrorIs(t, err, domain.ErrMissingRequiredFingers)
}

func TestEnroll_Invalid_QualityTooLow(t *testing.T) {
	quality := service.NewQualityService(60)
	enrollment := service.NewEnrollmentService(quality)

	req := domain.EnrollmentRequest{
		SubjectType:   domain.SubjectVictim,
		EnrollingUnit: "LABO-FORENSIC",
		Fingerprints: []domain.FingerprintCapture{
			{Position: domain.FingerRightThumb, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,AAA=", NFIQ2Score: 25},
			{Position: domain.FingerLeftIndex, Method: domain.CaptureLivescanner, ImageBase64: "data:image/wsq;base64,BBB=", NFIQ2Score: 30},
		},
	}

	_, _, err := enrollment.Enroll(context.Background(), req, uuid.New())
	assert.ErrorIs(t, err, domain.ErrQualityTooLow)
}

func TestEnroll_Invalid_EmptyCaptures(t *testing.T) {
	quality := service.NewQualityService(60)
	enrollment := service.NewEnrollmentService(quality)

	req := domain.EnrollmentRequest{
		SubjectType:   domain.SubjectCriminal,
		EnrollingUnit: "DCPJ",
		Fingerprints:  []domain.FingerprintCapture{},
	}

	_, _, err := enrollment.Enroll(context.Background(), req, uuid.New())
	assert.ErrorIs(t, err, domain.ErrMissingRequiredFingers)
}
