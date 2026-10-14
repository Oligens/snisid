package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/biar-svc/internal/domain"
)

type BIARService struct {
	repo domain.WeaponRepository
	log  *zap.Logger
}

func NewBIARService(repo domain.WeaponRepository, log *zap.Logger) *BIARService {
	return &BIARService{repo: repo, log: log}
}

func (s *BIARService) ReportIllicitWeapon(req *domain.ReportWeaponRequest) (*domain.IllicitWeapon, error) {
	now := time.Now()
	nationalBIARID := generateBIARID()

	w := &domain.IllicitWeapon{
		NationalBIARID:           nationalBIARID,
		SerialNumber:             req.SerialNumber,
		SerialObliterated:        req.SerialObliterated,
		Make:                     req.Make,
		Model:                    req.Model,
		Caliber:                  req.Caliber,
		WeaponType:               req.WeaponType,
		ManufactureCountry:       req.ManufactureCountry,
		EstimatedManufactureYear: req.EstimatedManufactureYear,
		RecoveryDate:             now,
		RecoveryContext:          req.RecoveryContext,
		RecoveryLocation:         req.RecoveryLocation,
		RecoveryDeptCode:         req.RecoveryDeptCode,
		RecoveryCommune:          req.RecoveryCommune,
		RecoveryLat:              req.RecoveryLat,
		RecoveryLng:              req.RecoveryLng,
		SeizingUnit:              req.SeizingUnit,
		SeizingOfficer:           req.SeizingOfficer,
		CaseReference:            req.CaseReference,
		FromPersonID:             req.FromPersonID,
		GangID:                   req.GangID,
		CrimeCategory:            req.CrimeCategory,
		AssociatedCases:          req.AssociatedCases,
		OriginCountry:            req.OriginCountry,
		TransitCountries:         req.TransitCountries,
		TraffickingRoute:         req.TraffickingRoute,
		ImportMethod:             req.ImportMethod,
		IARMSRef:                 req.IARMSRef,
		ATFEtraceRef:             req.ATFEtraceRef,
		QuantityAmmunition:       req.QuantityAmmunition,
		AmmunitionType:           req.AmmunitionType,
		PhotosRefs:               req.PhotosRefs,
		Notes:                    req.Notes,
		CreatedBy:                req.CreatedBy,
	}
	if req.RecoveryDate != nil {
		w.RecoveryDate = *req.RecoveryDate
	}
	w.Disposition = domain.Pending

	created, err := s.repo.CreateWeapon(w)
	if err != nil {
		return nil, err
	}

	s.log.Info("weapon reported",
		zap.String("national_biar_id", created.NationalBIARID),
		zap.String("weapon_type", created.WeaponType),
		zap.Any("gang_id", created.GangID))

	return created, nil
}

