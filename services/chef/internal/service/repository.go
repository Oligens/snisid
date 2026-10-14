package service

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/chef/internal/domain"
)

type InMemoryRepository struct {
	mu        sync.RWMutex
	members   map[uuid.UUID]*domain.CriminalMember
	notes     map[uuid.UUID][]*domain.IntelNote
	sightings map[uuid.UUID][]*domain.Sighting
	links     []*domain.CrossGangLink
	nextSeq   int
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		members:   make(map[uuid.UUID]*domain.CriminalMember),
		notes:     make(map[uuid.UUID][]*domain.IntelNote),
		sightings: make(map[uuid.UUID][]*domain.Sighting),
		nextSeq:   1,
	}
}

func (r *InMemoryRepository) NextNationalChefID() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextSeq
	r.nextSeq++
	return formatNationalChefID(id)
}

func formatNationalChefID(seq int) string {
	s := ""
	n := seq
	for n > 0 {
		s = string(rune('A'+rune((n-1)%26))) + s
		n = (n - 1) / 26
	}
	if s == "" {
		s = "A"
	}
	for len(s) < 6 {
		s = "A" + s
	}
	return "CHEF-HT-" + s
}

func (r *InMemoryRepository) Create(ctx context.Context, m *domain.CriminalMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[m.MemberID] = m
	return nil
}

func (r *InMemoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.CriminalMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.members[id]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	return m, nil
}

func (r *InMemoryRepository) Update(ctx context.Context, m *domain.CriminalMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[m.MemberID] = m
	return nil
}

func (r *InMemoryRepository) FindByGang(ctx context.Context, gangID uuid.UUID) ([]*domain.CriminalMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.CriminalMember
	for _, m := range r.members {
		if m.PrimaryGangID == gangID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) FindSanctioned(ctx context.Context) ([]*domain.CriminalMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.CriminalMember
	for _, m := range r.members {
		if m.UNDesignated || m.OFACDesignated {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) FindLeaders(ctx context.Context) ([]*domain.CriminalMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.CriminalMember
	for _, m := range r.members {
		if m.Status == domain.StatusActive && (m.RoleInGang == domain.RoleSupremeLeader || m.RoleInGang == domain.RoleZoneCommander) {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) CreateNote(ctx context.Context, n *domain.IntelNote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notes[n.MemberID] = append(r.notes[n.MemberID], n)
	return nil
}

func (r *InMemoryRepository) FindNotesByMember(ctx context.Context, memberID uuid.UUID) ([]*domain.IntelNote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	notes, ok := r.notes[memberID]
	if !ok {
		return []*domain.IntelNote{}, nil
	}
	return notes, nil
}

func (r *InMemoryRepository) CreateSighting(ctx context.Context, s *domain.Sighting) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sightings[s.MemberID] = append(r.sightings[s.MemberID], s)
	return nil
}

func (r *InMemoryRepository) FindSightingsByMember(ctx context.Context, memberID uuid.UUID) ([]*domain.Sighting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sightings, ok := r.sightings[memberID]
	if !ok {
		return []*domain.Sighting{}, nil
	}
	return sightings, nil
}

func (r *InMemoryRepository) CreateLink(ctx context.Context, l *domain.CrossGangLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links = append(r.links, l)
	return nil
}

func (r *InMemoryRepository) FindLinksByMember(ctx context.Context, memberID uuid.UUID) ([]*domain.CrossGangLink, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.CrossGangLink
	for _, l := range r.links {
		if l.MemberAID == memberID || l.MemberBID == memberID {
			result = append(result, l)
		}
	}
	return result, nil
}
