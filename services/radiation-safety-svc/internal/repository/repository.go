package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/snisid/radiation-safety-svc/internal/domain"
)

type RadiationRepository interface {
	CreateSource(ctx context.Context, s *domain.RadioactiveSource) error
	UpdateSourceStatus(ctx context.Context, id uuid.UUID, status domain.SourceStatus) error
	CreateAlert(ctx context.Context, a *domain.RadiationAlert) error
	GetUnrespondedAlerts(ctx context.Context) ([]domain.RadiationAlert, error)
	CreateChemical(ctx context.Context, c *domain.ChemicalPrecursor) error
	GetSuspiciousChemicals(ctx context.Context) ([]domain.ChemicalPrecursor, error)
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
}

type DashboardStats struct {
	TotalSources      int `json:"total_sources"`
	LostStolenSources int `json:"lost_stolen_sources"`
	UnrespondedAlerts int `json:"unresponded_alerts"`
	SuspiciousChems   int `json:"suspicious_chemicals"`
}

type radiationRepo struct {
	db *sql.DB
}

func NewRadiationRepository(db *sql.DB) RadiationRepository {
	return &radiationRepo{db: db}
}

func (r *radiationRepo) CreateSource(ctx context.Context, s *domain.RadioactiveSource) error {
	query := `INSERT INTO radiation_sources
		(source_id, source_type, isotope, activity_curie, location_building, location_lat, location_lng,
		 custodian_org, license_ref, status, last_verified_at, last_inventory_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.db.ExecContext(ctx, query,
		s.SourceID, s.SourceType, s.Isotope, s.ActivityCurie, s.LocationBuilding,
		s.LocationLat, s.LocationLng, s.CustodianOrg, s.LicenseRef, s.Status,
		s.LastVerifiedAt, s.LastInventoryAt)
	return err
}

func (r *radiationRepo) UpdateSourceStatus(ctx context.Context, id uuid.UUID, status domain.SourceStatus) error {
	query := `UPDATE radiation_sources SET status = $1 WHERE source_id = $2`
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *radiationRepo) CreateAlert(ctx context.Context, a *domain.RadiationAlert) error {
	query := `INSERT INTO radiation_alerts
		(detector_id, detector_location, detected_isotope, dose_rate_usv, alert_level,
		 responded_by, response_notes, cleared_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.ExecContext(ctx, query,
		a.DetectorID, a.DetectorLocation, a.DetectedIsotope, a.DoseRateUSv, a.AlertLevel,
		a.RespondedBy, a.ResponseNotes, a.ClearedAt, a.CreatedAt)
	return err
}

func (r *radiationRepo) GetUnrespondedAlerts(ctx context.Context) ([]domain.RadiationAlert, error) {
	query := `SELECT detector_id, detector_location, detected_isotope, dose_rate_usv, alert_level,
		responded_by, response_notes, cleared_at, created_at
		FROM radiation_alerts WHERE responded_by IS NULL ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.RadiationAlert
	for rows.Next() {
		var a domain.RadiationAlert
		if err := rows.Scan(&a.DetectorID, &a.DetectorLocation, &a.DetectedIsotope, &a.DoseRateUSv,
			&a.AlertLevel, &a.RespondedBy, &a.ResponseNotes, &a.ClearedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	if result == nil {
		result = []domain.RadiationAlert{}
	}
	return result, rows.Err()
}

func (r *radiationRepo) CreateChemical(ctx context.Context, c *domain.ChemicalPrecursor) error {
	query := `INSERT INTO radiation_chemicals
		(substance_name, cas_number, category, quantity_kg, storage_location, importer_entity,
		 end_user, end_use, permit_ref, reported_suspicious, flagged_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.ExecContext(ctx, query,
		c.SubstanceName, c.CASNumber, c.Category, c.QuantityKg, c.StorageLocation,
		c.ImporterEntity, c.EndUser, c.EndUse, c.PermitRef, c.ReportedSuspicious, c.FlaggedAt)
	return err
}

func (r *radiationRepo) GetSuspiciousChemicals(ctx context.Context) ([]domain.ChemicalPrecursor, error) {
	query := `SELECT substance_name, cas_number, category, quantity_kg, storage_location, importer_entity,
		end_user, end_use, permit_ref, reported_suspicious, flagged_at
		FROM radiation_chemicals WHERE reported_suspicious = true ORDER BY flagged_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.ChemicalPrecursor
	for rows.Next() {
		var c domain.ChemicalPrecursor
		if err := rows.Scan(&c.SubstanceName, &c.CASNumber, &c.Category, &c.QuantityKg,
			&c.StorageLocation, &c.ImporterEntity, &c.EndUser, &c.EndUse, &c.PermitRef,
			&c.ReportedSuspicious, &c.FlaggedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if result == nil {
		result = []domain.ChemicalPrecursor{}
	}
	return result, rows.Err()
}

func (r *radiationRepo) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM radiation_sources`).Scan(&stats.TotalSources)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM radiation_sources WHERE status IN ('LOST','STOLEN')`).Scan(&stats.LostStolenSources)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM radiation_alerts WHERE responded_by IS NULL`).Scan(&stats.UnrespondedAlerts)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM radiation_chemicals WHERE reported_suspicious = true`).Scan(&stats.SuspiciousChems)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

var _ RadiationRepository = (*radiationRepo)(nil)
