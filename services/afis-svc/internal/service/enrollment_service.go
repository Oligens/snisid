package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

type Vectorizer interface {
	Vectorize(ctx context.Context, captures []domain.FingerprintCapture) ([][]float32, error)
}

type EnrollmentService struct {
	fingerprintRepo FingerprintRepo
	subjectRepo     SubjectRepo
	imageRepo       ImageRepo
	vectorRepo      VectorRepo
	qualitySvc      *QualityService
	vectorizer      Vectorizer
}

func NewEnrollmentService(
	fpr FingerprintRepo,
	sr SubjectRepo,
	ir ImageRepo,
	vr VectorRepo,
	qs *QualityService,
	vz Vectorizer,
) *EnrollmentService {
	return &EnrollmentService{
		fingerprintRepo: fpr,
		subjectRepo:     sr,
		imageRepo:       ir,
		vectorRepo:      vr,
		qualitySvc:      qs,
		vectorizer:      vz,
	}
}

func (s *EnrollmentService) Enroll(ctx context.Context, req domain.EnrollmentRequest, officerID uuid.UUID) (*domain.SubjectProfile, error) {
	hasThumb := false
	hasIndex := false
	for _, fp := range req.Fingerprints {
		if fp.NFIQ2Score < 60 {
			return nil, domain.ErrQualityTooLow
		}
		if fp.Position == domain.FingerRightThumb || fp.Position == domain.FingerLeftThumb {
			hasThumb = true
		}
		if fp.Position == domain.FingerRightIndex || fp.Position == domain.FingerLeftIndex {
			hasIndex = true
		}
	}
	if !hasThumb || !hasIndex {
		return nil, domain.ErrMissingRequiredFingers
	}

	subjectID := uuid.New()
	nationalID := fmt.Sprintf("AFIS-%d-%07d", time.Now().Year(), subjectID.ID()%10000000)

	profile := &domain.SubjectProfile{
		SubjectID:      subjectID,
		SNISIDPersonID: req.SNISIDPersonID,
		FIRRecordID:    req.FIRRecordID,
		SubjectType:    req.SubjectType,
		NationalAFISID: &nationalID,
		EnrollingUnit:  req.EnrollingUnit,
	}

	if err := s.subjectRepo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}

	var fingerprints []domain.Fingerprint
	for _, cap := range req.Fingerprints {
		imgData, err := base64.StdEncoding.DecodeString(cap.ImageBase64)
		if err != nil {
			return nil, fmt.Errorf("decode image: %w", err)
		}

		printID := uuid.New()
		objectName := fmt.Sprintf("%s/%s.wsq", subjectID.String(), printID.String())
		if err := s.imageRepo.Upload(ctx, objectName, imgData, "image/x-wsq"); err != nil {
			return nil, fmt.Errorf("upload image: %w", err)
		}

		fp := &domain.Fingerprint{
			PrintID:         printID,
			SubjectID:       subjectID,
			FingerPosition:  cap.Position,
			CaptureMethod:   cap.Method,
			NFIQ2Score:      cap.NFIQ2Score,
			QualityAccepted: cap.NFIQ2Score >= 60,
			ImageRef:        objectName,
			CapturedAt:      time.Now(),
			CreatedBy:       officerID,
		}
		if err := s.fingerprintRepo.Create(ctx, fp); err != nil {
			return nil, fmt.Errorf("save fingerprint: %w", err)
		}
		fingerprints = append(fingerprints, *fp)

		if err := s.qualitySvc.ScheduleQualityCheck(ctx, fp); err != nil {
			return nil, fmt.Errorf("quality check: %w", err)
		}
	}

	vectors, err := s.vectorizer.Vectorize(ctx, req.Fingerprints)
	if err != nil {
		return nil, fmt.Errorf("vectorize: %w", err)
	}

	var printIDs []string
	for _, fp := range fingerprints {
		printIDs = append(printIDs, fp.PrintID.String())
	}
	if err := s.vectorRepo.InsertVectors(ctx, printIDs, vectors); err != nil {
		return nil, fmt.Errorf("insert vectors: %w", err)
	}

	profile.Fingerprints = fingerprints
	return profile, nil
}
