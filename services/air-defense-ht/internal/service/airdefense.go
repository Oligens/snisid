package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/snisid/air-defense-ht/internal/domain"
	"github.com/snisid/air-defense-ht/internal/kafka"
	"github.com/snisid/air-defense-ht/internal/repository"
)

type AirDefenseServiceInterface interface {
	IngestRadarContact(domain.RadarContact) error
	GetActiveTracks() ([]domain.RadarContact, error)
	GetTrackByID(uuid.UUID) (*domain.RadarContact, error)
	OpenIncident(domain.AirDefenseIncident) error
	ResolveIncident(uuid.UUID) error
	AddNoFlyEntry(domain.NoFlyListEntry) error
	CheckNoFly(string) (*domain.NoFlyListEntry, error)
}

type AirDefenseService struct {
	repo     repository.AirDefenseRepo
	producer *kafka.Producer
}

func NewAirDefenseService(repo repository.AirDefenseRepo, producer *kafka.Producer) *AirDefenseService {
	return &AirDefenseService{repo: repo, producer: producer}
}

func (s *AirDefenseService) IngestRadarContact(c domain.RadarContact) error {
	c.ContactID = uuid.New()
	c.FirstDetectedAt = time.Now()
	c.LastUpdatedAt = time.Now()
	if c.ThreatAssessment == "" {
		c.ThreatAssessment = domain.ThreatUnknown
	}
	if c.ContactType == "" {
		c.ContactType = domain.ContactUnknown
	}
	if err := s.repo.CreateRadarContact(c); err != nil {
		return err
	}
	s.producer.Publish("radar.contact.ingested", c.ContactID.String())
	return nil
}

func (s *AirDefenseService) GetActiveTracks() ([]domain.RadarContact, error) {
	return s.repo.GetActiveContacts()
}

func (s *AirDefenseService) GetTrackByID(id uuid.UUID) (*domain.RadarContact, error) {
	return s.repo.GetContactByID(id)
}

func (s *AirDefenseService) OpenIncident(incident domain.AirDefenseIncident) error {
	incident.IncidentID = uuid.New()
	incident.CreatedAt = time.Now()
	incident.UpdatedAt = time.Now()
	incident.Status = domain.StatusDetected
	if incident.PilotResponse == "" {
		incident.PilotResponse = domain.PilotCompliant
	}
	if err := s.repo.CreateIncident(incident); err != nil {
		return err
	}
	s.producer.Publish("incident.opened", incident.IncidentID.String())
	return nil
}

func (s *AirDefenseService) ResolveIncident(id uuid.UUID) error {
	if err := s.repo.ResolveIncident(id); err != nil {
		return err
	}
	s.producer.Publish("incident.resolved", id.String())
	return nil
}

func (s *AirDefenseService) AddNoFlyEntry(entry domain.NoFlyListEntry) error {
	entry.EntryID = uuid.New()
	entry.CreatedAt = time.Now()
	if err := s.repo.CreateNoFlyEntry(entry); err != nil {
		return err
	}
	s.producer.Publish("nofly.entry.added", entry.EntryID.String())
	return nil
}

func (s *AirDefenseService) CheckNoFly(identity string) (*domain.NoFlyListEntry, error) {
	return s.repo.GetNoFlyEntry(identity)
}
