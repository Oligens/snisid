package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir-svc/internal/domain"
)

type RecordService struct {
	repo       domain.CriminalRecordRepository
	arrestRepo domain.ArrestRepository
	convRepo   domain.ConvictionRepository
	eventPub   domain.EventPublisher
	snisid     domain.SNISIDClient
}

func NewRecordService(
	repo domain.CriminalRecordRepository,
	arrestRepo domain.ArrestRepository,
	convRepo domain.ConvictionRepository,
	eventPub domain.EventPublisher,
	snisid domain.SNISIDClient,
) *RecordService {
	return &RecordService{
		repo:       repo,
		arrestRepo: arrestRepo,
		convRepo:   convRepo,
		eventPub:   eventPub,
		snisid:     snisid,
	}
}

func (s *RecordService) GetOrCreateRecord(
	ctx context.Context,
	snisidPersonID uuid.UUID,
) (*domain.CriminalRecord, error) {
	existing, err := s.repo.FindByPersonID(ctx, snisidPersonID)
	if err == nil && existing != nil {
		return existing, nil
	}

	person, err := s.snisid.GetPerson(snisidPersonID)
	if err != nil {
		return nil, fmt.Errorf("personne SNISID introuvable: %w", err)
	}

	nextSeq, err := s.repo.NextSequence(ctx)
	if err != nil {
		return nil, fmt.Errorf("génération séquence FIR: %w", err)
	}

	record := &domain.CriminalRecord{
		RecordID:          uuid.New(),
		NationalFIRID:     fmt.Sprintf("FIR-HT-%d-%06d", time.Now().Year(), nextSeq),
		SNISIDPersonID:    snisidPersonID,
		IsHaitianNational: person.Nationality == "HTI",
		IsActive:          true,
		IsExpunged:        false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("création casier: %w", err)
	}

	_ = s.eventPub.Publish("fir.record.created", record)
	return record, nil
}

func (s *RecordService) GetRecord(ctx context.Context, personID uuid.UUID) (*domain.CriminalRecord, error) {
	return s.repo.FindByPersonID(ctx, personID)
}

func (s *RecordService) AddArrest(ctx context.Context, recordID uuid.UUID, arrest *domain.Arrest) error {
	arrest.RecordID = recordID
	arrest.CreatedAt = time.Now()
	return s.arrestRepo.Create(ctx, arrest)
}

func (s *RecordService) AddConviction(ctx context.Context, recordID uuid.UUID, conviction *domain.Conviction) error {
	conviction.RecordID = recordID
	conviction.CreatedAt = time.Now()
	return s.convRepo.Create(ctx, conviction)
}

func (s *RecordService) GetArrests(ctx context.Context, recordID uuid.UUID) ([]*domain.Arrest, error) {
	return s.arrestRepo.FindByRecordID(ctx, recordID)
}

func (s *RecordService) GetConvictions(ctx context.Context, recordID uuid.UUID) ([]*domain.Conviction, error) {
	return s.convRepo.FindByRecordID(ctx, recordID)
}

func (s *RecordService) HasActiveConvictions(ctx context.Context, recordID uuid.UUID) (bool, error) {
	convictions, err := s.convRepo.FindByRecordID(ctx, recordID)
	if err != nil {
		return false, err
	}
	for _, c := range convictions {
		if c.CaseStatus == domain.CaseStatusConvicted || c.CaseStatus == domain.CaseStatusPendingTrial {
			return true, nil
		}
	}
	return false, nil
}
