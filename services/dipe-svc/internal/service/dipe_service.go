package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/dipe-svc/internal/domain"
)

type MissingPersonService struct {
	repo domain.MissingRepository
	log  *zap.Logger
}

func NewMissingPersonService(repo domain.MissingRepository, log *zap.Logger) *MissingPersonService {
	return &MissingPersonService{repo: repo, log: log}
}

func (s *MissingPersonService) ReportDisappearance(req *domain.ReportDisappearanceRequest) (*domain.MissingPerson, error) {
	now := time.Now()
	nationalID := fmt.Sprintf("DIPE-HT-%s-%06d",
		now.Format("2006"),
		rand.Intn(1000000))

	lastSeen := now
	if req.LastSeenDate != nil {
		lastSeen = *req.LastSeenDate
	}

	caseObj := &domain.MissingPerson{
		NationalDipeID:      nationalID,
		CaseType:            req.CaseType,
		Status:              domain.CaseStatusOpen,
		SnisidPersonID:      req.SnisidPersonID,
		FullName:            req.FullName,
		Aliases:             req.Aliases,
		DOB:                 req.DOB,
		Gender:              req.Gender,
		Nationality:         req.Nationality,
		Occupation:          req.Occupation,
		PhotoRefs:           req.PhotoRefs,
		HeightCM:            req.HeightCM,
		WeightKG:            req.WeightKG,
		SkinTone:            req.SkinTone,
		EyeColor:            req.EyeColor,
		HairColor:           req.HairColor,
		DistinguishingMarks: req.DistinguishingMarks,
		ClothingLastSeen:    req.ClothingLastSeen,
		LastSeenDate:        lastSeen,
		LastSeenLocation:    req.LastSeenLocation,
		LastSeenDeptCode:    req.LastSeenDeptCode,
		LastSeenCommune:     req.LastSeenCommune,
		LastSeenLat:         req.LastSeenLat,
		LastSeenLng:         req.LastSeenLng,
		Circumstances:       req.Circumstances,
		SivcAlertID:         req.SivcAlertID,
		GangID:              req.GangID,
		ExtorsCaseID:        req.ExtorsCaseID,
		ReportedByName:      req.ReportedByName,
		ReportedByPhone:     req.ReportedByPhone,
		ReportedBySnisid:    req.ReportedBySnisid,
		ReportDate:          now,
		ReportingUnit:       req.ReportingUnit,
	}

	if err := s.repo.CreateCase(caseObj); err != nil {
		s.log.Error("failed to create case", zap.Error(err))
		return nil, err
	}
	s.log.Info("case created", zap.String("case_id", caseObj.CaseID.String()), zap.String("national_dipe_id", nationalID))
	return caseObj, nil
}

func (s *MissingPersonService) GetCaseDetail(id uuid.UUID) (*domain.MissingPerson, error) {
	return s.repo.FindByID(id)
}

func (s *MissingPersonService) GetOpenCases(limit, offset int) ([]*domain.MissingPerson, int, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetOpenCases(limit, offset)
}

func (s *MissingPersonService) AddSighting(caseID uuid.UUID, req *domain.AddSightingRequest) (*domain.Sighting, error) {
	sighting := &domain.Sighting{
		CaseID:       caseID,
		SightingDate: req.SightingDate,
		LocationDesc: req.LocationDesc,
		DeptCode:     req.DeptCode,
		Lat:          req.Lat,
		Lng:          req.Lng,
		ReportedBy:   req.ReportedBy,
		ReportMethod: req.ReportMethod,
		Confidence:   req.Confidence,
		PhotoRef:     req.PhotoRef,
	}
	if err := s.repo.AddSighting(sighting); err != nil {
		s.log.Error("failed to add sighting", zap.Error(err), zap.String("case_id", caseID.String()))
		return nil, err
	}
	s.log.Info("sighting added", zap.String("case_id", caseID.String()), zap.String("sighting_id", sighting.SightingID.String()))
	return sighting, nil
}

func (s *MissingPersonService) MatchWithRVIN(id uuid.UUID) (*domain.MatchResult, error) {
	person, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	sightings, err := s.repo.GetSightingsByCase(id)
	if err != nil {
		return nil, err
	}

	query := &domain.MorphQuery{
		HeightCM: person.HeightCM,
		WeightKG: person.WeightKG,
		Gender:   person.Gender,
		SkinTone: person.SkinTone,
		DeptCode: person.LastSeenDeptCode,
	}

	result := &domain.MatchResult{
		Person:     person,
		Sightings:  sightings,
		HasMatch:   false,
		Confidence: 0,
	}

	_ = query

	return result, nil
}

func (s *MissingPersonService) ResolveCase(id uuid.UUID, req *domain.ResolveCaseRequest) error {
	if err := s.repo.ResolveCase(id, req.Status, req.ResolutionNotes); err != nil {
		s.log.Error("failed to resolve case", zap.Error(err), zap.String("case_id", id.String()))
		return err
	}
	s.log.Info("case resolved", zap.String("case_id", id.String()), zap.String("status", string(req.Status)))
	return nil
}

func (s *MissingPersonService) GetStatsByType() (map[domain.CaseType]int, error) {
	return s.repo.GetStatsByType()
}
