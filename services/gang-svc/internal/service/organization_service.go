package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang-svc/internal/domain"
)

type OrganizationService struct {
	orgRepo  domain.OrganizationRepository
	incRepo  domain.IncidentRepository
	alliRepo domain.AllianceRepository
	eventPub domain.EventPublisher
}

func NewOrganizationService(
	orgRepo domain.OrganizationRepository,
	incRepo domain.IncidentRepository,
	alliRepo domain.AllianceRepository,
	eventPub domain.EventPublisher,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:  orgRepo,
		incRepo:  incRepo,
		alliRepo: alliRepo,
		eventPub: eventPub,
	}
}

type CreateOrgRequest struct {
	Name                       string   `json:"name" binding:"required"`
	Aliases                    []string `json:"aliases"`
	StructureType              string   `json:"structure_type"`
	PrimaryActivity            string   `json:"primary_activity" binding:"required"`
	ActivityLevel              string   `json:"activity_level"`
	EstimatedMembers           int      `json:"estimated_members"`
	ArmedMembersPct            int      `json:"armed_members_pct"`
	HeavyWeapons               bool     `json:"heavy_weapons"`
	PrimaryDeptCode            string   `json:"primary_dept_code" binding:"required"`
	TerritoryCommunes          []string `json:"territory_communes"`
	EstimatedRevenueUSDMonthly float64  `json:"estimated_revenue_usd_monthly"`
	PrimaryIncomeSources       []string `json:"primary_income_sources"`
	OFACDesignation            bool     `json:"ofac_designation"`
	OFACSDNRef                 string   `json:"ofac_sdn_ref"`
	EstablishedDate            string   `json:"established_date"`
	IntelConfidence            int      `json:"intel_confidence"`
	CreatedBy                  string   `json:"created_by" binding:"required"`
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*domain.Organization, error) {
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("UUID invalide: %w", err)
	}

	org := &domain.Organization{
		GangID:                     uuid.New(),
		NationalGangID:             fmt.Sprintf("GANG-HT-%s", uuid.New().String()[:8]),
		Name:                       req.Name,
		Aliases:                    req.Aliases,
		StructureType:              domain.StructureType(req.StructureType),
		PrimaryActivity:            domain.PrimaryActivity(req.PrimaryActivity),
		ActivityLevel:              domain.ActivityLevel(req.ActivityLevel),
		EstimatedMembers:           req.EstimatedMembers,
		ArmedMembersPct:            req.ArmedMembersPct,
		HeavyWeapons:               req.HeavyWeapons,
		PrimaryDeptCode:            req.PrimaryDeptCode,
		TerritoryCommunes:          req.TerritoryCommunes,
		EstimatedRevenueUSDMonthly: req.EstimatedRevenueUSDMonthly,
		PrimaryIncomeSources:       req.PrimaryIncomeSources,
		OFACDesignation:            req.OFACDesignation,
		OFACSDNRef:                 req.OFACSDNRef,
		IntelConfidence:            req.IntelConfidence,
		IsActive:                   true,
		CreatedBy:                  createdBy,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	if req.EstablishedDate != "" {
		if t, err := time.Parse("2006-01-02", req.EstablishedDate); err == nil {
			org.EstablishedDate = &t
		}
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, nil
	}

	_ = s.eventPub.Publish("gang.organization.created", org)
	return org, nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, gangID uuid.UUID) (*domain.Organization, error) {
	return s.orgRepo.FindByID(ctx, gangID)
}

func (s *OrganizationService) ListOrganizations(ctx context.Context) ([]*domain.Organization, error) {
	return s.orgRepo.FindAll(ctx)
}

func (s *OrganizationService) GetByDeptCode(ctx context.Context, deptCode string) ([]*domain.Organization, error) {
	return s.orgRepo.FindByDeptCode(ctx, deptCode)
}

func (s *OrganizationService) GetSanctioned(ctx context.Context) ([]*domain.Organization, error) {
	return s.orgRepo.FindSanctioned(ctx)
}
