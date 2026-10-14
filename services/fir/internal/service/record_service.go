package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir/internal/domain"
)

var ErrRecordNotFound = errors.New("casier judiciaire non trouvé")
var ErrPersonNotFound = errors.New("personne SNISID introuvable")

type RecordService struct {
	mu       sync.RWMutex
	records  map[uuid.UUID]*domain.CriminalRecord
	bySNISID map[uuid.UUID]uuid.UUID
	seq      int
}

func NewRecordService() *RecordService {
	return &RecordService{
		records:  make(map[uuid.UUID]*domain.CriminalRecord),
		bySNISID: make(map[uuid.UUID]uuid.UUID),
	}
}

func (s *RecordService) generateFIRID() string {
	s.seq++
	return fmt.Sprintf("FIR-HT-%d-%06d", time.Now().Year(), s.seq)
}

func (s *RecordService) Create(ctx context.Context, personID uuid.UUID, isHaitian bool, afisSubjectID *uuid.UUID) (*domain.CriminalRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.bySNISID[personID]; exists {
		return nil, fmt.Errorf("casier existe déjà pour la personne: %s", personID)
	}

	record := &domain.CriminalRecord{
		RecordID:          uuid.New(),
		NationalFIRID:     s.generateFIRID(),
		SNISIDPersonID:    personID,
		AFISSubjectID:     afisSubjectID,
		IsHaitianNational: isHaitian,
		IsActive:          true,
		IsExpunged:        false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	s.records[record.RecordID] = record
	s.bySNISID[personID] = record.RecordID

	return record, nil
}

func (s *RecordService) GetByID(ctx context.Context, recordID uuid.UUID) (*domain.CriminalRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[recordID]
	if !ok {
		return nil, ErrRecordNotFound
	}
	return record, nil
}

func (s *RecordService) GetByPersonID(ctx context.Context, personID uuid.UUID) (*domain.CriminalRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recordID, ok := s.bySNISID[personID]
	if !ok {
		return nil, ErrRecordNotFound
	}
	record, ok := s.records[recordID]
	if !ok {
		return nil, ErrRecordNotFound
	}
	return record, nil
}

func (s *RecordService) List(ctx context.Context) ([]domain.CriminalRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]domain.CriminalRecord, 0, len(s.records))
	for _, rec := range s.records {
		records = append(records, *rec)
	}
	return records, nil
}

func (s *RecordService) Expunge(ctx context.Context, recordID uuid.UUID) (*domain.CriminalRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[recordID]
	if !ok {
		return nil, ErrRecordNotFound
	}

	record.IsExpunged = true
	record.IsActive = false
	record.UpdatedAt = time.Now()

	return record, nil
}

func (s *RecordService) Reactivate(ctx context.Context, recordID uuid.UUID) (*domain.CriminalRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[recordID]
	if !ok {
		return nil, ErrRecordNotFound
	}

	record.IsExpunged = false
	record.IsActive = true
	record.UpdatedAt = time.Now()

	return record, nil
}

func (s *RecordService) Update(ctx context.Context, recordID uuid.UUID, isHaitian *bool, afisSubjectID *uuid.UUID) (*domain.CriminalRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[recordID]
	if !ok {
		return nil, ErrRecordNotFound
	}

	if isHaitian != nil {
		record.IsHaitianNational = *isHaitian
	}
	if afisSubjectID != nil {
		record.AFISSubjectID = afisSubjectID
	}
	record.UpdatedAt = time.Now()

	return record, nil
}

func (s *RecordService) Search(ctx context.Context, query string) ([]domain.CriminalRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []domain.CriminalRecord
	for _, rec := range s.records {
		if rec.NationalFIRID == query {
			results = append(results, *rec)
			continue
		}
	}
	return results, nil
}
