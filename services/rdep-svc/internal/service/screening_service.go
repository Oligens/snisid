package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/rdep-svc/internal/domain"
)

type ScreeningService struct {
	deporteeRepo domain.DeporteeRepository
	foreignRepo  domain.ForeignRecordRepository
	fbiClient    domain.FBIRecordClient
	interpClient domain.InterpolClient
	afisClient   domain.AFISClient
	eventPub     domain.EventPublisher
}

func NewScreeningService(
	deporteeRepo domain.DeporteeRepository,
	foreignRepo domain.ForeignRecordRepository,
	fbiClient domain.FBIRecordClient,
	interpClient domain.InterpolClient,
	afisClient domain.AFISClient,
	eventPub domain.EventPublisher,
) *ScreeningService {
	return &ScreeningService{
		deporteeRepo: deporteeRepo,
		foreignRepo:  foreignRepo,
		fbiClient:    fbiClient,
		interpClient: interpClient,
		afisClient:   afisClient,
		eventPub:     eventPub,
	}
}

type ScreenRequest struct {
	DeporteeID      string `json:"deportee_id" binding:"required"`
	FingerprintData string `json:"fingerprint_data"`
	FBINumber       string `json:"fbi_number"`
	GangName        string `json:"gang_name"`
}

func (s *ScreeningService) ScreenDeportee(ctx context.Context, req ScreenRequest) (*domain.ScreeningResult, error) {
	deporteeID, err := uuid.Parse(req.DeporteeID)
	if err != nil {
		return nil, fmt.Errorf("UUID invalide: %w", err)
	}

	deportee, err := s.deporteeRepo.FindByID(ctx, deporteeID)
	if err != nil {
		return nil, fmt.Errorf("déporté introuvable: %w", err)
	}

	result := &domain.ScreeningResult{
		PersonID:  deportee.SNISIDPersonID,
		RiskLevel: domain.RiskNone,
	}

	if req.FingerprintData != "" && s.afisClient != nil {
		afisHit, err := s.afisClient.CheckPrint(ctx, req.FingerprintData)
		if err == nil && afisHit != nil {
			result.HasLocalRecord = true
			result.LocalRecordID = &afisHit.SubjectID
			result.RiskLevel = elevateRisk(result.RiskLevel, domain.RiskMedium)
		}
	}

	if req.FBINumber != "" && s.fbiClient != nil {
		fbiRecord, err := s.fbiClient.GetRecord(ctx, req.FBINumber)
		if err == nil && fbiRecord != nil {
			result.HasForeignRecord = true
			result.ForeignRecords = append(result.ForeignRecords, *fbiRecord)
			if fbiRecord.HasViolentOffenses() {
				result.RiskLevel = elevateRisk(result.RiskLevel, domain.RiskHigh)
			}
		}
	}

	gangName := req.GangName
	if gangName == "" {
		gangName = deportee.GangName
	}
	if gangName != "" {
		result.GangAffiliated = true
		result.RiskLevel = elevateRisk(result.RiskLevel, domain.RiskVeryHigh)
	}

	if s.interpClient != nil {
		notices, err := s.interpClient.CheckNotices(ctx, deportee.SNISIDPersonID)
		if err == nil && len(notices) > 0 {
			result.InterpolNotices = notices
			result.RiskLevel = elevateRisk(result.RiskLevel, domain.RiskVeryHigh)
		}
	}

	deportee.CriminalRiskLevel = result.RiskLevel
	deportee.HasForeignRecord = result.HasForeignRecord
	deportee.GangAffiliated = result.GangAffiliated
	deportee.MonitoringRequired = result.RiskLevel == domain.RiskHigh || result.RiskLevel == domain.RiskVeryHigh
	deportee.UpdatedAt = time.Now()
	_ = s.deporteeRepo.Update(ctx, deportee)

	_ = s.eventPub.Publish("rdep.screening.completed", result)

	return result, nil
}

func elevateRisk(current, proposed domain.CriminalRisk) domain.CriminalRisk {
	riskOrder := map[domain.CriminalRisk]int{
		domain.RiskNone:     0,
		domain.RiskLow:      1,
		domain.RiskMedium:   2,
		domain.RiskHigh:     3,
		domain.RiskVeryHigh: 4,
	}
	if riskOrder[proposed] > riskOrder[current] {
		return proposed
	}
	return current
}
