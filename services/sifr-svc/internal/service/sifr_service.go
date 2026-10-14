package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/sifr-svc/internal/domain"
)

type BorderService struct {
	repo domain.BorderRepository
	log  *zap.Logger
}

func NewBorderService(repo domain.BorderRepository, log *zap.Logger) *BorderService {
	return &BorderService{repo: repo, log: log}
}

func (s *BorderService) ProcessCrossing(req *domain.CrossingRequest) (*domain.Crossing, *domain.CrossingResult, error) {
	result := &domain.CrossingResult{
		Clearance:   domain.ClearanceGranted,
		IsDangerous: false,
	}

	if req.PostID == uuid.Nil {
		return nil, nil, fmt.Errorf("post_id is required")
	}
	dir := domain.CrossingDirection(req.Direction)
	if dir != domain.ENTRY && dir != domain.EXIT {
		return nil, nil, fmt.Errorf("direction must be ENTRY or EXIT")
	}
	docType := domain.DocType(req.DocumentType)
	if docType == domain.NONE && req.SNISIDPersonID == nil {
		return nil, nil, fmt.Errorf("document_type or snisid_person_id is required")
	}

	now := time.Now().UTC()
	crossing := &domain.Crossing{
		PostID:              req.PostID,
		Direction:           dir,
		CrossingDatetime:    now,
		SNISIDPersonID:      req.SNISIDPersonID,
		DocumentType:        docType,
		DocumentNumber:      req.DocumentNumber,
		DocumentCountry:     req.DocumentCountry,
		DocumentExpiry:      req.DocumentExpiry,
		TravelerName:        req.TravelerName,
		TravelerDob:         req.TravelerDob,
		TravelerNationality: req.TravelerNationality,
		VehiclePlate:        req.VehiclePlate,
		LaneNumber:          req.LaneNumber,
		ProcessingOfficer:   req.ProcessingOfficer,
	}
	if req.CrossingDatetime != nil {
		crossing.CrossingDatetime = *req.CrossingDatetime
	}

	alertTriggered, alertType, alertSource := s.checkAlerts(crossing)
	crossing.AlertTriggered = alertTriggered
	if alertTriggered {
		crossing.AlertType = alertType
		result.Clearance = domain.ClearanceDenied
		result.IsDangerous = true
		result.AlertType = alertType
		result.AlertSource = alertSource
	}

	saved, err := s.repo.CreateCrossing(crossing)
	if err != nil {
		return nil, nil, fmt.Errorf("save crossing: %w", err)
	}

	s.log.Info("crossing processed",
		zap.String("crossing_id", saved.CrossingID.String()),
		zap.String("direction", string(saved.Direction)),
		zap.Bool("alert", alertTriggered),
	)

	return saved, result, nil
}

func (s *BorderService) checkAlerts(crossing *domain.Crossing) (bool, *domain.AlertType, string) {
	if crossing.DocumentExpiry != nil && crossing.DocumentExpiry.Before(time.Now()) {
		t := domain.STOLEN_DOCUMENT
		return true, &t, "DOCUMENT_EXPIRY_CHECK"
	}

	if crossing.SNISIDPersonID != nil {
		s.log.Info("checking person against watchlists",
			zap.String("person_id", crossing.SNISIDPersonID.String()),
		)
	}

	return false, nil, ""
}

func (s *BorderService) SearchCrossings(postID *uuid.UUID, direction string, dateFrom, dateTo *time.Time, page, pageSize int) ([]domain.Crossing, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.FindCrossingsByPost(postID, pageSize, offset)
}

func (s *BorderService) GetPersonHistory(personID uuid.UUID) ([]domain.Crossing, error) {
	if personID == uuid.Nil {
		return nil, fmt.Errorf("person_id is required")
	}
	return s.repo.FindCrossingsByPerson(personID)
}

func (s *BorderService) GetActiveAlerts() ([]domain.AlertLog, error) {
	return s.repo.FindActiveAlerts()
}

func (s *BorderService) ListPosts() ([]domain.BorderPost, error) {
	return s.repo.GetBorderPosts()
}

func (s *BorderService) GetDailyStats(postID *uuid.UUID) (map[string]interface{}, error) {
	return s.repo.GetDailyStats(postID)
}

func (s *BorderService) ReportClandestineCrossing(req *domain.ClandestineCrossing) (*domain.ClandestineCrossing, error) {
	if req.EstimatedPersons < 1 {
		return nil, fmt.Errorf("estimated_persons must be at least 1")
	}
	if req.ReportedBy == uuid.Nil {
		return nil, fmt.Errorf("reported_by is required")
	}
	return s.repo.CreateClandestineReport(req)
}
