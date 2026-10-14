package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

type LatentService struct {
	latentRepo LatentRepo
	imageRepo  ImageRepo
	searchSvc  *SearchService
}

func NewLatentService(lr LatentRepo, ir ImageRepo, ss *SearchService) *LatentService {
	return &LatentService{
		latentRepo: lr,
		imageRepo:  ir,
		searchSvc:  ss,
	}
}

func (s *LatentService) Submit(ctx context.Context, req domain.LatentSubmission) (*domain.LatentPrint, error) {
	imgData, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	latentID := uuid.New()
	objectName := fmt.Sprintf("latents/%s/%s.png", req.CaseReference, latentID.String())
	if err := s.imageRepo.Upload(ctx, objectName, imgData, "image/png"); err != nil {
		return nil, fmt.Errorf("upload latent: %w", err)
	}

	lp := &domain.LatentPrint{
		LatentID:       latentID,
		CaseReference:  req.CaseReference,
		CrimeSceneID:   req.CrimeSceneID,
		LocationDesc:   req.LocationDesc,
		DeptCode:       req.DeptCode,
		FoundAt:        req.FoundAt,
		ImageRef:       objectName,
		FingerPosition: req.Position,
		ExaminedBy:     &req.ExaminedBy,
		CreatedAt:      time.Now(),
	}

	if err := s.latentRepo.Create(ctx, lp); err != nil {
		return nil, fmt.Errorf("save latent: %w", err)
	}
	return lp, nil
}

func (s *LatentService) SearchLatent(ctx context.Context, latentID uuid.UUID) ([]domain.SearchResult, error) {
	lp, err := s.latentRepo.GetByID(ctx, latentID)
	if err != nil {
		return nil, fmt.Errorf("get latent: %w", err)
	}

	imgData, err := s.imageRepo.Download(ctx, lp.ImageRef)
	if err != nil {
		return nil, fmt.Errorf("download latent image: %w", err)
	}

	capture := domain.FingerprintCapture{
		Position:    lp.FingerPosition,
		Method:      domain.CaptureLatentLift,
		ImageBase64: base64.StdEncoding.EncodeToString(imgData),
	}

	results, err := s.searchSvc.SearchTenprint(ctx, domain.EnrollmentRequest{
		SubjectType:  domain.SubjectSuspect,
		Fingerprints: []domain.FingerprintCapture{capture},
	})
	if err != nil {
		return nil, fmt.Errorf("search tenprint: %w", err)
	}
	return results, nil
}

func (s *LatentService) ConfirmMatch(ctx context.Context, latentID uuid.UUID, req domain.LatentMatchConfirm) error {
	return s.latentRepo.ConfirmMatch(ctx, latentID, req.MatchedSubjectID, req.MatchScore, req.ExaminedBy)
}
