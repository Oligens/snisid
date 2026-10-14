package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep/internal/domain"
)

type InmateService struct {
	mu         sync.RWMutex
	inmates    map[uuid.UUID]*domain.Inmate
	detentions map[uuid.UUID][]*domain.Detention
}

func NewInmateService() *InmateService {
	return &InmateService{
		inmates:    make(map[uuid.UUID]*domain.Inmate),
		detentions: make(map[uuid.UUID][]*domain.Detention),
	}
}

func (s *InmateService) Intake(req domain.IntakeRequest) (*domain.Inmate, *domain.Detention, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inmateID := uuid.New()
	nationalID := fmt.Sprintf("SIPEP-HT-%s-%06d", time.Now().Format("2006"), rand.Intn(999999))

	now := time.Now()
	inmate := &domain.Inmate{
		InmateID:            inmateID,
		NationalInmateID:    nationalID,
		SNISIDPersonID:      req.SNISIDPersonID,
		CurrentFacility:     req.Facility,
		CellBlock:           req.CellBlock,
		IsCurrentlyDetained: true,
		IsMinor:             req.IsMinor,
		IsFemale:            req.IsFemale,
		HasSpecialNeeds:     req.HasSpecialNeeds,
		SpecialNeedsNotes:   req.SpecialNeedsNotes,
		IntakeDate:          now,
		ExpectedReleaseDate: req.ExpectedReleaseDate,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	legalStatus := req.LegalStatus
	if legalStatus == "" {
		legalStatus = domain.LegalStatusAwaitingTrial
	}

	detention := &domain.Detention{
		DetentionID:          uuid.New(),
		InmateID:             inmateID,
		Facility:             req.Facility,
		DetentionBasis:       req.DetentionBasis,
		LegalStatus:          legalStatus,
		CaseReference:        req.CaseReference,
		CourtName:            req.CourtName,
		ArrestingAuthority:   req.ArrestingAuthority,
		WarrantNumber:        req.WarrantNumber,
		IntakeDate:           now,
		IntakeOfficer:        req.IntakeOfficer,
		SentenceDurationDays: req.SentenceDurationDays,
		CreatedAt:            now,
	}

	s.inmates[inmateID] = inmate
	s.detentions[inmateID] = []*domain.Detention{detention}

	return inmate, detention, nil
}

func (s *InmateService) GetInmate(id uuid.UUID) (*domain.Inmate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inmate, ok := s.inmates[id]
	if !ok {
		return nil, fmt.Errorf("inmate not found: %s", id)
	}
	return inmate, nil
}

func (s *InmateService) Search(query string) ([]*domain.Inmate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*domain.Inmate
	for _, inmate := range s.inmates {
		if query == "" ||
			inmate.NationalInmateID == query ||
			inmate.SNISIDPersonID.String() == query ||
			inmate.CurrentFacility == query {
			results = append(results, inmate)
		}
	}
	return results, nil
}

func (s *InmateService) ProcessRelease(inmateID uuid.UUID, req domain.ReleaseRequest, authorizedBy uuid.UUID) (*domain.Detention, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inmate, ok := s.inmates[inmateID]
	if !ok {
		return nil, fmt.Errorf("inmate not found: %s", inmateID)
	}
	if !inmate.IsCurrentlyDetained {
		return nil, fmt.Errorf("inmate is not currently detained")
	}

	detentions, ok := s.detentions[inmateID]
	if !ok || len(detentions) == 0 {
		return nil, fmt.Errorf("no active detention record found")
	}

	activeDet := detentions[len(detentions)-1]
	now := time.Now()
	activeDet.ReleaseDate = &now
	rt := req.ReleaseType
	activeDet.ReleaseType = &rt
	activeDet.ReleasingAuthority = req.Authority

	inmate.IsCurrentlyDetained = false
	inmate.UpdatedAt = now

	return activeDet, nil
}

func (s *InmateService) UpdateInmate(inmate *domain.Inmate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.inmates[inmate.InmateID]; !ok {
		return fmt.Errorf("inmate not found: %s", inmate.InmateID)
	}
	inmate.UpdatedAt = time.Now()
	s.inmates[inmate.InmateID] = inmate
	return nil
}

func (s *InmateService) RecordDetention(detention *domain.Detention) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.detentions[detention.InmateID] = append(s.detentions[detention.InmateID], detention)
	return nil
}
