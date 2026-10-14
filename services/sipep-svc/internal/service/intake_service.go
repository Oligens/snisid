package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep-svc/internal/domain"
)

type IntakeService struct {
	inmateRepo    domain.InmateRepository
	detentionRepo domain.DetentionRepository
	eventPub      domain.EventPublisher
	snisid        domain.SNISIDClient
}

func NewIntakeService(
	inmateRepo domain.InmateRepository,
	detentionRepo domain.DetentionRepository,
	eventPub domain.EventPublisher,
	snisid domain.SNISIDClient,
) *IntakeService {
	return &IntakeService{
		inmateRepo:    inmateRepo,
		detentionRepo: detentionRepo,
		eventPub:      eventPub,
		snisid:        snisid,
	}
}

type IntakeRequest struct {
	SNISIDPersonID     string `json:"snisid_person_id" binding:"required"`
	Facility           string `json:"facility" binding:"required"`
	DetentionBasis     string `json:"detention_basis" binding:"required"`
	CaseReference      string `json:"case_reference"`
	CourtName          string `json:"court_name"`
	ArrestingAuthority string `json:"arresting_authority"`
	WarrantNumber      string `json:"warrant_number"`
	IntakeOfficer      string `json:"intake_officer" binding:"required"`
	IsMinor            bool   `json:"is_minor"`
	IsFemale           bool   `json:"is_female"`
	HasSpecialNeeds    bool   `json:"has_special_needs"`
	SpecialNeedsNotes  string `json:"special_needs_notes"`
}

func (s *IntakeService) ProcessIntake(ctx context.Context, req IntakeRequest) (*domain.Inmate, *domain.Detention, error) {
	personID, err := uuid.Parse(req.SNISIDPersonID)
	if err != nil {
		return nil, nil, fmt.Errorf("UUID invalide: %w", err)
	}

	intakeOfficer, err := uuid.Parse(req.IntakeOfficer)
	if err != nil {
		return nil, nil, fmt.Errorf("UUID officier invalide: %w", err)
	}

	existing, _ := s.inmateRepo.FindByPersonID(ctx, personID)
	if existing != nil && existing.IsCurrentlyDetained {
		return nil, nil, fmt.Errorf("personne déjà incarcérée")
	}

	inmate := &domain.Inmate{
		InmateID:            uuid.New(),
		NationalInmateID:    fmt.Sprintf("SIPEP-HT-%d-%s", time.Now().Year(), uuid.New().String()[:8]),
		SNISIDPersonID:      personID,
		CurrentFacility:     req.Facility,
		IsCurrentlyDetained: true,
		IsMinor:             req.IsMinor,
		IsFemale:            req.IsFemale,
		HasSpecialNeeds:     req.HasSpecialNeeds,
		SpecialNeedsNotes:   req.SpecialNeedsNotes,
		IntakeDate:          time.Now(),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.inmateRepo.Create(ctx, inmate); err != nil {
		return nil, nil, fmt.Errorf("création détenu: %w", err)
	}

	detention := &domain.Detention{
		DetentionID:        uuid.New(),
		InmateID:           inmate.InmateID,
		Facility:           req.Facility,
		DetentionBasis:     domain.DetentionBasis(req.DetentionBasis),
		LegalStatus:        domain.LegalStatusAwaitingTrial,
		CaseReference:      req.CaseReference,
		CourtName:          req.CourtName,
		ArrestingAuthority: req.ArrestingAuthority,
		WarrantNumber:      req.WarrantNumber,
		IntakeDate:         time.Now(),
		IntakeOfficer:      intakeOfficer,
		CreatedAt:          time.Now(),
	}

	if err := s.detentionRepo.Create(ctx, detention); err != nil {
		return nil, nil, fmt.Errorf("création détention: %w", err)
	}

	_ = s.eventPub.Publish("sipep.inmate.intake", map[string]interface{}{
		"inmate_id": inmate.InmateID,
		"person_id": personID,
		"facility":  req.Facility,
		"intake_at": time.Now(),
	})

	return inmate, detention, nil
}

func (s *IntakeService) GetInmate(ctx context.Context, inmateID uuid.UUID) (*domain.Inmate, error) {
	return s.inmateRepo.FindByID(ctx, inmateID)
}

func (s *IntakeService) SearchInmates(ctx context.Context, query string) ([]*domain.Inmate, error) {
	return s.inmateRepo.Search(ctx, query)
}
