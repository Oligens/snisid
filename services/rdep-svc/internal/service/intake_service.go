package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/rdep-svc/internal/domain"
)

type IntakeService struct {
	deporteeRepo domain.DeporteeRepository
	foreignRepo  domain.ForeignRecordRepository
	eventPub     domain.EventPublisher
}

func NewIntakeService(
	deporteeRepo domain.DeporteeRepository,
	foreignRepo domain.ForeignRecordRepository,
	eventPub domain.EventPublisher,
) *IntakeService {
	return &IntakeService{
		deporteeRepo: deporteeRepo,
		foreignRepo:  foreignRepo,
		eventPub:     eventPub,
	}
}

func (s *IntakeService) ProcessIntake(ctx context.Context, req domain.DeporteeIntakeRequest) (*domain.Deportee, error) {
	personID, err := uuid.Parse(req.SNISIDPersonID)
	if err != nil {
		return nil, fmt.Errorf("UUID invalide: %w", err)
	}

	existing, _ := s.deporteeRepo.FindByPersonID(ctx, personID)
	if existing != nil {
		return nil, fmt.Errorf("déporté déjà enregistré")
	}

	deportationDate, err := time.Parse("2006-01-02", req.DeportationDate)
	if err != nil {
		return nil, fmt.Errorf("date de déportation invalide: %w", err)
	}

	deportee := &domain.Deportee{
		DeporteeID:         uuid.New(),
		NationalRdepID:     fmt.Sprintf("RDEP-HT-%d-%s", time.Now().Year(), uuid.New().String()[:8]),
		SNISIDPersonID:     personID,
		DeportationCountry: domain.DeportationCountry(req.DeportationCountry),
		DeportationDate:    deportationDate,
		ArrivalPort:        req.ArrivalPort,
		DeportingAgency:    req.DeportingAgency,
		DeportationReason:  req.DeportationReason,
		FlightNumber:       req.FlightNumber,
		ForeignName:        req.ForeignName,
		ForeignIDNumber:    req.ForeignIDNumber,
		GangName:           req.GangName,
		CurrentAddress:     req.CurrentAddress,
		CurrentCommune:     req.CurrentCommune,
		CriminalRiskLevel:  domain.RiskNone,
		MonitoringStatus:   domain.MonitoringActive,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.deporteeRepo.Create(ctx, deportee); err != nil {
		return nil, fmt.Errorf("création déporté: %w", err)
	}

	_ = s.eventPub.Publish("rdep.deportee.intake", map[string]interface{}{
		"deportee_id": deportee.DeporteeID,
		"person_id":   personID,
		"country":     req.DeportationCountry,
		"intake_at":   time.Now(),
	})

	return deportee, nil
}

func (s *IntakeService) GetDeportee(ctx context.Context, deporteeID uuid.UUID) (*domain.Deportee, error) {
	return s.deporteeRepo.FindByID(ctx, deporteeID)
}

func (s *IntakeService) GetHighRisk(ctx context.Context) ([]*domain.Deportee, error) {
	return s.deporteeRepo.FindHighRisk(ctx)
}

func (s *IntakeService) GetGangAffiliated(ctx context.Context) ([]*domain.Deportee, error) {
	return s.deporteeRepo.FindGangAffiliated(ctx)
}

func (s *IntakeService) GetStatsByCountry(ctx context.Context) (map[string]int, error) {
	return s.deporteeRepo.GetStatsByCountry(ctx)
}
