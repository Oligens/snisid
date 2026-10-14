package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/aero-svc/internal/domain"
)

type AeroService struct {
	repo domain.Repository
	log  *zap.Logger
}

func NewAeroService(repo domain.Repository, log *zap.Logger) *AeroService {
	return &AeroService{repo: repo, log: log}
}

func (s *AeroService) CheckRegistration(reg string) (*domain.RegistrationCheckResult, error) {
	aircraft, err := s.repo.FindByRegistration(reg)
	if err != nil {
		return &domain.RegistrationCheckResult{
			IsRegistered: false,
			IsSuspected:  false,
			IsStolen:     false,
		}, nil
	}

	return &domain.RegistrationCheckResult{
		IsRegistered: aircraft.IsRegistered,
		IsSuspected:  aircraft.IsSuspected,
		IsStolen:     aircraft.IsStolen,
		Aircraft:     aircraft,
	}, nil
}

func (s *AeroService) ReportStrip(req *domain.ReportStripRequest) (*domain.ClandestineStrip, error) {
	strip := &domain.ClandestineStrip{
		StripName:         req.StripName,
		DeptCode:          req.DeptCode,
		Commune:           req.Commune,
		Lat:               req.Lat,
		Lng:               req.Lng,
		LengthM:           req.LengthM,
		SurfaceType:       req.SurfaceType,
		Status:            domain.ACTIVE,
		CapableAircraft:   req.CapableAircraft,
		SourceIntel:       req.SourceIntel,
		SatelliteImageRef: req.SatelliteImageRef,
		CreatedBy:         uuid.New(),
	}

	if req.Status != "" {
		strip.Status = domain.StripStatus(req.Status)
	}

	if req.GangID != nil {
		gangID, err := uuid.Parse(*req.GangID)
		if err == nil {
			strip.GangID = &gangID
		}
	}

	if req.FirstDetected != nil {
		if t, err := time.Parse("2006-01-02", *req.FirstDetected); err == nil {
			strip.FirstDetected = &t
		}
	}

	if req.LastActivityDate != nil {
		if t, err := time.Parse("2006-01-02", *req.LastActivityDate); err == nil {
			strip.LastActivityDate = &t
		}
	}

	return s.repo.CreateStrip(strip)
}

func (s *AeroService) GetStripMap() (*domain.GeoJSONFeatureCollection, error) {
	strips, err := s.repo.GetStripsMap()
	if err != nil {
		return nil, err
	}

	fc := &domain.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]domain.GeoJSONFeature, 0, len(strips)),
	}

	for _, strip := range strips {
		props := map[string]interface{}{
			"strip_id":   strip.StripID.String(),
			"strip_name": strip.StripName,
			"dept_code":  strip.DeptCode,
			"status":     string(strip.Status),
			"created_at": strip.CreatedAt.Format(time.RFC3339),
		}
		if strip.Commune != nil {
			props["commune"] = *strip.Commune
		}
		if strip.LengthM != nil {
			props["length_m"] = *strip.LengthM
		}
		if strip.SurfaceType != nil {
			props["surface_type"] = *strip.SurfaceType
		}
		if strip.GangID != nil {
			props["gang_id"] = strip.GangID.String()
		}

		feature := domain.GeoJSONFeature{
			Type: "Feature",
			Geometry: domain.GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{strip.Lng, strip.Lat},
			},
			Properties: props,
		}
		fc.Features = append(fc.Features, feature)
	}

	return fc, nil
}

func (s *AeroService) ReportSuspiciousFlight(req *domain.ReportFlightRequest) (*domain.SuspiciousFlight, error) {
	flight := &domain.SuspiciousFlight{
		RegistrationMark:   &req.RegistrationMark,
		OriginAirport:      &req.OriginAirport,
		DestinationAirport: &req.DestinationAirport,
		OriginCountry:      &req.OriginCountry,
		DestinationCountry: &req.DestinationCountry,
		LandingLocation:    &req.LandingLocation,
		FlightType:         &req.FlightType,
		CargoSuspected:     &req.CargoSuspected,
		SourceRadar:        &req.SourceRadar,
		SourceInformant:    req.SourceInformant,
		CaseReference:      &req.CaseReference,
		CreatedBy:          uuid.New(),
	}

	if t, err := time.Parse(time.RFC3339, req.FlightDate); err == nil {
		flight.FlightDate = t
	} else if t, err := time.Parse("2006-01-02", req.FlightDate); err == nil {
		flight.FlightDate = t
	}

	if req.AircraftID != "" {
		if id, err := uuid.Parse(req.AircraftID); err == nil {
			flight.AircraftID = &id
		}
	}

	if req.LandingStripID != "" {
		if id, err := uuid.Parse(req.LandingStripID); err == nil {
			flight.LandingStripID = &id
		}
	}

	if req.DestinationCountry == "" {
		defaultCountry := "HTI"
		flight.DestinationCountry = &defaultCountry
	}

	result, err := s.repo.CreateFlight(flight)
	if err != nil {
		return nil, fmt.Errorf("failed to report flight: %w", err)
	}

	return result, nil
}

func (s *AeroService) GetStripStats() (*domain.StripStats, error) {
	return s.repo.GetStripStats()
}
