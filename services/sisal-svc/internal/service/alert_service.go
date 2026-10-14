package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/sisal-svc/internal/domain"
)

type AlertService struct {
	repo domain.AlertRepository
	log  *zap.Logger
}

func NewAlertService(repo domain.AlertRepository, log *zap.Logger) *AlertService {
	return &AlertService{repo: repo, log: log}
}

func (s *AlertService) IssueAlert(req *domain.IssueAlertRequest) (*domain.SISALAlert, error) {
	alert := &domain.SISALAlert{
		HazardType:    domain.HazardType(req.HazardType),
		Severity:      domain.Severity(req.Severity),
		Title:         req.Title,
		MessageFR:     req.MessageFR,
		MessageHT:     req.MessageHT,
		AffectedDepts: req.AffectedDepts,
		IssuedAt:      time.Now(),
		SourceAgency:  req.SourceAgency,
		CreatedBy:     uuid.New(),
	}

	if req.AffectedPopEst != nil {
		alert.AffectedPopEst = req.AffectedPopEst
	}

	return s.repo.Create(alert)
}

func (s *AlertService) ListActiveAlerts() ([]domain.SISALAlert, error) {
	return s.repo.FindActive()
}

func (s *AlertService) ListHistory() ([]domain.SISALAlert, error) {
	return s.repo.FindHistory()
}

func (s *AlertService) CancelAlert(id uuid.UUID, reason string) error {
	return s.repo.Cancel(id, reason)
}

func (s *AlertService) Subscribe(req *domain.SubscribeRequest) (*domain.Subscription, error) {
	sub := &domain.Subscription{
		MinSeverity: domain.Warning,
	}

	if req.PhoneNumber != "" {
		sub.PhoneNumber = &req.PhoneNumber
	}
	if req.Email != "" {
		sub.Email = &req.Email
	}
	if req.DeptCode != "" {
		sub.DeptCode = &req.DeptCode
	}
	if req.Commune != "" {
		sub.Commune = &req.Commune
	}
	if req.MinSeverity != "" {
		sub.MinSeverity = domain.Severity(req.MinSeverity)
	}

	return s.repo.CreateSubscription(sub)
}
