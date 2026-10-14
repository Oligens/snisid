package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/biar/internal/domain"
	"github.com/snisid/platform/services/biar/internal/repository"
)

type WeaponService struct {
	weaponRepo repository.WeaponRepository
}

func NewWeaponService(weaponRepo repository.WeaponRepository) *WeaponService {
	return &WeaponService{weaponRepo: weaponRepo}
}

func (s *WeaponService) DeclareIllicitWeapon(ctx context.Context, req domain.CreateWeaponRequest, createdBy uuid.UUID) (*domain.IllicitWeapon, error) {
	nationalID := fmt.Sprintf("BIAR-HT-%06d", time.Now().UnixMilli()%1000000)
	now := time.Now()

	w := &domain.IllicitWeapon{
		WeaponID:                 uuid.New(),
		NationalBIARID:           nationalID,
		SerialNumber:             req.SerialNumber,
		SerialObliterated:        req.SerialObliterated,
		Make:                     req.Make,
		Model:                    req.Model,
		Caliber:                  req.Caliber,
		WeaponType:               req.WeaponType,
		ManufactureCountry:       req.ManufactureCountry,
		EstimatedManufactureYear: req.EstimatedManufactureYear,

		RecoveryDate:     req.RecoveryDate,
		RecoveryContext:  req.RecoveryContext,
		RecoveryLocation: req.RecoveryLocation,
		RecoveryDeptCode: req.RecoveryDeptCode,
		RecoveryCommune:  req.RecoveryCommune,
		RecoveryLat:      req.RecoveryLat,
		RecoveryLng:      req.RecoveryLng,
		SeizingUnit:      req.SeizingUnit,
		SeizingOfficer:   req.SeizingOfficer,
		CaseReference:    req.CaseReference,

		FromPersonID:    req.FromPersonID,
		GangID:          req.GangID,
		CrimeCategory:   req.CrimeCategory,
		AssociatedCases: req.AssociatedCases,

		OriginCountry:    req.OriginCountry,
		TransitCountries: req.TransitCountries,
		TraffickingRoute: req.TraffickingRoute,
		ImportMethod:     req.ImportMethod,

		Disposition:        domain.DispositionHeldAsEvidence,
		QuantityAmmunition: req.QuantityAmmunition,
		AmmunitionType:     req.AmmunitionType,
		PhotosRefs:         req.PhotosRefs,
		Notes:              req.Notes,

		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.weaponRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("erreur déclaration arme illicite: %w", err)
	}
	return w, nil
}

func (s *WeaponService) GetWeapon(ctx context.Context, id uuid.UUID) (*domain.IllicitWeapon, error) {
	return s.weaponRepo.GetByID(ctx, id)
}

func (s *WeaponService) CheckSerial(ctx context.Context, serial string) ([]*domain.IllicitWeapon, error) {
	return s.weaponRepo.CheckSerial(ctx, serial)
}

func (s *WeaponService) ListAll(ctx context.Context) ([]*domain.IllicitWeapon, error) {
	return s.weaponRepo.List(ctx)
}