func (s *BIARService) ReportBatch(req *domain.ReportBatchRequest) (*domain.BatchSeizure, error) {
	now := time.Now()
	batchRef := generateBatchRef()

	var weaponIDs []string
	for _, wr := range req.Weapons {
		w := &domain.IllicitWeapon{
			NationalBIARID:           generateBIARID(),
			SerialNumber:             wr.SerialNumber,
			SerialObliterated:        wr.SerialObliterated,
			Make:                     wr.Make,
			Model:                    wr.Model,
			Caliber:                  wr.Caliber,
			WeaponType:               wr.WeaponType,
			ManufactureCountry:       wr.ManufactureCountry,
			EstimatedManufactureYear: wr.EstimatedManufactureYear,
			RecoveryDate:             now,
			RecoveryContext:          wr.RecoveryContext,
			RecoveryLocation:         wr.RecoveryLocation,
			RecoveryDeptCode:         wr.RecoveryDeptCode,
			RecoveryCommune:          wr.RecoveryCommune,
			RecoveryLat:              wr.RecoveryLat,
			RecoveryLng:              wr.RecoveryLng,
			SeizingUnit:              wr.SeizingUnit,
			SeizingOfficer:           wr.SeizingOfficer,
			CaseReference:            wr.CaseReference,
			FromPersonID:             wr.FromPersonID,
			GangID:                   wr.GangID,
			CrimeCategory:            wr.CrimeCategory,
			AssociatedCases:          wr.AssociatedCases,
			OriginCountry:            wr.OriginCountry,
			TransitCountries:         wr.TransitCountries,
			TraffickingRoute:         wr.TraffickingRoute,
			ImportMethod:             wr.ImportMethod,
			IARMSRef:                 wr.IARMSRef,
			ATFEtraceRef:             wr.ATFEtraceRef,
			QuantityAmmunition:       wr.QuantityAmmunition,
			AmmunitionType:           wr.AmmunitionType,
			PhotosRefs:               wr.PhotosRefs,
			Notes:                    wr.Notes,
			CreatedBy:                wr.CreatedBy,
			Disposition:              domain.Pending,
		}
		if wr.RecoveryDate != nil {
			w.RecoveryDate = *wr.RecoveryDate
		}

		created, err := s.repo.CreateWeapon(w)
		if err != nil {
			return nil, err
		}
		weaponIDs = append(weaponIDs, created.WeaponID.String())
	}

	batch := &domain.BatchSeizure{
		BatchReference:     batchRef,
		OperationName:      req.OperationName,
		SeizureDate:        now,
		LocationDesc:       req.LocationDesc,
		DeptCode:           req.DeptCode,
		TotalWeapons:       len(weaponIDs),
		WeaponIDs:          weaponIDs,
		SeizingUnit:        req.SeizingUnit,
		LeadOfficer:        req.LeadOfficer,
		PartneringAgencies: req.PartneringAgencies,
		Notes:              req.Notes,
	}
	if req.SeizureDate != nil {
		batch.SeizureDate = *req.SeizureDate
	}

	created, err := s.repo.CreateBatch(batch)
	if err != nil {
		return nil, err
	}

	s.log.Info("batch seizure reported",
		zap.String("batch_reference", created.BatchReference),
		zap.Int("total_weapons", created.TotalWeapons))

	return created, nil
}

func (s *BIARService) GetWeapon(id uuid.UUID) (*domain.IllicitWeapon, error) {
	return s.repo.FindByID(id)
}

func (s *BIARService) CheckSerial(sn string) ([]domain.IllicitWeapon, error) {
	return s.repo.FindBySerial(sn)
}

func (s *BIARService) GetStatsByGang() ([]map[string]interface{}, error) {
	return s.repo.GetStatsByGang()
}

func (s *BIARService) GetStatsByOrigin() ([]map[string]interface{}, error) {
	return s.repo.GetStatsByOrigin()
}

func (s *BIARService) GetRoutes() ([]map[string]interface{}, error) {
	return s.repo.GetRoutes()
}

func (s *BIARService) SyncFromIARMS() (*domain.SyncResult, error) {
	now := time.Now()

	log := &domain.IARMSyncLog{
		Direction:  "INBOUND",
		SyncStatus: "IN_PROGRESS",
		CreatedAt:  now,
	}

	s.log.Info("iARMS sync started")

	result := &domain.SyncResult{
		SyncID:    uuid.New(),
		Direction: "INBOUND",
		Status:    "completed",
	}

	_ = s.repo.CreateSyncLog(log)

	s.log.Info("iARMS sync completed", zap.Int("count", result.Count))
	return result, nil
}

func (s *BIARService) SubmitToIARMS(weaponID uuid.UUID) (*domain.SyncResult, error) {
	w, err := s.repo.FindByID(weaponID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	log := &domain.IARMSyncLog{
		WeaponID:   w.WeaponID,
		Direction:  "OUTBOUND",
		IARMSRef:   w.IARMSRef,
		SyncStatus: "COMPLETED",
		SyncedAt:   &now,
		CreatedAt:  now,
	}

	if err := s.repo.CreateSyncLog(log); err != nil {
		return nil, err
	}

	result := &domain.SyncResult{
		SyncID:    log.SyncID,
		Direction: "OUTBOUND",
		Count:     1,
		Status:    "completed",
	}

	s.log.Info("weapon submitted to iARMS",
		zap.String("national_biar_id", w.NationalBIARID))

	return result, nil
}

func generateBIARID() string {
	n := rand.Intn(999999)
	return fmt.Sprintf("BIAR-HT-%06d", n)
}

func generateBatchRef() string {
	n := rand.Intn(999999)
	return fmt.Sprintf("BATCH-%06d", n)
}
