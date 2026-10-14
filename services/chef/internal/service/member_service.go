package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/chef/internal/domain"
)

type MemberService struct {
	repo *InMemoryRepository
}

func NewMemberService(repo *InMemoryRepository) *MemberService {
	return &MemberService{repo: repo}
}

func (s *MemberService) CreateMember(ctx context.Context, req domain.CreateMemberRequest) (*domain.CriminalMember, error) {
	now := time.Now()
	status := req.Status
	if status == "" {
		status = domain.StatusActive
	}

	member := &domain.CriminalMember{
		MemberID:            uuid.New(),
		NationalChefID:      s.repo.NextNationalChefID(),
		SNISIDPersonID:      req.SNISIDPersonID,
		FIRRecordID:         req.FIRRecordID,
		AFISSubjectID:       req.AFISSubjectID,
		RDePDeporteeID:      req.RDePDeporteeID,
		PrimaryGangID:       req.PrimaryGangID,
		RoleInGang:          req.RoleInGang,
		RoleDescription:     req.RoleDescription,
		Aliases:             req.Aliases,
		TerritoryDept:       req.TerritoryDept,
		KnownArmed:          req.KnownArmed,
		Status:              status,
		IntelClassification: "SECRET",
		CreatedBy:           req.CreatedBy,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.repo.Create(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *MemberService) GetMember(ctx context.Context, memberID uuid.UUID) (*domain.CriminalMember, error) {
	return s.repo.FindByID(ctx, memberID)
}

func (s *MemberService) GetMembersByGang(ctx context.Context, gangID uuid.UUID) ([]*domain.CriminalMember, error) {
	return s.repo.FindByGang(ctx, gangID)
}

func (s *MemberService) GetSanctionedMembers(ctx context.Context) ([]*domain.CriminalMember, error) {
	return s.repo.FindSanctioned(ctx)
}

func (s *MemberService) GetActiveLeaders(ctx context.Context) ([]*domain.CriminalMember, error) {
	return s.repo.FindLeaders(ctx)
}

func (s *MemberService) UpdateStatus(ctx context.Context, memberID uuid.UUID, newStatus domain.ChefStatus, updatedBy uuid.UUID, notes string) error {
	member, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return err
	}

	old := member.Status
	member.Status = newStatus
	member.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, member); err != nil {
		return err
	}

	_ = domain.StatusChangedEvent{
		MemberID:  memberID,
		GangID:    member.PrimaryGangID,
		OldStatus: old,
		NewStatus: newStatus,
		ChangedBy: updatedBy,
		Notes:     notes,
	}

	if newStatus == domain.StatusArrested {
		_ = domain.MemberArrestedEvent{
			MemberID:   memberID,
			PersonID:   member.SNISIDPersonID,
			GangID:     member.PrimaryGangID,
			RoleInGang: member.RoleInGang,
		}
	}

	return nil
}

func (s *MemberService) AddIntelNote(ctx context.Context, memberID uuid.UUID, req domain.CreateIntelNoteRequest, createdBy uuid.UUID) (*domain.IntelNote, error) {
	_, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	note := &domain.IntelNote{
		NoteID:        uuid.New(),
		MemberID:      memberID,
		NoteDate:      time.Now(),
		IntelType:     req.IntelType,
		Content:       req.Content,
		SourceClassif: req.SourceClassif,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateNote(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *MemberService) GetIntelNotes(ctx context.Context, memberID uuid.UUID) ([]*domain.IntelNote, error) {
	return s.repo.FindNotesByMember(ctx, memberID)
}

func (s *MemberService) AddSighting(ctx context.Context, memberID uuid.UUID, req domain.CreateSightingRequest, reportedBy uuid.UUID) (*domain.Sighting, error) {
	_, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	sighting := &domain.Sighting{
		SightingID:   uuid.New(),
		MemberID:     memberID,
		SightedAt:    req.SightedAt,
		LocationDesc: req.LocationDesc,
		DeptCode:     req.DeptCode,
		Commune:      req.Commune,
		Lat:          req.Lat,
		Lng:          req.Lng,
		SourceType:   req.SourceType,
		Confidence:   req.Confidence,
		PhotoRef:     req.PhotoRef,
		ReportedBy:   reportedBy,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateSighting(ctx, sighting); err != nil {
		return nil, err
	}
	return sighting, nil
}

func (s *MemberService) GetSightings(ctx context.Context, memberID uuid.UUID) ([]*domain.Sighting, error) {
	return s.repo.FindSightingsByMember(ctx, memberID)
}

func (s *MemberService) GetMemberNetwork(ctx context.Context, memberID uuid.UUID) ([]*domain.CrossGangLink, error) {
	return s.repo.FindLinksByMember(ctx, memberID)
}
