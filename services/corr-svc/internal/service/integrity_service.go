package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/corr-svc/internal/domain"
)

type IntegrityService struct {
	repo domain.IntegrityRepository
	log  *zap.Logger
}

func NewIntegrityService(repo domain.IntegrityRepository, log *zap.Logger) *IntegrityService {
	return &IntegrityService{repo: repo, log: log}
}

func (s *IntegrityService) OpenCase(req *domain.OpenCaseRequest) (*domain.IntegrityCase, error) {
	officerID, err := uuid.Parse(req.OfficerSNISIDID)
	if err != nil {
		return nil, err
	}

	c := &domain.IntegrityCase{
		OfficerSNISIDID:   officerID,
		AllegationType:    domain.AllegationType(req.AllegationType),
		Severity:          domain.Severity(req.Severity),
		Status:            domain.Reported,
		AllegationSummary: req.AllegationSummary,
		ReportingDate:     time.Now(),
		CreatedBy:         uuid.New(),
	}

	if req.GangID != "" {
		gid, _ := uuid.Parse(req.GangID)
		c.GangID = &gid
	}

	return s.repo.CreateCase(c)
}

func (s *IntegrityService) GetCase(id uuid.UUID) (*domain.IntegrityCase, error) {
	return s.repo.FindByID(id)
}

func (s *IntegrityService) ListActiveCases() ([]domain.IntegrityCase, error) {
	return s.repo.FindActive()
}

func (s *IntegrityService) SubmitWhistleblower(req *domain.SubmitWhistleblowerRequest) (*domain.WhistleblowerReport, error) {
	token := generateToken(64)

	wr := &domain.WhistleblowerReport{
		ReportToken:    token,
		AllegationType: domain.AllegationType(req.AllegationType),
		Description:    req.Description,
		SubmissionDate: time.Now(),
	}

	if req.SeverityEstimate != "" {
		sev := domain.Severity(req.SeverityEstimate)
		wr.SeverityEstimate = &sev
	}

	return s.repo.CreateWhistleblowerReport(wr)
}

func (s *IntegrityService) TrackWhistleblower(token string) (*domain.WhistleblowerReport, error) {
	return s.repo.FindByToken(token)
}

func (s *IntegrityService) ListBehavioralAlerts() ([]domain.BehavioralAlert, error) {
	return s.repo.FindBehavioralAlerts()
}

func (s *IntegrityService) SubmitDeclaration(req *domain.SubmitDeclarationRequest) (*domain.AssetDeclaration, error) {
	officerID, err := uuid.Parse(req.OfficerSNISIDID)
	if err != nil {
		return nil, err
	}

	ad := &domain.AssetDeclaration{
		OfficerSNISIDID: officerID,
		DeclarationYear: &req.DeclarationYear,
		RealEstateUSD:   &req.RealEstateUSD,
		VehiclesUSD:     &req.VehiclesUSD,
		BankAccountsUSD: &req.BankAccountsUSD,
		OtherAssetsUSD:  &req.OtherAssetsUSD,
	}

	return s.repo.CreateAssetDeclaration(ad)
}

func (s *IntegrityService) ListFlaggedDeclarations() ([]domain.AssetDeclaration, error) {
	return s.repo.FindFlaggedDeclarations()
}

func generateToken(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
