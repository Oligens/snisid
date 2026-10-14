package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/idcore-svc/internal/domain"
	"github.com/snisid/idcore-svc/internal/kafka"
	"github.com/snisid/idcore-svc/internal/milvus"
	"github.com/snisid/idcore-svc/internal/nin"
	"github.com/snisid/idcore-svc/internal/repository"
)

type IdentityService struct {
	repo          repository.Repository
	milvusClient  *milvus.Client
	producer      *kafka.Producer
	ninGenerator  *nin.Generator
	bioThreshold  float64
	demoThreshold float64
}

func NewIdentityService(
	repo repository.Repository,
	milvusClient *milvus.Client,
	producer *kafka.Producer,
	ninGenerator *nin.Generator,
	bioThreshold string,
	demoThreshold string,
) *IdentityService {
	bioThresh, _ := strconv.ParseFloat(bioThreshold, 64)
	if bioThresh == 0 {
		bioThresh = 0.95
	}
	demoThresh, _ := strconv.ParseFloat(demoThreshold, 64)
	if demoThresh == 0 {
		demoThresh = 0.85
	}
	return &IdentityService{
		repo:          repo,
		milvusClient:  milvusClient,
		producer:      producer,
		ninGenerator:  ninGenerator,
		bioThreshold:  bioThresh,
		demoThreshold: demoThresh,
	}
}

func (s *IdentityService) EnrollCitizen(ctx context.Context, req domain.EnrollmentRequest) (*domain.EnrollmentResult, error) {
	age := int(time.Since(req.DOB).Hours() / 24 / 365)

	if age >= 5 {
		bioCheck, err := s.milvusClient.CheckDuplicate(ctx, req.BiometricSample)
		if err != nil {
			return nil, fmt.Errorf("biometric verification: %w", err)
		}
		if bioCheck.HasMatch && bioCheck.Confidence > s.bioThreshold {
			_ = s.repo.CreateDedupCandidate(ctx, domain.DedupCandidate{
				CitizenIDA:     bioCheck.MatchedCitizenID,
				BiometricScore: bioCheck.Confidence,
			})
			return nil, domain.ErrDuplicateDetected
		}
	}

	demoMatches, err := s.repo.FindDemographicMatches(ctx, req.FullNameLegal, req.DOB)
	if err != nil {
		log.Printf("demographic match query failed: %v", err)
	} else {
		for _, m := range demoMatches {
			_ = s.repo.CreateDedupCandidate(ctx, domain.DedupCandidate{
				CitizenIDA:       m.CitizenID,
				DemographicScore: m.Score,
			})
		}
	}

	nin, err := s.ninGenerator.Generate(ctx, req.DeptCode, req.DOB.Year())
	if err != nil {
		return nil, fmt.Errorf("NIN generation: %w", err)
	}

	createdByUUID, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		createdByUUID = uuid.New()
	}

	citizen := &domain.Citizen{
		CitizenID:      uuid.New(),
		NIN:            nin,
		Status:         domain.StatusActive,
		EnrollmentType: req.EnrollmentType,
		FullNameLegal:  req.FullNameLegal,
		FirstName:      req.FirstName,
		MiddleNames:    req.MiddleNames,
		LastName:       req.LastName,
		MaidenName:     req.MaidenName,
		DOB:            req.DOB,
		PobCommune:     req.PobCommune,
		PobDeptCode:    req.PobDeptCode,
		Gender:         req.Gender,
		Nationality:    req.Nationality,
		DeptCode:       req.DeptCode,
		CurrentAddress: req.CurrentAddress,
		CurrentCommune: req.CurrentCommune,
		PhotoRef:       req.PhotoRef,
		MotherNIN:      req.MotherNIN,
		FatherNIN:      req.FatherNIN,
		CreatedBy:      createdByUUID,
	}

	if err := s.repo.Create(ctx, citizen); err != nil {
		return nil, fmt.Errorf("create citizen: %w", err)
	}

	if req.BiometricSample != nil {
		bioTemplateID, err := s.milvusClient.StoreTemplate(ctx, citizen.CitizenID, req.BiometricSample)
		if err != nil {
			log.Printf("failed to store biometric template: %v", err)
		} else {
			citizen.BiometricTemplateID = &bioTemplateID
			_ = s.repo.Update(ctx, citizen)
		}
	}

	s.publishEvent(ctx, "identity.citizen.enrolled", citizen, req.CreatedBy)
	return &domain.EnrollmentResult{Citizen: citizen, NIN: nin}, nil
}

func (s *IdentityService) VerifyIdentity(ctx context.Context, nin string) (*domain.Citizen, error) {
	return s.repo.FindByNIN(ctx, nin)
}

func (s *IdentityService) GetCitizen(ctx context.Context, id string) (*domain.Citizen, error) {
	citizenUUID, err := uuid.Parse(id)
	if err != nil {
		return s.repo.FindByNIN(ctx, id)
	}
	return s.repo.FindByID(ctx, citizenUUID)
}

func (s *IdentityService) UpdateStatus(ctx context.Context, nin string, status domain.IDStatus, reason string, authorizedBy string) error {
	authUUID, err := uuid.Parse(authorizedBy)
	if err != nil {
		authUUID = uuid.Nil
	}
	return s.repo.UpdateStatus(ctx, nin, status, reason, authUUID)
}

func (s *IdentityService) GetHistory(ctx context.Context, id string) ([]domain.ChangeHistory, error) {
	citizenUUID, err := uuid.Parse(id)
	if err != nil {
		cit, findErr := s.repo.FindByNIN(ctx, id)
		if findErr != nil {
			return nil, findErr
		}
		citizenUUID = cit.CitizenID
	}
	return s.repo.GetHistory(ctx, citizenUUID)
}

func (s *IdentityService) GetPopulationStats(ctx context.Context) (*domain.PopulationStats, error) {
	return s.repo.GetPopulationStats(ctx)
}

func (s *IdentityService) SearchCitizens(ctx context.Context, query string) ([]*domain.Citizen, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *IdentityService) ResolveDedup(ctx context.Context, candidateID string, resolution string, reviewedBy string) error {
	return fmt.Errorf("not implemented")
}

func (s *IdentityService) publishEvent(ctx context.Context, eventType string, citizen *domain.Citizen, actorID string) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType:     eventType,
		CitizenID:     citizen.CitizenID.String(),
		NIN:           citizen.NIN,
		CorrelationID: uuid.New().String(),
		ActorID:       actorID,
		Timestamp:     time.Now().UTC(),
		Data:          citizen,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
