package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/trait-svc/internal/domain"
)

type TraiffickingService struct {
	repo domain.TraiffickingRepository
	log  *zap.Logger
}

func NewTraiffickingService(repo domain.TraiffickingRepository, log *zap.Logger) *TraiffickingService {
	return &TraiffickingService{repo: repo, log: log}
}

func (s *TraiffickingService) OpenCase(req *domain.OpenCaseRequest) (*domain.TraiffickingCase, error) {
	year := time.Now().Year()
	seq, err := s.repo.CountCasesByTypeAndYear(req.TrafficType, year)
	if err != nil {
		return nil, err
	}

	nationalID := fmt.Sprintf("TRAIT-HT-%d-%06d", year, seq+1)

	status := req.Status
	if status == "" {
		status = "OPEN"
	}

	origin := req.OriginCountry
	if origin == "" {
		origin = "HTI"
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	c := &domain.TraiffickingCase{
		NationalTraitID:    nationalID,
		TrafficType:        req.TrafficType,
		Status:             status,
		VictimCount:        0,
		MinorCount:         0,
		OriginCountry:      origin,
		TransitCountries:   req.TransitCountries,
		DestinationCountry: req.DestinationCountry,
		RouteDescription:   req.RouteDescription,
		TransportMode:      req.TransportMode,
		MarIncidentID:      req.MarIncidentID,
		GangID:             req.GangID,
		Currency:           currency,
		InvestigatingUnit:  req.InvestigatingUnit,
		CaseReference:      req.CaseReference,
		IomCaseRef:         req.IomCaseRef,
		CreatedBy:          req.CreatedBy,
	}

	if c.TransitCountries == nil {
		c.TransitCountries = []string{}
	}
	if c.TransportMode == nil {
		c.TransportMode = []string{}
	}
	if c.SifrcrossingIDs == nil {
		c.SifrcrossingIDs = []uuid.UUID{}
	}
	if c.RecruiterIDs == nil {
		c.RecruiterIDs = []uuid.UUID{}
	}

	return s.repo.CreateCase(c)
}

func (s *TraiffickingService) GetCaseDetail(id uuid.UUID) (*domain.TraiffickingCase, error) {
	return s.repo.FindByID(id)
}

func (s *TraiffickingService) AddVictim(caseID uuid.UUID, req *domain.AddVictimRequest) (*domain.TraiffickingVictim, error) {
	nationality := req.Nationality
	if nationality == "" {
		nationality = "HTI"
	}

	v := &domain.TraiffickingVictim{
		CaseID:             caseID,
		SnisidPersonID:     req.SnisidPersonID,
		VictimStatus:       req.VictimStatus,
		FullName:           req.FullName,
		Nationality:        nationality,
		Dob:                req.Dob,
		Gender:             req.Gender,
		IsMinor:            req.IsMinor,
		ExploitationType:   req.ExploitationType,
		RescueDate:         req.RescueDate,
		RescueLocation:     req.RescueLocation,
		CurrentLocation:    req.CurrentLocation,
		AssistanceProvided: req.AssistanceProvided,
		DipeCaseID:         req.DipeCaseID,
		AfisSubjectID:      req.AfisSubjectID,
	}

	if v.AssistanceProvided == nil {
		v.AssistanceProvided = []string{}
	}

	return s.repo.AddVictim(v)
}

func (s *TraiffickingService) GetMinorVictims() ([]domain.TraiffickingVictim, error) {
	return s.repo.GetMinorVictims()
}

func (s *TraiffickingService) DocumentNetwork(req *domain.DocumentNetworkRequest) (*domain.TraiffickingNetwork, error) {
	n := &domain.TraiffickingNetwork{
		NetworkName:      req.NetworkName,
		PrimaryRoute:     req.PrimaryRoute,
		OriginDept:       req.OriginDept,
		KnownMembers:     req.KnownMembers,
		GangAffiliations: req.GangAffiliations,
		MonthlyVolumeEst: req.MonthlyVolumeEst,
		FeePerPersonUsd:  req.FeePerPersonUsd,
		IntelConfidence:  req.IntelConfidence,
		CreatedBy:        req.CreatedBy,
	}

	if n.KnownMembers == nil {
		n.KnownMembers = []uuid.UUID{}
	}
	if n.GangAffiliations == nil {
		n.GangAffiliations = []uuid.UUID{}
	}
	if n.LinkedCases == nil {
		n.LinkedCases = []uuid.UUID{}
	}

	return s.repo.CreateNetwork(n)
}

func (s *TraiffickingService) GetActiveNetworks() ([]domain.TraiffickingNetwork, error) {
	return s.repo.GetActiveNetworks()
}

func (s *TraiffickingService) GetStatsByType() ([]domain.TypeStats, error) {
	return s.repo.GetStatsByType()
}

func (s *TraiffickingService) GetMaritimeCases() ([]domain.TraiffickingCase, error) {
	return s.repo.GetMaritimeCases()
}
