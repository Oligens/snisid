package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/extors-svc/internal/domain"
)

type ExtorsService struct {
	repo domain.Repository
	log  *zap.Logger
}

func NewExtorsService(repo domain.Repository, log *zap.Logger) *ExtorsService {
	return &ExtorsService{repo: repo, log: log}
}

func (s *ExtorsService) OpenCase(req *domain.OpenCaseRequest) (*domain.ExtorsCase, error) {
	now := time.Now()
	nationalID := fmt.Sprintf("EXTORS-HT-%s-000000", now.Format("AAAA"))

	c := &domain.ExtorsCase{
		NationalExtorsID:  nationalID,
		ExtorsType:        req.ExtorsType,
		Status:            domain.ACTIVE,
		GangID:            req.GangID,
		GangName:          req.GangName,
		PerpetratorIDs:    req.PerpetratorIDs,
		ChefMemberIDs:     req.ChefMemberIDs,
		VictimCount:       1,
		VictimSNISIDs:     req.VictimSNISIDs,
		VictimTypes:       req.VictimTypes,
		VictimNationality: req.VictimNationality,
		IsForeignerVictim: req.IsForeignerVictim,
		IncidentLocation:  req.IncidentLocation,
		DeptCode:          req.DeptCode,
		Commune:           req.Commune,
		Lat:               req.Lat,
		Lng:               req.Lng,
		RouteNumber:       req.RouteNumber,
		DemandedAmount:    req.DemandedAmount,
		DemandedCurrency:  req.DemandedCurrency,
		FirstContactDate:  now,
		CaseReference:     req.CaseReference,
		InvestigatingUnit: req.InvestigatingUnit,
		Notes:             req.Notes,
		CreatedBy:         req.CreatedBy,
	}
	if req.VictimCount != nil {
		c.VictimCount = *req.VictimCount
	}
	if req.FirstContactDate != nil {
		c.FirstContactDate = *req.FirstContactDate
	}

	return s.repo.CreateCase(c)
}

func (s *ExtorsService) GetCaseDetail(id uuid.UUID) (*domain.ExtorsCase, error) {
	return s.repo.FindByID(id)
}

func (s *ExtorsService) AddNegotiation(caseID uuid.UUID, req *domain.AddNegotiationRequest) (*domain.Negotiation, error) {
	_, err := s.repo.FindByID(caseID)
	if err != nil {
		return nil, err
	}

	n := &domain.Negotiation{
		CaseID:          caseID,
		NegotiationDate: time.Now(),
		ContactMethod:   req.ContactMethod,
		ContactNumber:   req.ContactNumber,
		DemandUpdated:   req.DemandUpdated,
		DemandCurrency:  req.DemandCurrency,
		PositionUpdate:  req.PositionUpdate,
		RecordedBy:      req.RecordedBy,
	}
	if req.NegotiationDate != nil {
		n.NegotiationDate = *req.NegotiationDate
	}

	return s.repo.AddNegotiation(n)
}

func (s *ExtorsService) DocumentTollPoint(req *domain.CreateTollPointRequest) (*domain.RoadTollPoint, error) {
	now := time.Now()
	t := &domain.RoadTollPoint{
		GangID:            req.GangID,
		LocationDesc:      req.LocationDesc,
		RouteNumber:       req.RouteNumber,
		DeptCode:          req.DeptCode,
		Commune:           req.Commune,
		Lat:               req.Lat,
		Lng:               req.Lng,
		DailyRevenueUSD:   req.DailyRevenueUSD,
		VehicleTypesTaxed: req.VehicleTypesTaxed,
		TollRates:         req.TollRates,
		ActiveSince:       &now,
		SourceIntel:       req.SourceIntel,
		CreatedBy:         req.CreatedBy,
	}

	return s.repo.CreateTollPoint(t)
}

func (s *ExtorsService) GetTollsMapGeoJSON() (*domain.GeoJSONCollection, error) {
	tolls, err := s.repo.GetTollsMap()
	if err != nil {
		return nil, err
	}

	features := make([]domain.GeoJSONFeature, 0, len(tolls))
	for _, t := range tolls {
		features = append(features, domain.GeoJSONFeature{
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        "Point",
				"coordinates": []interface{}{t.Lng, t.Lat},
			},
			Properties: map[string]interface{}{
				"toll_id":       t.TollID,
				"gang_id":       t.GangID,
				"location_desc": t.LocationDesc,
				"route_number":  t.RouteNumber,
				"dept_code":     t.DeptCode,
				"daily_revenue": t.DailyRevenueUSD,
				"active_since":  t.ActiveSince,
			},
		})
	}

	return &domain.GeoJSONCollection{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

func (s *ExtorsService) ComputeGangRevenue(gangID uuid.UUID, gangName string) (*domain.GangRevenueReport, error) {
	report := &domain.GangRevenueReport{
		GangID:   gangID,
		GangName: gangName,
	}

	tolls, err := s.repo.FindActiveTollsByGang(gangID)
	if err != nil {
		return nil, err
	}
	for _, t := range tolls {
		if t.DailyRevenueUSD != nil {
			report.TollRevenue += *t.DailyRevenueUSD
		}
	}
	report.ActiveTolls = len(tolls)

	ransoms, err := s.repo.FindPaidRansomsByGang(gangID)
	if err != nil {
		return nil, err
	}
	for _, c := range ransoms {
		if c.PaidAmount != nil {
			report.RansomRevenue += *c.PaidAmount
		}
	}
	report.PaidRansoms = len(ransoms)

	rackets, err := s.repo.FindActiveRacketsByGang(gangID)
	if err != nil {
		return nil, err
	}
	report.ActiveRackets = len(rackets)

	report.TotalRevenue = report.TollRevenue + report.RansomRevenue

	return report, nil
}

func (s *ExtorsService) GetStatsByType() ([]domain.TypeStats, error) {
	return s.repo.GetStatsByType()
}
