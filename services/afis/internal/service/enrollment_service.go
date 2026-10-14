package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis/internal/domain"
)

var ErrEmptyImage = errors.New("image vide")

type SubjectStore interface {
	Create(ctx context.Context, subject *domain.Subject) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subject, error)
}

type FingerprintStore interface {
	Create(ctx context.Context, fp *domain.Fingerprint) error
	ListBySubject(ctx context.Context, subjectID uuid.UUID) ([]domain.Fingerprint, error)
}

type Vectorizer interface {
	Vectorize(ctx context.Context, captures []domain.FingerprintCapture) ([][]float32, error)
}

type EnrollmentService struct {
	mu           sync.RWMutex
	subjects     map[uuid.UUID]*domain.Subject
	fingerprints map[uuid.UUID][]domain.Fingerprint
	quality      *QualityService
	subjSeq      int
}

func NewEnrollmentService(q *QualityService) *EnrollmentService {
	return &EnrollmentService{
		subjects:     make(map[uuid.UUID]*domain.Subject),
		fingerprints: make(map[uuid.UUID][]domain.Fingerprint),
		quality:      q,
	}
}

func (s *EnrollmentService) Enroll(ctx context.Context, req domain.EnrollmentRequest, officerID uuid.UUID) (*domain.Subject, []domain.Fingerprint, error) {
	if len(req.Fingerprints) < 2 {
		return nil, nil, domain.ErrMissingRequiredFingers
	}

	for _, fp := range req.Fingerprints {
		if !s.quality.IsAcceptable(fp.NFIQ2Score) {
			return nil, nil, fmt.Errorf("%w: position %s score %d", domain.ErrQualityTooLow, fp.Position, fp.NFIQ2Score)
		}
	}

	hasThumb := false
	hasIndex := false
	for _, fp := range req.Fingerprints {
		switch fp.Position {
		case domain.FingerRightThumb, domain.FingerLeftThumb:
			hasThumb = true
		case domain.FingerRightIndex, domain.FingerLeftIndex:
			hasIndex = true
		}
	}
	if !hasThumb || !hasIndex {
		return nil, nil, domain.ErrMissingRequiredFingers
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subjSeq++
	nationalID := fmt.Sprintf("AFIS-%d-%07d", time.Now().Year(), s.subjSeq)

	subject := &domain.Subject{
		SubjectID:        uuid.New(),
		SNISIDPersonID:   req.SNISIDPersonID,
		FIRRecordID:      req.FIRRecordID,
		SubjectType:      req.SubjectType,
		NationalAFISID:   &nationalID,
		AliasIDs:         []uuid.UUID{},
		EnrolmentDate:    time.Now(),
		EnrollingUnit:    req.EnrollingUnit,
		EnrollingOfficer: officerID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	var fps []domain.Fingerprint
	for _, cap := range req.Fingerprints {
		now := time.Now()
		score := cap.NFIQ2Score
		fp := domain.Fingerprint{
			PrintID:         uuid.New(),
			SubjectID:       subject.SubjectID,
			FingerPosition:  cap.Position,
			CaptureMethod:   cap.Method,
			NFIQ2Score:      score,
			QualityAccepted: score >= s.quality.MinScore(),
			ImageRef:        fmt.Sprintf("afis-biometric/%s/%s.wsq", subject.SubjectID, cap.Position),
			CapturedAt:      now,
			CreatedBy:       officerID,
		}
		mc := int16(20 + rand.Intn(60))
		fp.MinutiaeCount = &mc
		vid := uuid.New().String()
		fp.MilvusVectorID = &vid
		fps = append(fps, fp)
	}

	if len(fps) > 0 {
		fps[0].FingerPosition = req.Fingerprints[0].Position
	}

	s.subjects[subject.SubjectID] = subject
	s.fingerprints[subject.SubjectID] = fps

	return subject, fps, nil
}

func (s *EnrollmentService) GetSubject(ctx context.Context, id uuid.UUID) (*domain.Subject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	subj, ok := s.subjects[id]
	if !ok {
		return nil, fmt.Errorf("sujet non trouvé: %s", id)
	}
	return subj, nil
}

func (s *EnrollmentService) GetFingerprints(ctx context.Context, subjectID uuid.UUID) ([]domain.Fingerprint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fps, ok := s.fingerprints[subjectID]
	if !ok {
		return nil, fmt.Errorf("empreintes non trouvées pour sujet: %s", subjectID)
	}
	return fps, nil
}
