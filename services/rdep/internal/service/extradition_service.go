package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/rdep/internal/domain"
)

var ErrExtraditionNotFound = errors.New("extradition non trouvée")

type ExtraditionService struct {
	mu           sync.RWMutex
	extraditions map[uuid.UUID]*domain.Extradition
	byPerson     map[uuid.UUID]uuid.UUID
	seq          int
}

func NewExtraditionService() *ExtraditionService {
	return &ExtraditionService{
		extraditions: make(map[uuid.UUID]*domain.Extradition),
		byPerson:     make(map[uuid.UUID]uuid.UUID),
	}
}

func (s *ExtraditionService) generateExtraditionID() string {
	s.seq++
	return fmt.Sprintf("RDEP-EXT-%d-%06d", time.Now().Year(), s.seq)
}

func (s *ExtraditionService) Create(ctx context.Context, req domain.CreateExtraditionRequest) (*domain.Extradition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	extradition := &domain.Extradition{
		ExtraditionID:     uuid.New(),
		NationalRDEPID:    s.generateExtraditionID(),
		SNISIDPersonID:    req.SNISIDPersonID,
		FIRRecordID:       req.FIRRecordID,
		RequestingCountry: req.RequestingCountry,
		ExtraditionStatus: domain.ExtraditionRequested,
		RequestDate:       req.RequestDate,
		ChargesSummary:    req.ChargesSummary,
		LegalReference:    req.LegalReference,
		TreatyArticle:     req.TreatyArticle,
		DeparturePort:     req.DeparturePort,
		DepartureDeptCode: req.DepartureDeptCode,
		EscortingAgency:   req.EscortingAgency,
		Notes:             req.Notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.extraditions[extradition.ExtraditionID] = extradition
	s.byPerson[req.SNISIDPersonID] = extradition.ExtraditionID

	return extradition, nil
}

func (s *ExtraditionService) GetByID(ctx context.Context, extraditionID uuid.UUID) (*domain.Extradition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	extradition, ok := s.extraditions[extraditionID]
	if !ok {
		return nil, ErrExtraditionNotFound
	}
	return extradition, nil
}

func (s *ExtraditionService) List(ctx context.Context) ([]domain.Extradition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]domain.Extradition, 0, len(s.extraditions))
	for _, e := range s.extraditions {
		results = append(results, *e)
	}
	return results, nil
}

func (s *ExtraditionService) UpdateStatus(ctx context.Context, extraditionID uuid.UUID, req domain.UpdateExtraditionStatusRequest) (*domain.Extradition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	extradition, ok := s.extraditions[extraditionID]
	if !ok {
		return nil, ErrExtraditionNotFound
	}

	extradition.ExtraditionStatus = req.Status
	if req.ExecutionDate != nil {
		extradition.ExecutionDate = req.ExecutionDate
	}
	if req.ExtraditionOfficer != nil {
		extradition.ExtraditionOfficer = req.ExtraditionOfficer
	}
	if req.Notes != nil {
		extradition.Notes = req.Notes
	}
	extradition.UpdatedAt = time.Now()

	return extradition, nil
}
