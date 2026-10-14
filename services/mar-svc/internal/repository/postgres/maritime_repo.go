package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/mar-svc/internal/domain"
)

type maritimeRepo struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewMaritimeRepo(pool *pgxpool.Pool, log *zap.Logger) *maritimeRepo {
	return &maritimeRepo{pool: pool, log: log}
}

func (r *maritimeRepo) CreateVessel(v *domain.Vessel) error {
	ctx := context.Background()
	v.VesselID = uuid.New()
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mar_vessels
		 (vessel_id, national_mar_id, vessel_name, imo_number, mmsi, call_sign,
		  vessel_type, flag_country, hull_color, length_m, tonnage_gt, engine_count, horsepower,
		  owner_name, owner_snisid_id, registration_number, registration_port, status,
		  gang_id, interpol_svd_ref, notes, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		v.VesselID, v.NationalMarID, v.VesselName, v.IMONumber, v.MMSI, v.CallSign,
		v.VesselType, v.FlagCountry, v.HullColor, v.LengthM, v.TonnageGT, v.EngineCount, v.Horsepower,
		v.OwnerName, v.OwnerSnisidID, v.RegistrationNumber, v.RegistrationPort, v.Status,
		v.GangID, v.InterpolSVDRef, v.Notes, v.CreatedAt, v.UpdatedAt)
	return err
}

