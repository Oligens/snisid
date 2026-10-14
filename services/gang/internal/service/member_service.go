package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang/internal/domain"
	"github.com/snisid/platform/services/gang/internal/repository"
)

type MemberService struct {
	memberRepo repository.MemberRepository
	gangRepo   repository.GangRepository
}

func NewMemberService(memberRepo repository.MemberRepository, gangRepo repository.GangRepository) *MemberService {
	return &MemberService{memberRepo: memberRepo, gangRepo: gangRepo}
}

func (s *MemberService) CreateMember(ctx context.Context, req domain.CreateMemberRequest, createdBy uuid.UUID) (*domain.Member, error) {
	if _, err := s.gangRepo.GetByID(ctx, req.GangID); err != nil {
		return nil, fmt.Errorf("gang introuvable: %w", err)
	}
	nationalID := fmt.Sprintf("GANG-M-%06d", time.Now().UnixMilli()%1000000)
	member := &domain.Member{
		MemberID:         uuid.New(),
		GangID:           req.GangID,
		NationalMemberID: nationalID,
		FullName:         req.FullName,
		Aliases:          req.Aliases,
		Role:             req.Role,
		DateOfBirth:      req.DateOfBirth,
		PlaceOfBirth:     req.PlaceOfBirth,
		IDType:           req.IDType,
		IDNumber:         req.IDNumber,
		IsLeader:         req.IsLeader,
		OFACDesignated:   req.OFACDesignated,
		IntelConfidence:  req.IntelConfidence,
		Nationality:      "HTI",
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("erreur création membre: %w", err)
	}
	return member, nil
}

func (s *MemberService) GetMembers(ctx context.Context, gangID uuid.UUID) ([]*domain.Member, error) {
	return s.memberRepo.ByGangID(ctx, gangID)
}
