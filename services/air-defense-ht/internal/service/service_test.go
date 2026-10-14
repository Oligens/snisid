package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/air-defense-ht/internal/domain"
	"github.com/snisid/air-defense-ht/internal/kafka"
)

type mockRepo struct {
	createContactFn   func(domain.RadarContact) error
	getActiveFn       func() ([]domain.RadarContact, error)
	getByIDFn         func(uuid.UUID) (*domain.RadarContact, error)
	createIncidentFn  func(domain.AirDefenseIncident) error
	resolveIncidentFn func(uuid.UUID) error
	createNoFlyFn     func(domain.NoFlyListEntry) error
	getNoFlyFn        func(string) (*domain.NoFlyListEntry, error)
}

func (m *mockRepo) CreateRadarContact(c domain.RadarContact) error {
	return m.createContactFn(c)
}
func (m *mockRepo) GetActiveContacts() ([]domain.RadarContact, error) {
	return m.getActiveFn()
}
func (m *mockRepo) GetContactByID(id uuid.UUID) (*domain.RadarContact, error) {
	return m.getByIDFn(id)
}
func (m *mockRepo) CreateIncident(i domain.AirDefenseIncident) error {
	return m.createIncidentFn(i)
}
func (m *mockRepo) ResolveIncident(id uuid.UUID) error {
	return m.resolveIncidentFn(id)
}
func (m *mockRepo) CreateNoFlyEntry(e domain.NoFlyListEntry) error {
	return m.createNoFlyFn(e)
}
func (m *mockRepo) GetNoFlyEntry(identity string) (*domain.NoFlyListEntry, error) {
	return m.getNoFlyFn(identity)
}

func TestIngestRadarContact(t *testing.T) {
	repo := &mockRepo{
		createContactFn: func(c domain.RadarContact) error {
			if c.TrackNumber != "TRK001" {
				return errors.New("unexpected track number")
			}
			return nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	c := domain.RadarContact{TrackNumber: "TRK001", Latitude: 10.0, Longitude: 20.0}
	if err := svc.IngestRadarContact(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIngestRadarContactRepoError(t *testing.T) {
	repo := &mockRepo{
		createContactFn: func(c domain.RadarContact) error {
			return errors.New("db error")
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	c := domain.RadarContact{TrackNumber: "TRK001", Latitude: 10.0, Longitude: 20.0}
	if err := svc.IngestRadarContact(c); err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestGetActiveTracks(t *testing.T) {
	repo := &mockRepo{
		getActiveFn: func() ([]domain.RadarContact, error) {
			return []domain.RadarContact{{TrackNumber: "TRK001"}}, nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	tracks, err := svc.GetActiveTracks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
}

func TestGetTrackByID(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{
		getByIDFn: func(uid uuid.UUID) (*domain.RadarContact, error) {
			return &domain.RadarContact{ContactID: uid}, nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	track, err := svc.GetTrackByID(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track == nil {
		t.Fatal("expected track, got nil")
	}
	if track.ContactID != id {
		t.Fatalf("expected id %v, got %v", id, track.ContactID)
	}
}

func TestGetTrackByIDNotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(uid uuid.UUID) (*domain.RadarContact, error) {
			return nil, nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	track, err := svc.GetTrackByID(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track != nil {
		t.Fatal("expected nil track")
	}
}

func TestOpenIncident(t *testing.T) {
	repo := &mockRepo{
		createIncidentFn: func(i domain.AirDefenseIncident) error {
			if i.AircraftID == uuid.Nil {
				return errors.New("expected non-nil aircraft_id")
			}
			return nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	incident := domain.AirDefenseIncident{AircraftID: uuid.New()}
	if err := svc.OpenIncident(incident); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenIncidentRepoError(t *testing.T) {
	repo := &mockRepo{
		createIncidentFn: func(i domain.AirDefenseIncident) error {
			return errors.New("db error")
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	incident := domain.AirDefenseIncident{AircraftID: uuid.New()}
	if err := svc.OpenIncident(incident); err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestResolveIncident(t *testing.T) {
	repo := &mockRepo{
		resolveIncidentFn: func(id uuid.UUID) error {
			return nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	if err := svc.ResolveIncident(uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddNoFlyEntry(t *testing.T) {
	repo := &mockRepo{
		createNoFlyFn: func(e domain.NoFlyListEntry) error {
			if e.IdentityRef == "" {
				return errors.New("expected identity_ref")
			}
			return nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	entry := domain.NoFlyListEntry{IdentityRef: "ID123", FullName: "John Doe", Reason: "test", AddedBy: "admin"}
	if err := svc.AddNoFlyEntry(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckNoFlyFound(t *testing.T) {
	repo := &mockRepo{
		getNoFlyFn: func(identity string) (*domain.NoFlyListEntry, error) {
			return &domain.NoFlyListEntry{IdentityRef: identity}, nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	entry, err := svc.CheckNoFly("ID123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
}

func TestCheckNoFlyNotFound(t *testing.T) {
	repo := &mockRepo{
		getNoFlyFn: func(identity string) (*domain.NoFlyListEntry, error) {
			return nil, nil
		},
	}
	svc := NewAirDefenseService(repo, &kafka.Producer{})
	entry, err := svc.CheckNoFly("UNKNOWN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry != nil {
		t.Fatal("expected nil entry")
	}
}