func (r *maritimeRepo) FindVesselByID(id uuid.UUID) (*domain.Vessel, error) {
	ctx := context.Background()
	v := &domain.Vessel{}
	err := r.pool.QueryRow(ctx,
		`SELECT vessel_id, national_mar_id, vessel_name, imo_number, mmsi, call_sign,
		        vessel_type, flag_country, hull_color, length_m, tonnage_gt, engine_count, horsepower,
		        owner_name, owner_snisid_id, registration_number, registration_port, status,
		        gang_id, interpol_svd_ref, notes, created_at, updated_at
		 FROM mar_vessels WHERE vessel_id = $1`, id).Scan(
		&v.VesselID, &v.NationalMarID, &v.VesselName, &v.IMONumber, &v.MMSI, &v.CallSign,
		&v.VesselType, &v.FlagCountry, &v.HullColor, &v.LengthM, &v.TonnageGT, &v.EngineCount, &v.Horsepower,
		&v.OwnerName, &v.OwnerSnisidID, &v.RegistrationNumber, &v.RegistrationPort, &v.Status,
		&v.GangID, &v.InterpolSVDRef, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *maritimeRepo) CreateAISSighting(s *domain.AISSighting) error {
	ctx := context.Background()
	s.SightingID = uuid.New()
	s.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mar_ais_sightings
		 (sighting_id, vessel_id, mmsi, vessel_name, sighting_timestamp, lat, lng,
		  speed_knots, heading_degrees, destination, source_type, zone_code, alert_triggered, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		s.SightingID, s.VesselID, s.MMSI, s.VesselName, s.SightingTimestamp, s.Lat, s.Lng,
		s.SpeedKnots, s.HeadingDegrees, s.Destination, s.SourceType, s.ZoneCode, s.AlertTriggered, s.CreatedAt)
	return err
}

func (r *maritimeRepo) GetLastSighting(mmsi string) (*domain.AISSighting, error) {
	ctx := context.Background()
	s := &domain.AISSighting{}
	err := r.pool.QueryRow(ctx,
		`SELECT sighting_id, vessel_id, mmsi, vessel_name, sighting_timestamp, lat, lng,
		        speed_knots, heading_degrees, destination, source_type, zone_code, alert_triggered, created_at
		 FROM mar_ais_sightings WHERE mmsi = $1
		 ORDER BY sighting_timestamp DESC LIMIT 1`, mmsi).Scan(
		&s.SightingID, &s.VesselID, &s.MMSI, &s.VesselName, &s.SightingTimestamp, &s.Lat, &s.Lng,
		&s.SpeedKnots, &s.HeadingDegrees, &s.Destination, &s.SourceType, &s.ZoneCode, &s.AlertTriggered, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *maritimeRepo) GetLiveAIS(limit int) ([]domain.AISSighting, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT sighting_id, vessel_id, mmsi, vessel_name, sighting_timestamp, lat, lng,
		        speed_knots, heading_degrees, destination, source_type, zone_code, alert_triggered, created_at
		 FROM mar_ais_sightings
		 ORDER BY sighting_timestamp DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSightings(rows)
}

func (r *maritimeRepo) CreateIncident(i *domain.Incident) error {
	ctx := context.Background()
	i.IncidentID = uuid.New()
	i.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mar_incidents
		 (incident_id, vessel_id, incident_type, incident_date, lat, lng, zone_desc,
		  responding_unit, outcome, persons_involved, snisid_person_ids, drug_types,
		  drug_weight_kg, weapons_found, weapons_count, migrants_count, biar_refs,
		  case_reference, photo_refs, created_by, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		i.IncidentID, i.VesselID, i.IncidentType, i.IncidentDate, i.Lat, i.Lng, i.ZoneDesc,
		i.RespondingUnit, i.Outcome, i.PersonsInvolved, i.SnisidPersonIds, i.DrugTypes,
		i.DrugWeightKg, i.WeaponsFound, i.WeaponsCount, i.MigrantsCount, i.BiarRefs,
		i.CaseReference, i.PhotoRefs, i.CreatedBy, i.CreatedAt)
	return err
}

func (r *maritimeRepo) GetRecentIncidents(limit int) ([]domain.Incident, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT incident_id, vessel_id, incident_type, incident_date, lat, lng, zone_desc,
		        responding_unit, outcome, persons_involved, snisid_person_ids, drug_types,
		        drug_weight_kg, weapons_found, weapons_count, migrants_count, biar_refs,
		        case_reference, photo_refs, created_by, created_at
		 FROM mar_incidents
		 ORDER BY incident_date DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIncidents(rows)
}

func (r *maritimeRepo) CreateWatch(w *domain.WatchVessel) error {
	ctx := context.Background()
	w.WatchID = uuid.New()
	w.IsActive = true
	w.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mar_watch_vessels
		 (watch_id, vessel_id, mmsi, vessel_name, watch_reason, alert_level,
		  requesting_unit, is_active, expiry_date, created_by, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		w.WatchID, w.VesselID, w.MMSI, w.VesselName, w.WatchReason, w.AlertLevel,
		w.RequestingUnit, w.IsActive, w.ExpiryDate, w.CreatedBy, w.CreatedAt)
	return err
}

func (r *maritimeRepo) GetActiveWatches() ([]domain.WatchVessel, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT watch_id, vessel_id, mmsi, vessel_name, watch_reason, alert_level,
		        requesting_unit, is_active, expiry_date, created_by, created_at
		 FROM mar_watch_vessels WHERE is_active = true
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWatches(rows)
}

func (r *maritimeRepo) GetIncidentsByZone(zone string, limit int) ([]domain.Incident, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT incident_id, vessel_id, incident_type, incident_date, lat, lng, zone_desc,
		        responding_unit, outcome, persons_involved, snisid_person_ids, drug_types,
		        drug_weight_kg, weapons_found, weapons_count, migrants_count, biar_refs,
		        case_reference, photo_refs, created_by, created_at
		 FROM mar_incidents WHERE zone_desc = $1
		 ORDER BY incident_date DESC LIMIT $2`, zone, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIncidents(rows)
}

func (r *maritimeRepo) GetIncidentStats() (map[string]int64, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT incident_type, COUNT(*) as cnt FROM mar_incidents GROUP BY incident_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := make(map[string]int64)
	for rows.Next() {
		var t string
		var c int64
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		stats[t] = c
	}
	return stats, nil
}

func scanSightings(rows pgx.Rows) ([]domain.AISSighting, error) {
	var sightings []domain.AISSighting
	for rows.Next() {
		var s domain.AISSighting
		if err := rows.Scan(
			&s.SightingID, &s.VesselID, &s.MMSI, &s.VesselName, &s.SightingTimestamp,
			&s.Lat, &s.Lng, &s.SpeedKnots, &s.HeadingDegrees, &s.Destination,
			&s.SourceType, &s.ZoneCode, &s.AlertTriggered, &s.CreatedAt); err != nil {
			return nil, err
		}
		sightings = append(sightings, s)
	}
	return sightings, nil
}

func scanIncidents(rows pgx.Rows) ([]domain.Incident, error) {
	var incidents []domain.Incident
	for rows.Next() {
		var i domain.Incident
		if err := rows.Scan(
			&i.IncidentID, &i.VesselID, &i.IncidentType, &i.IncidentDate,
			&i.Lat, &i.Lng, &i.ZoneDesc, &i.RespondingUnit, &i.Outcome,
			&i.PersonsInvolved, &i.SnisidPersonIds, &i.DrugTypes,
			&i.DrugWeightKg, &i.WeaponsFound, &i.WeaponsCount, &i.MigrantsCount,
			&i.BiarRefs, &i.CaseReference, &i.PhotoRefs, &i.CreatedBy, &i.CreatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	return incidents, nil
}

func scanWatches(rows pgx.Rows) ([]domain.WatchVessel, error) {
	var watches []domain.WatchVessel
	for rows.Next() {
		var w domain.WatchVessel
		if err := rows.Scan(
			&w.WatchID, &w.VesselID, &w.MMSI, &w.VesselName, &w.WatchReason,
			&w.AlertLevel, &w.RequestingUnit, &w.IsActive, &w.ExpiryDate,
			&w.CreatedBy, &w.CreatedAt); err != nil {
			return nil, err
		}
		watches = append(watches, w)
	}
	return watches, nil
}
