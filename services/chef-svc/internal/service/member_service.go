package service

import (
	"time"

	"github.com/snisid/platform/services/chef-svc/internal/domain"
)

type MemberService struct {
	memberRepo domain.MemberRepository
	intelRepo  domain.IntelNoteRepository
	sightRepo  domain.SightingRepository
	publisher  domain.EventPublisher
}

func NewMemberService(mr domain.MemberRepository, ir domain.IntelNoteRepository, sr domain.SightingRepository, p domain.EventPublisher) *MemberService {
	return &MemberService{
		memberRepo: mr,
		intelRepo:  ir,
		sightRepo:  sr,
		publisher:  p,
	}
}

func (s *MemberService) CreateMember(m *domain.CriminalMember) error {
	if m.Status == "" {
		m.Status = domain.StatusActive
	}
	if err := s.memberRepo.Create(m); err != nil {
		return err
	}
	s.publisher.PublishEvent("member.created", m)
	return nil
}

func (s *MemberService) GetMember(id string) (*domain.CriminalMember, error) {
	return s.memberRepo.GetByID(id)
}

func (s *MemberService) GetByGang(gangID string) ([]domain.CriminalMember, error) {
	return s.memberRepo.GetByGang(gangID)
}

func (s *MemberService) GetSanctioned() ([]domain.CriminalMember, error) {
	return s.memberRepo.GetSanctioned()
}

func (s *MemberService) GetLeaders() ([]domain.CriminalMember, error) {
	return s.memberRepo.GetLeaders()
}

func (s *MemberService) UpdateStatus(id string, status domain.MemberStatus) error {
	if err := s.memberRepo.UpdateStatus(id, status); err != nil {
		return err
	}
	s.publisher.PublishEvent("member.status_updated", map[string]interface{}{
		"member_id":  id,
		"status":     status,
		"updated_at": time.Now(),
	})
	return nil
}

func (s *MemberService) AddIntelligenceNote(note *domain.IntelligenceNote) error {
	return s.intelRepo.Create(note)
}

func (s *MemberService) GetIntelligenceNotes(memberID string) ([]domain.IntelligenceNote, error) {
	return s.intelRepo.GetByMemberID(memberID)
}

func (s *MemberService) RecordSighting(sighting *domain.Sighting) error {
	return s.sightRepo.Create(sighting)
}

func (s *MemberService) GetSightings(memberID string) ([]domain.Sighting, error) {
	return s.sightRepo.GetByMemberID(memberID)
}
