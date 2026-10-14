package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/snisid/platform/services/bio-adn/pkg/models"
)

type Postgres struct {
	database *sql.DB
}

func NewPostgres(ctx context.Context, url string) (*Postgres, error) {
	database, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := database.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{database: database}, nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.database.PingContext(ctx)
}

var _ models.Database = (*Postgres)(nil)

func (p *Postgres) CreateDNAProfile(ctx context.Context, profile *models.DNAProfile) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bio_str_profiles (id, specimen_number, index_type, loci_hash, quality_score, loci_count, lab_id, case_number, collected_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		profile.SampleID, profile.SpecimenNumber, profile.IndexType,
		profile.LociHash, profile.QualityScore, profile.LociCount,
		profile.LabID, profile.CaseNumber, profile.CollectedDate,
	)
	return err
}

func (p *Postgres) GetDNAProfileByHash(ctx context.Context, hash string) (*models.DNAProfile, error) {
	row := p.database.QueryRowContext(ctx, `
		SELECT id, specimen_number, index_type, loci_hash, quality_score, loci_count, COALESCE(lab_id,''), COALESCE(case_number,''), collected_date::text, is_expunged
		FROM bio_str_profiles WHERE loci_hash = $1 AND is_expunged = FALSE`, hash)
	prof := &models.DNAProfile{}
	err := row.Scan(&prof.SampleID, &prof.SpecimenNumber, &prof.IndexType, &prof.LociHash,
		&prof.QualityScore, &prof.LociCount, &prof.LabID, &prof.CaseNumber, &prof.CollectedDate, &prof.IsExpunged)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return prof, err
}

func (p *Postgres) GetDNAProfileBySpecimen(ctx context.Context, specimen string) (*models.DNAProfile, error) {
	row := p.database.QueryRowContext(ctx, `
		SELECT id, specimen_number, index_type, loci_hash, quality_score, loci_count, COALESCE(lab_id,''), COALESCE(case_number,''), collected_date::text, is_expunged
		FROM bio_str_profiles WHERE specimen_number = $1 AND is_expunged = FALSE`, specimen)
	prof := &models.DNAProfile{}
	err := row.Scan(&prof.SampleID, &prof.SpecimenNumber, &prof.IndexType, &prof.LociHash,
		&prof.QualityScore, &prof.LociCount, &prof.LabID, &prof.CaseNumber, &prof.CollectedDate, &prof.IsExpunged)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return prof, err
}

func (p *Postgres) SearchDNAProfiles(ctx context.Context, indexType string, limit, offset int) ([]models.DNAProfile, int, error) {
	var total int
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE index_type = $1 AND is_expunged = FALSE`, indexType).Scan(&total)

	rows, err := p.database.QueryContext(ctx, `
		SELECT id, specimen_number, index_type, loci_hash, quality_score, loci_count, COALESCE(lab_id,''), COALESCE(case_number,''), collected_date::text, is_expunged
		FROM bio_str_profiles WHERE index_type = $1 AND is_expunged = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		indexType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []models.DNAProfile
	for rows.Next() {
		var prof models.DNAProfile
		if err := rows.Scan(&prof.SampleID, &prof.SpecimenNumber, &prof.IndexType, &prof.LociHash,
			&prof.QualityScore, &prof.LociCount, &prof.LabID, &prof.CaseNumber, &prof.CollectedDate, &prof.IsExpunged); err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, prof)
	}
	return profiles, total, rows.Err()
}

func (p *Postgres) GetUnuploadedDNAProfiles(ctx context.Context, level string) ([]map[string]any, error) {
	column := "uploaded_ldis"
	if level == "SDIS" {
		column = "uploaded_sdis"
	} else if level != "LDIS" {
		return nil, fmt.Errorf("unknown level: %s", level)
	}
	query := fmt.Sprintf(
		`SELECT id, specimen_number, index_type FROM bio_str_profiles WHERE %s = FALSE AND is_expunged = FALSE LIMIT 100`, column)
	rows, err := p.database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []map[string]any
	for rows.Next() {
		var id, specimen, indexType string
		if err := rows.Scan(&id, &specimen, &indexType); err != nil {
			return nil, err
		}
		profiles = append(profiles, map[string]any{"id": id, "specimen_number": specimen, "index_type": indexType})
	}
	return profiles, rows.Err()
}

func (p *Postgres) MarkUploaded(ctx context.Context, id, level string) error {
	var column string
	switch level {
	case "LDIS":
		column = "uploaded_ldis"
	case "SDIS":
		column = "uploaded_sdis"
	case "NDIS":
		column = "uploaded_ndis"
	default:
		return fmt.Errorf("unknown level: %s", level)
	}
	_, err := p.database.ExecContext(ctx,
		fmt.Sprintf(`UPDATE bio_str_profiles SET %s = TRUE, updated_at = NOW() WHERE id = $1`, column), id)
	return err
}

func (p *Postgres) MarkExpunged(ctx context.Context, id string) error {
	_, err := p.database.ExecContext(ctx,
		`UPDATE bio_str_profiles SET is_expunged = TRUE, expunge_date = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (p *Postgres) CreateWantedPerson(ctx context.Context, wp *models.WantedPerson) error {
	charges, _ := json.Marshal(wp.Charges)
	aliases, _ := json.Marshal(wp.Aliases)
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_wanted_persons (id, record_number, niu, last_name, first_name, aliases, date_of_birth, gender, nationality, warrant_type, warrant_number, issuing_court, issuing_date, charges, danger_level, armed_dangerous, entering_agency, mco_contact, entering_officer, status, expiry_date, interpol_notice, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,NOW(),NOW())`,
		wp.RecordID, wp.RecordNumber, nullIfEmpty(wp.NIU), wp.LastName, wp.FirstName, aliases,
		nullIfEmpty(wp.DateOfBirth), nullIfEmpty(wp.Gender), nullIfEmpty(wp.Nationality),
		wp.WarrantType, wp.WarrantNumber, nullIfEmpty(wp.IssuingCourt), wp.IssuingDate,
		charges, wp.DangerLevel, wp.ArmedDangerous, wp.EnteringAgency, nullIfEmpty(wp.MCOContact),
		nullIfEmpty(wp.EnteringOfficer), wp.Status, nullIfEmpty(wp.ExpiryDate), nullIfEmpty(wp.InterpolNotice))
	return err
}

func (p *Postgres) QueryWantedPersons(ctx context.Context, q *models.WantedQuery) ([]models.WantedPerson, int, error) {
	var total int
	where := "WHERE is_deleted = FALSE"
	var args []any
	argIdx := 1

	if q.LastName != "" {
		where += fmt.Sprintf(" AND last_name ILIKE $%d", argIdx)
		args = append(args, "%"+q.LastName+"%")
		argIdx++
	}
	if q.FirstName != "" {
		where += fmt.Sprintf(" AND first_name ILIKE $%d", argIdx)
		args = append(args, "%"+q.FirstName+"%")
		argIdx++
	}
	if q.NIU != "" {
		where += fmt.Sprintf(" AND niu = $%d", argIdx)
		args = append(args, q.NIU)
		argIdx++
	}
	if q.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, q.Status)
		argIdx++
	}

	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_wanted_persons `+where, args...).Scan(&total)

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	query := fmt.Sprintf(`SELECT id, record_number, COALESCE(niu,''), COALESCE(last_name,''), COALESCE(first_name,''), warrant_type, danger_level, armed_dangerous, entering_agency, COALESCE(mco_contact,''), status, COALESCE(expiry_date::text,''), COALESCE(interpol_notice,'') FROM per_wanted_persons %s ORDER BY created_at DESC LIMIT %d OFFSET %d`, where, limit, q.Offset)

	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var persons []models.WantedPerson
	for rows.Next() {
		var wp models.WantedPerson
		if err := rows.Scan(&wp.RecordID, &wp.RecordNumber, &wp.NIU, &wp.LastName, &wp.FirstName,
			&wp.WarrantType, &wp.DangerLevel, &wp.ArmedDangerous, &wp.EnteringAgency, &wp.MCOContact,
			&wp.Status, &wp.ExpiryDate, &wp.InterpolNotice); err != nil {
			return nil, 0, err
		}
		persons = append(persons, wp)
	}
	return persons, total, rows.Err()
}

func (p *Postgres) GetWantedByID(ctx context.Context, id string) (*models.WantedPerson, error) {
	row := p.database.QueryRowContext(ctx, `
		SELECT id, record_number, COALESCE(niu,''), COALESCE(last_name,''), COALESCE(first_name,''), warrant_type, danger_level, armed_dangerous, entering_agency, COALESCE(mco_contact,''), status, COALESCE(expiry_date::text,''), COALESCE(interpol_notice,'')
		FROM per_wanted_persons WHERE id = $1 AND is_deleted = FALSE`, id)
	wp := &models.WantedPerson{}
	err := row.Scan(&wp.RecordID, &wp.RecordNumber, &wp.NIU, &wp.LastName, &wp.FirstName,
		&wp.WarrantType, &wp.DangerLevel, &wp.ArmedDangerous, &wp.EnteringAgency, &wp.MCOContact,
		&wp.Status, &wp.ExpiryDate, &wp.InterpolNotice)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return wp, err
}

func (p *Postgres) UpdateWantedStatus(ctx context.Context, id, status string) error {
	_, err := p.database.ExecContext(ctx,
		`UPDATE per_wanted_persons SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (p *Postgres) QueryPlateIndex(ctx context.Context, plate string) (*models.PlateHitResult, error) {
	row := p.database.QueryRowContext(ctx, `
		SELECT record_number, 'STOLEN_VEHICLE' as hit_type,
			CASE WHEN status = 'STOLEN' THEN 'HIGH' ELSE 'LOW' END as alert_level,
			COALESCE(entering_agency, 'PNH') as mco_contact
		FROM bie_stolen_vehicles WHERE plate_number = $1 AND status = 'STOLEN' AND is_deleted = FALSE LIMIT 1`, plate)
	hit := &models.PlateHitResult{}
	err := row.Scan(&hit.RecordNumber, &hit.HitType, &hit.AlertLevel, &hit.MCOContact)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	hit.HitFound = true
	return hit, err
}

func (p *Postgres) QueryVINIndex(ctx context.Context, vin string) (*models.PlateHitResult, error) {
	row := p.database.QueryRowContext(ctx, `
		SELECT record_number, 'STOLEN_VEHICLE' as hit_type,
			CASE WHEN status = 'STOLEN' THEN 'HIGH' ELSE 'LOW' END as alert_level,
			COALESCE(entering_agency, 'PNH')
		FROM bie_stolen_vehicles WHERE vin = $1 AND status = 'STOLEN' AND is_deleted = FALSE LIMIT 1`, vin)
	hit := &models.PlateHitResult{}
	err := row.Scan(&hit.RecordNumber, &hit.HitType, &hit.AlertLevel, &hit.MCOContact)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	hit.HitFound = true
	return hit, err
}

func (p *Postgres) QueryPlateClones(ctx context.Context, plate string) (int, error) {
	var count int
	err := p.database.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM bie_stolen_vehicles WHERE plate_number = $1 AND status = 'STOLEN' AND is_deleted = FALSE`, plate).Scan(&count)
	return count, err
}

func (p *Postgres) CreateStolenVehicle(ctx context.Context, v *models.StolenVehicle) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bie_stolen_vehicles (id, record_number, vin, plate_number, vehicle_make, vehicle_model, vehicle_year, vehicle_color, theft_date, theft_location, owner_niu, owner_name, entering_agency, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW())`,
		v.RecordID, v.RecordNumber, nullIfEmpty(v.VIN), v.PlateNumber, v.VehicleMake, v.VehicleModel,
		v.VehicleYear, v.VehicleColor, v.TheftDate, v.TheftLocation, nullIfEmpty(v.OwnerNIU),
		nullIfEmpty(v.OwnerName), v.EnteringAgency, v.Status)
	return err
}

func (p *Postgres) UpdateVehicleStatus(ctx context.Context, id, status, location, agency string) error {
	_, err := p.database.ExecContext(ctx, `
		UPDATE bie_stolen_vehicles SET status = $1, recovered_date = CASE WHEN $1 = 'RECOVERED' THEN CURRENT_DATE ELSE recovered_date END, recovered_location = $2, updated_at = NOW() WHERE id = $3`, status, location, id)
	return err
}

func (p *Postgres) CreateStolenFirearm(ctx context.Context, f *models.StolenFirearm) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bie_stolen_firearms (id, record_number, serial_number, make, model, caliber, firearm_type, barrel_length, theft_date, theft_location, owner_niu, entering_agency, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())`,
		f.RecordID, f.RecordNumber, f.SerialNumber, nullIfEmpty(f.Make), nullIfEmpty(f.Model),
		nullIfEmpty(f.Caliber), nullIfEmpty(f.FirearmType), zeroIfNil(f.BarrelLength), f.TheftDate,
		nullIfEmpty(f.TheftLocation), nullIfEmpty(f.OwnerNIU), f.EnteringAgency, f.Status)
	return err
}

func (p *Postgres) CreateStolenDocument(ctx context.Context, d *models.StolenDocument) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bie_stolen_documents (id, record_number, document_type, document_number, issuing_agency, issue_date, report_date, owner_niu, theft_type, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())`,
		d.RecordID, d.RecordNumber, d.DocumentType, nullIfEmpty(d.DocumentNumber),
		nullIfEmpty(d.IssuingAgency), nullIfEmpty(d.IssueDate), d.ReportDate,
		nullIfEmpty(d.OwnerNIU), d.TheftType, d.Status)
	return err
}

func (p *Postgres) CreateStolenVessel(ctx context.Context, v *models.StolenVessel) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bie_stolen_vessels (id, record_number, vessel_name, registration_number, hull_id_number, vessel_type, vessel_make, vessel_length_m, hull_color, home_port, engine_serial, distinctive_marks, theft_location, theft_date, owner_niu, owner_name, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,NOW())`,
		v.RecordID, v.RecordNumber, nullIfEmpty(v.VesselName), nullIfEmpty(v.RegistrationNumber),
		nullIfEmpty(v.HullIDNumber), nullIfEmpty(v.VesselType), nullIfEmpty(v.VesselMake),
		zeroIfNil(v.VesselLengthM), nullIfEmpty(v.HullColor), nullIfEmpty(v.HomePort),
		nullIfEmpty(v.EngineSerial), nullIfEmpty(v.DistinctiveMarks),
		v.TheftLocation, v.TheftDate, nullIfEmpty(v.OwnerNIU), nullIfEmpty(v.OwnerName), v.Status)
	return err
}

func (p *Postgres) CreateStolenArticle(ctx context.Context, a *models.StolenArticle) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bie_stolen_articles (id, record_number, category, description, serial_number, estimated_value, currency_code, theft_date, theft_location, owner_niu, status, entering_agency, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())`,
		a.RecordID, a.RecordNumber, a.Category, a.Description, nullIfEmpty(a.SerialNumber),
		zeroIfNil(a.EstimatedValue), a.CurrencyCode, a.TheftDate, a.TheftLocation,
		nullIfEmpty(a.OwnerNIU), a.Status, a.EnteringAgency)
	return err
}

func (p *Postgres) CreateStolenSecurity(ctx context.Context, s *models.StolenSecurity) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bie_stolen_securities (id, record_number, security_type, issuer, security_number, face_value, currency_code, issue_date, theft_date, theft_location, owner_niu, status, entering_agency, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())`,
		s.RecordID, s.RecordNumber, s.SecurityType, s.Issuer, s.SecurityNumber,
		zeroIfNil(s.FaceValue), s.CurrencyCode, nullIfEmpty(s.IssueDate), s.TheftDate,
		s.TheftLocation, nullIfEmpty(s.OwnerNIU), s.Status, s.EnteringAgency)
	return err
}

func (p *Postgres) GetUnuploadedProfiles(ctx context.Context, level string) ([]map[string]any, error) {
	return p.GetUnuploadedDNAProfiles(ctx, level)
}

func (p *Postgres) WriteAuditLog(ctx context.Context, event map[string]any) error {
	details, _ := json.Marshal(event)
	officerNIU, _ := event["officer_niu"].(string)
	if officerNIU == "" {
		officerNIU = "system"
	}
	agencyCode, _ := event["agency_code"].(string)
	if agencyCode == "" {
		agencyCode = "BIO-ADN"
	}
	action, _ := event["action"].(string)
	if action == "" {
		action = "SYSTEM"
	}
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bio_audit_log (event_type, table_name, record_id, officer_niu, agency_code, purpose, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		event["event_type"], event["table_name"], event["record_id"],
		officerNIU, agencyCode, event["purpose"], action, string(details))
	return err
}

func (p *Postgres) Close() error {
	return p.database.Close()
}

// ── PER-FUG ─────────────────────────────────────────────────────────────────

func (p *Postgres) CreateForeignFugitive(ctx context.Context, f *models.ForeignFugitive) error {
	charges, _ := json.Marshal(f.Charges)
	aliases, _ := json.Marshal(f.Aliases)
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_foreign_fugitives (id, record_number, interpol_notice_number, notice_type, last_name, first_name, aliases, date_of_birth, gender, nationality, charges, issuing_country, entering_agency, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW())`,
		f.RecordID, f.RecordNumber, f.InterpolNoticeNumber, f.NoticeType,
		f.LastName, nullIfEmpty(f.FirstName), aliases,
		nullIfEmpty(f.DateOfBirth), nullIfEmpty(f.Gender), nullIfEmpty(f.Nationality),
		charges, f.IssuingCountry, f.EnteringAgency, f.Status)
	return err
}

func (p *Postgres) QueryForeignFugitives(ctx context.Context, lastName, nationality, noticeType string, limit, offset int) ([]models.ForeignFugitive, int, error) {
	var total int
	where := "WHERE is_deleted = FALSE"
	var args []any
	argIdx := 1
	if lastName != "" {
		where += fmt.Sprintf(" AND last_name ILIKE $%d", argIdx)
		args = append(args, "%"+lastName+"%")
		argIdx++
	}
	if nationality != "" {
		where += fmt.Sprintf(" AND nationality = $%d", argIdx)
		args = append(args, nationality)
		argIdx++
	}
	if noticeType != "" {
		where += fmt.Sprintf(" AND notice_type = $%d", argIdx)
		args = append(args, noticeType)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_foreign_fugitives `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, interpol_notice_number, notice_type, COALESCE(last_name,''), COALESCE(first_name,''), COALESCE(nationality,''), issuing_country, entering_agency, status FROM per_foreign_fugitives %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ForeignFugitive
	for rows.Next() {
		var x models.ForeignFugitive
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.InterpolNoticeNumber, &x.NoticeType,
			&x.LastName, &x.FirstName, &x.Nationality, &x.IssuingCountry, &x.EnteringAgency, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetForeignFugitiveByID(ctx context.Context, id string) (*models.ForeignFugitive, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, record_number, interpol_notice_number, notice_type, COALESCE(last_name,''), COALESCE(first_name,''), COALESCE(nationality,''), issuing_country, entering_agency, status FROM per_foreign_fugitives WHERE id = $1 AND is_deleted = FALSE`, id)
	var x models.ForeignFugitive
	err := row.Scan(&x.RecordID, &x.RecordNumber, &x.InterpolNoticeNumber, &x.NoticeType, &x.LastName, &x.FirstName, &x.Nationality, &x.IssuingCountry, &x.EnteringAgency, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── PER-NID ─────────────────────────────────────────────────────────────────

func (p *Postgres) CreateUnidentifiedPerson(ctx context.Context, u *models.UnidentifiedPerson) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_unidentified_persons (id, record_number, discovery_date, discovery_location, discovery_department, estimated_age_min, estimated_age_max, gender, estimated_height_cm, estimated_weight_kg, hair_color, eye_color, distinctive_features, clothing_description, dna_sample_ref, fingerprint_ref, dental_records_ref, entering_agency, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,NOW(),NOW())`,
		u.RecordID, u.RecordNumber, u.DiscoveryDate, u.DiscoveryLocation, nullIfEmpty(u.DiscoveryDepartment),
		u.EstimatedAgeMin, u.EstimatedAgeMax, nullIfEmpty(u.Gender), u.EstimatedHeightCM, u.EstimatedWeightKG,
		nullIfEmpty(u.HairColor), nullIfEmpty(u.EyeColor), nullIfEmpty(u.DistinctiveFeatures),
		nullIfEmpty(u.ClothingDescription), nullIfEmpty(u.DNASampleRef), nullIfEmpty(u.FingerprintRef),
		nullIfEmpty(u.DentalRecordsRef), u.EnteringAgency, u.Status)
	return err
}

func (p *Postgres) QueryUnidentifiedPersons(ctx context.Context, dept, gender string, ageMin, ageMax int, limit, offset int) ([]models.UnidentifiedPerson, int, error) {
	var total int
	where := "WHERE is_deleted = FALSE"
	var args []any
	argIdx := 1
	if dept != "" {
		where += fmt.Sprintf(" AND discovery_department = $%d", argIdx)
		args = append(args, dept)
		argIdx++
	}
	if gender != "" {
		where += fmt.Sprintf(" AND gender = $%d", argIdx)
		args = append(args, gender)
		argIdx++
	}
	if ageMin > 0 {
		where += fmt.Sprintf(" AND estimated_age_max >= $%d", argIdx)
		args = append(args, ageMin)
		argIdx++
	}
	if ageMax > 0 {
		where += fmt.Sprintf(" AND estimated_age_min <= $%d", argIdx)
		args = append(args, ageMax)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_unidentified_persons `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, discovery_date, discovery_location, COALESCE(discovery_department,''), COALESCE(gender,''), estimated_age_min, estimated_age_max, COALESCE(entering_agency,''), status FROM per_unidentified_persons %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.UnidentifiedPerson
	for rows.Next() {
		var x models.UnidentifiedPerson
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.DiscoveryDate, &x.DiscoveryLocation, &x.DiscoveryDepartment, &x.Gender, &x.EstimatedAgeMin, &x.EstimatedAgeMax, &x.EnteringAgency, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetUnidentifiedByID(ctx context.Context, id string) (*models.UnidentifiedPerson, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, record_number, discovery_date, discovery_location, COALESCE(discovery_department,''), COALESCE(gender,''), estimated_age_min, estimated_age_max, COALESCE(entering_agency,''), status FROM per_unidentified_persons WHERE id = $1 AND is_deleted = FALSE`, id)
	var x models.UnidentifiedPerson
	err := row.Scan(&x.RecordID, &x.RecordNumber, &x.DiscoveryDate, &x.DiscoveryLocation, &x.DiscoveryDepartment, &x.Gender, &x.EstimatedAgeMin, &x.EstimatedAgeMax, &x.EnteringAgency, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── PER-TER ─────────────────────────────────────────────────────────────────

func (p *Postgres) CreateTerrorismWatch(ctx context.Context, t *models.TerrorismWatch) error {
	groups, _ := json.Marshal(t.Groups)
	aliases, _ := json.Marshal(t.Aliases)
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_terrorism_watch (id, record_number, niu, last_name, first_name, aliases, date_of_birth, nationality, risk_level, threat_type, groups, last_known_location, entering_agency, approved_by_director, approved_by_pg, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NOW(),NOW())`,
		t.RecordID, t.RecordNumber, nullIfEmpty(t.NIU), t.LastName, nullIfEmpty(t.FirstName), aliases,
		nullIfEmpty(t.DateOfBirth), nullIfEmpty(t.Nationality), t.RiskLevel, t.ThreatType, groups,
		nullIfEmpty(t.LastKnownLocation), t.EnteringAgency, t.ApprovedByDirector, t.ApprovedByPG, t.Status)
	return err
}

func (p *Postgres) QueryTerrorismWatches(ctx context.Context, riskLevel, threatType, nationality string, limit, offset int) ([]models.TerrorismWatch, int, error) {
	var total int
	where := "WHERE is_deleted = FALSE"
	var args []any
	argIdx := 1
	if riskLevel != "" {
		where += fmt.Sprintf(" AND risk_level = $%d", argIdx)
		args = append(args, riskLevel)
		argIdx++
	}
	if threatType != "" {
		where += fmt.Sprintf(" AND threat_type ILIKE $%d", argIdx)
		args = append(args, "%"+threatType+"%")
		argIdx++
	}
	if nationality != "" {
		where += fmt.Sprintf(" AND nationality = $%d", argIdx)
		args = append(args, nationality)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_terrorism_watch `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, COALESCE(niu,''), last_name, COALESCE(first_name,''), COALESCE(nationality,''), risk_level, threat_type, entering_agency, approved_by_director, status FROM per_terrorism_watch %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.TerrorismWatch
	for rows.Next() {
		var x models.TerrorismWatch
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.NIU, &x.LastName, &x.FirstName, &x.Nationality, &x.RiskLevel, &x.ThreatType, &x.EnteringAgency, &x.ApprovedByDirector, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetTerrorismWatchByID(ctx context.Context, id string) (*models.TerrorismWatch, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, record_number, COALESCE(niu,''), last_name, COALESCE(first_name,''), COALESCE(nationality,''), risk_level, threat_type, entering_agency, approved_by_director, status FROM per_terrorism_watch WHERE id = $1 AND is_deleted = FALSE`, id)
	var x models.TerrorismWatch
	err := row.Scan(&x.RecordID, &x.RecordNumber, &x.NIU, &x.LastName, &x.FirstName, &x.Nationality, &x.RiskLevel, &x.ThreatType, &x.EnteringAgency, &x.ApprovedByDirector, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── PER-OPR ─────────────────────────────────────────────────────────────────

func (p *Postgres) CreateProtectionOrder(ctx context.Context, po *models.ProtectionOrder) error {
	restrictions, _ := json.Marshal(po.Restrictions)
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_protection_orders (id, record_number, order_type, issuing_court, issuing_judge, beneficiary_niu, beneficiary_name, protected_person, restrained_person, restrictions, issue_date, expiration_date, emergency_contact, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW())`,
		po.RecordID, po.RecordNumber, po.OrderType, po.IssuingCourt, po.IssuingJudge,
		nullIfEmpty(po.BeneficiaryNIU), po.BeneficiaryName, po.BeneficiaryName, po.RestrainedPerson,
		restrictions, po.IssueDate, nullIfEmpty(po.ExpirationDate), nullIfEmpty(po.EmergencyContact), po.Status)
	return err
}

func (p *Postgres) QueryProtectionOrders(ctx context.Context, beneficiaryName, restrainedPerson, orderType string, limit, offset int) ([]models.ProtectionOrder, int, error) {
	var total int
	where := "WHERE is_deleted = FALSE"
	var args []any
	argIdx := 1
	if beneficiaryName != "" {
		where += fmt.Sprintf(" AND (beneficiary_name ILIKE $%d OR protected_person ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+beneficiaryName+"%")
		argIdx++
	}
	if restrainedPerson != "" {
		where += fmt.Sprintf(" AND restrained_person ILIKE $%d", argIdx)
		args = append(args, "%"+restrainedPerson+"%")
		argIdx++
	}
	if orderType != "" {
		where += fmt.Sprintf(" AND order_type = $%d", argIdx)
		args = append(args, orderType)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_protection_orders `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, order_type, issuing_court, COALESCE(beneficiary_name,''), COALESCE(restrained_person,''), COALESCE(issue_date,''), COALESCE(expiration_date,''), status FROM per_protection_orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ProtectionOrder
	for rows.Next() {
		var x models.ProtectionOrder
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.OrderType, &x.IssuingCourt, &x.BeneficiaryName, &x.RestrainedPerson, &x.IssueDate, &x.ExpirationDate, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetActiveProtectionOrdersByBeneficiary(ctx context.Context, beneficiaryNIU string) ([]models.ProtectionOrder, error) {
	rows, err := p.database.QueryContext(ctx, `
		SELECT id, record_number, order_type, issuing_court, COALESCE(beneficiary_name,''), COALESCE(restrained_person,''), COALESCE(issue_date,''), COALESCE(expiration_date,''), status
		FROM per_protection_orders WHERE beneficiary_niu = $1 AND status = 'ACTIVE' AND (expiration_date IS NULL OR expiration_date >= CURRENT_DATE)`, beneficiaryNIU)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.ProtectionOrder
	for rows.Next() {
		var x models.ProtectionOrder
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.OrderType, &x.IssuingCourt, &x.BeneficiaryName, &x.RestrainedPerson, &x.IssueDate, &x.ExpirationDate, &x.Status); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.ProtectionOrder{}
	}
	return items, rows.Err()
}

// ── PER-LIB ─────────────────────────────────────────────────────────────────

func (p *Postgres) CreateSupervisedRelease(ctx context.Context, s *models.SupervisedRelease) error {
	conditions, _ := json.Marshal(s.Conditions)
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_supervised_releases (id, record_number, niu, last_name, first_name, supervision_type, start_date, end_date, conditions, supervising_officer, supervising_agency, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW(),NOW())`,
		s.RecordID, s.RecordNumber, s.NIU, s.LastName, nullIfEmpty(s.FirstName),
		s.SupervisionType, s.StartDate, nullIfEmpty(s.EndDate), conditions,
		s.SupervisingOfficer, s.SupervisingAgency, s.Status)
	return err
}

func (p *Postgres) QuerySupervisedReleases(ctx context.Context, niu, supervisionType, status string, limit, offset int) ([]models.SupervisedRelease, int, error) {
	var total int
	where := "WHERE is_deleted = FALSE"
	var args []any
	argIdx := 1
	if niu != "" {
		where += fmt.Sprintf(" AND niu = $%d", argIdx)
		args = append(args, niu)
		argIdx++
	}
	if supervisionType != "" {
		where += fmt.Sprintf(" AND supervision_type = $%d", argIdx)
		args = append(args, supervisionType)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_supervised_releases `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, niu, last_name, COALESCE(first_name,''), supervision_type, start_date, COALESCE(end_date,''), supervising_officer, supervising_agency, status FROM per_supervised_releases %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.SupervisedRelease
	for rows.Next() {
		var x models.SupervisedRelease
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.NIU, &x.LastName, &x.FirstName, &x.SupervisionType, &x.StartDate, &x.EndDate, &x.SupervisingOfficer, &x.SupervisingAgency, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetSupervisedReleaseByID(ctx context.Context, id string) (*models.SupervisedRelease, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, record_number, niu, last_name, COALESCE(first_name,''), supervision_type, start_date, COALESCE(end_date,''), supervising_officer, supervising_agency, status FROM per_supervised_releases WHERE id = $1 AND is_deleted = FALSE`, id)
	var x models.SupervisedRelease
	err := row.Scan(&x.RecordID, &x.RecordNumber, &x.NIU, &x.LastName, &x.FirstName, &x.SupervisionType, &x.StartDate, &x.EndDate, &x.SupervisingOfficer, &x.SupervisingAgency, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── PER-SEX / PER-GNG ──────────────────────────────────────────────────────

func (p *Postgres) UpdateSexOffenderRisk(ctx context.Context, id, riskLevel, address string) error {
	_, err := p.database.ExecContext(ctx, `UPDATE per_sex_offenders SET risk_level = $1, current_address = COALESCE($2, current_address), updated_at = NOW() WHERE id = $3`, riskLevel, nullIfEmpty(address), id)
	return err
}

func (p *Postgres) RecordGangMemberReview(ctx context.Context, id string) error {
	_, err := p.database.ExecContext(ctx, `UPDATE per_gang_members SET last_review_date = CURRENT_DATE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

// ── Lab Equipment ──────────────────────────────────────────────────────────

func (p *Postgres) CreateLabEquipment(ctx context.Context, e *models.LabEquipment) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bio_lab_equipment (id, lab_code, equipment_name, model, serial_number, role, calibration_date, calibration_due, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`,
		e.ID, e.LabCode, e.EquipmentName, nullIfEmpty(e.Model), e.SerialNumber, e.Role,
		nullIfEmpty(e.CalibrationDate), nullIfEmpty(e.CalibrationDue), e.Status)
	return err
}

func (p *Postgres) QueryLabEquipment(ctx context.Context, labCode string) ([]models.LabEquipment, error) {
	query := `SELECT id, lab_code, equipment_name, COALESCE(model,''), serial_number, role, status FROM bio_lab_equipment WHERE is_deleted = FALSE`
	var args []any
	if labCode != "" {
		query += " AND lab_code = $1"
		args = append(args, labCode)
	}
	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.LabEquipment
	for rows.Next() {
		var x models.LabEquipment
		if err := rows.Scan(&x.ID, &x.LabCode, &x.EquipmentName, &x.Model, &x.SerialNumber, &x.Role, &x.Status); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.LabEquipment{}
	}
	return items, rows.Err()
}

func (p *Postgres) GetLabEquipmentByID(ctx context.Context, id string) (*models.LabEquipment, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, lab_code, equipment_name, COALESCE(model,''), serial_number, role, status FROM bio_lab_equipment WHERE id = $1 AND is_deleted = FALSE`, id)
	var x models.LabEquipment
	err := row.Scan(&x.ID, &x.LabCode, &x.EquipmentName, &x.Model, &x.SerialNumber, &x.Role, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

func (p *Postgres) UpdateEquipmentCalibration(ctx context.Context, id, calibrationDate, calibrationDue, status string) error {
	_, err := p.database.ExecContext(ctx, `UPDATE bio_lab_equipment SET calibration_date = COALESCE($1, calibration_date), calibration_due = COALESCE($2, calibration_due), status = COALESCE($3, status), updated_at = NOW() WHERE id = $4`,
		nullIfEmpty(calibrationDate), nullIfEmpty(calibrationDue), nullIfEmpty(status), id)
	return err
}

// ── Staff Training ──────────────────────────────────────────────────────────

func (p *Postgres) CreateStaffTraining(ctx context.Context, t *models.StaffTraining) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bio_staff_training (id, staff_niu, training_name, training_code, duration_hours, completed_date, valid_until, issued_by, frequency, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`,
		t.ID, t.StaffNIU, t.TrainingName, t.TrainingCode, t.DurationHours, t.CompletedDate,
		nullIfEmpty(t.ValidUntil), t.IssuedBy, nullIfEmpty(t.Frequency))
	return err
}

func (p *Postgres) QueryStaffTraining(ctx context.Context, staffNIU string) ([]models.StaffTraining, error) {
	query := `SELECT id, staff_niu, training_name, training_code, duration_hours, completed_date, COALESCE(valid_until::text,''), issued_by FROM bio_staff_training WHERE is_deleted = FALSE`
	var args []any
	if staffNIU != "" {
		query += " AND staff_niu = $1"
		args = append(args, staffNIU)
	}
	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.StaffTraining
	for rows.Next() {
		var x models.StaffTraining
		if err := rows.Scan(&x.ID, &x.StaffNIU, &x.TrainingName, &x.TrainingCode, &x.DurationHours, &x.CompletedDate, &x.ValidUntil, &x.IssuedBy); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.StaffTraining{}
	}
	return items, rows.Err()
}

func (p *Postgres) GetStaffTrainingByID(ctx context.Context, id string) (*models.StaffTraining, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, staff_niu, training_name, training_code, duration_hours, completed_date, COALESCE(valid_until::text,''), issued_by FROM bio_staff_training WHERE id = $1 AND is_deleted = FALSE`, id)
	var x models.StaffTraining
	err := row.Scan(&x.ID, &x.StaffNIU, &x.TrainingName, &x.TrainingCode, &x.DurationHours, &x.CompletedDate, &x.ValidUntil, &x.IssuedBy)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── Specimen Duplicate Detection ─────────────────────────────────────────────

func (p *Postgres) CheckDuplicateSpecimen(ctx context.Context, specimen string) (bool, error) {
	var count int
	err := p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE specimen_number = $1`, specimen).Scan(&count)
	return count > 0, err
}

func (p *Postgres) MarkSpecimenSubmitted(ctx context.Context, specimen, sampleID string) error {
	return nil
}

// ── NDIS ─────────────────────────────────────────────────────────────────────

func (p *Postgres) RecordCrossDeptHit(ctx context.Context, h *models.NdisCrossDeptHit) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO ndis_cross_dept_hits (id, query_sample_id, match_sample_id, match_type, confidence, query_sdis, match_sdis, alert_level, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
		h.HitID, h.QuerySampleID, h.MatchSampleID, h.MatchType, h.Confidence,
		h.QuerySDIS, h.MatchSDIS, h.AlertLevel)
	return err
}

func (p *Postgres) QueryCrossDeptHits(ctx context.Context, sdis, matchType string, limit, offset int) ([]models.NdisCrossDeptHit, int, error) {
	var total int
	where := "WHERE 1=1"
	var args []any
	argIdx := 1
	if sdis != "" {
		where += fmt.Sprintf(" AND (query_sdis = $%d OR match_sdis = $%d)", argIdx, argIdx)
		args = append(args, sdis)
		argIdx++
	}
	if matchType != "" {
		where += fmt.Sprintf(" AND match_type = $%d", argIdx)
		args = append(args, matchType)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM ndis_cross_dept_hits `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, query_sample_id, match_sample_id, match_type, confidence, query_sdis, match_sdis, alert_level FROM ndis_cross_dept_hits %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.NdisCrossDeptHit
	for rows.Next() {
		var x models.NdisCrossDeptHit
		if err := rows.Scan(&x.HitID, &x.QuerySampleID, &x.MatchSampleID, &x.MatchType, &x.Confidence, &x.QuerySDIS, &x.MatchSDIS, &x.AlertLevel); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetNdisStats(ctx context.Context) (*models.NdisStats, error) {
	stats := &models.NdisStats{}
	var weekHits, interpolCount int
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM ndis_cross_dept_hits WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&weekHits)
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM ndis_interpol_submissions WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&interpolCount)
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE index_type = 'BIO-CON' AND is_expunged = FALSE`).Scan(&stats.TotalBIOCon)
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE index_type = 'BIO-ARR' AND is_expunged = FALSE`).Scan(&stats.TotalBIOArr)
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE index_type = 'BIO-FSC' AND is_expunged = FALSE`).Scan(&stats.TotalBIOFsc)
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE index_type = 'BIO-DIS' AND is_expunged = FALSE`).Scan(&stats.TotalBIODis)
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM bio_str_profiles WHERE index_type = 'BIO-RNI' AND is_expunged = FALSE`).Scan(&stats.TotalBIORni)
	stats.CrossDeptHitsThisWeek = weekHits
	totalProfiles := stats.TotalBIOCon + stats.TotalBIOArr + stats.TotalBIOFsc
	if totalProfiles > 0 {
		stats.HitRatePercent = float64(weekHits) / float64(totalProfiles) * 100
	}
	return stats, nil
}

func (p *Postgres) CreateNdisReport(ctx context.Context, r *models.NdisReport) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO ndis_reports (id, report_type, status, file_path, created_at) VALUES ($1,$2,$3,$4,NOW())`,
		r.ID, r.ReportType, r.Status, nullIfEmpty(r.FilePath))
	return err
}

func (p *Postgres) QueryNdisReports(ctx context.Context) ([]models.NdisReport, error) {
	rows, err := p.database.QueryContext(ctx, `SELECT id, report_type, status, COALESCE(file_path,''), created_at::text FROM ndis_reports ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.NdisReport
	for rows.Next() {
		var x models.NdisReport
		if err := rows.Scan(&x.ID, &x.ReportType, &x.Status, &x.FilePath, &x.GeneratedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.NdisReport{}
	}
	return items, nil
}

func (p *Postgres) CreateInterpolSubmission(ctx context.Context, s *models.InterpolSubmission) error {
	sampleIDs, _ := json.Marshal(s.SampleIDs)
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO ndis_interpol_submissions (id, sample_ids, reason, case_number, status, created_at) VALUES ($1,$2,$3,$4,$5,NOW())`,
		s.ID, sampleIDs, s.Reason, nullIfEmpty(s.CaseNumber), s.Status)
	return err
}

func (p *Postgres) CountInterpolSubmissionsThisWeek(ctx context.Context) (int, error) {
	var count int
	err := p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM ndis_interpol_submissions WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&count)
	return count, err
}

// ── PER-VIO: Known Violence ──────────────────────────────────────────────────

func (p *Postgres) CreateViolenceRecord(ctx context.Context, v *models.ViolenceRecord) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_violence_records (id, record_number, niu, last_name, first_name, incident_type, incident_date, location, victim_niu, victim_name, arresting_agency, court_case_ref, risk_level, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW())`,
		v.RecordID, v.RecordNumber, nullIfEmpty(v.NIU), nullIfEmpty(v.LastName), nullIfEmpty(v.FirstName),
		v.IncidentType, v.IncidentDate, v.Location, nullIfEmpty(v.VictimNIU), nullIfEmpty(v.VictimName),
		v.ArrestingAgency, nullIfEmpty(v.CourtCaseRef), v.RiskLevel, "ACTIVE")
	return err
}

func (p *Postgres) QueryViolenceRecords(ctx context.Context, niu, incidentType, status string, limit, offset int) ([]models.ViolenceRecord, int, error) {
	var total int
	where := "WHERE 1=1"
	var args []any
	argIdx := 1
	if niu != "" {
		where += fmt.Sprintf(" AND niu = $%d", argIdx)
		args = append(args, niu)
		argIdx++
	}
	if incidentType != "" {
		where += fmt.Sprintf(" AND incident_type = $%d", argIdx)
		args = append(args, incidentType)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_violence_records `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, COALESCE(niu,''), COALESCE(last_name,''), COALESCE(first_name,''), incident_type, incident_date, location, COALESCE(victim_niu,''), COALESCE(victim_name,''), arresting_agency, COALESCE(court_case_ref,''), risk_level, status FROM per_violence_records %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ViolenceRecord
	for rows.Next() {
		var x models.ViolenceRecord
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.NIU, &x.LastName, &x.FirstName, &x.IncidentType, &x.IncidentDate, &x.Location, &x.VictimNIU, &x.VictimName, &x.ArrestingAgency, &x.CourtCaseRef, &x.RiskLevel, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.ViolenceRecord{}
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetViolenceRecordByID(ctx context.Context, id string) (*models.ViolenceRecord, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, record_number, COALESCE(niu,''), COALESCE(last_name,''), COALESCE(first_name,''), incident_type, incident_date, location, COALESCE(victim_niu,''), COALESCE(victim_name,''), arresting_agency, COALESCE(court_case_ref,''), risk_level, status FROM per_violence_records WHERE id = $1`, id)
	var x models.ViolenceRecord
	err := row.Scan(&x.RecordID, &x.RecordNumber, &x.NIU, &x.LastName, &x.FirstName, &x.IncidentType, &x.IncidentDate, &x.Location, &x.VictimNIU, &x.VictimName, &x.ArrestingAgency, &x.CourtCaseRef, &x.RiskLevel, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── PER-IDV: Identity Theft ─────────────────────────────────────────────────

func (p *Postgres) CreateIdentityTheft(ctx context.Context, i *models.IdentityTheft) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO per_identity_thefts (id, record_number, victim_niu, victim_name, fraud_type, document_type_used, perpetrator_known, perpetrator_name, report_date, reporting_agency, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW())`,
		i.RecordID, i.RecordNumber, i.VictimNIU, nullIfEmpty(i.VictimName), i.FraudType,
		nullIfEmpty(i.DocumentTypeUsed), i.PerpetratorKnown, nullIfEmpty(i.PerpetratorName),
		i.ReportDate, i.ReportingAgency, "ACTIVE")
	return err
}

func (p *Postgres) QueryIdentityThefts(ctx context.Context, victimNIU, fraudType, status string, limit, offset int) ([]models.IdentityTheft, int, error) {
	var total int
	where := "WHERE 1=1"
	var args []any
	argIdx := 1
	if victimNIU != "" {
		where += fmt.Sprintf(" AND victim_niu = $%d", argIdx)
		args = append(args, victimNIU)
		argIdx++
	}
	if fraudType != "" {
		where += fmt.Sprintf(" AND fraud_type = $%d", argIdx)
		args = append(args, fraudType)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	p.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM per_identity_thefts `+where, args...).Scan(&total)
	query := fmt.Sprintf(`SELECT id, record_number, victim_niu, COALESCE(victim_name,''), fraud_type, COALESCE(document_type_used,''), perpetrator_known, COALESCE(perpetrator_name,''), report_date, reporting_agency, status FROM per_identity_thefts %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args2 := append(args, limit, offset)
	rows, err := p.database.QueryContext(ctx, query, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.IdentityTheft
	for rows.Next() {
		var x models.IdentityTheft
		if err := rows.Scan(&x.RecordID, &x.RecordNumber, &x.VictimNIU, &x.VictimName, &x.FraudType, &x.DocumentTypeUsed, &x.PerpetratorKnown, &x.PerpetratorName, &x.ReportDate, &x.ReportingAgency, &x.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.IdentityTheft{}
	}
	return items, total, rows.Err()
}

func (p *Postgres) GetIdentityTheftByID(ctx context.Context, id string) (*models.IdentityTheft, error) {
	row := p.database.QueryRowContext(ctx, `SELECT id, record_number, victim_niu, COALESCE(victim_name,''), fraud_type, COALESCE(document_type_used,''), perpetrator_known, COALESCE(perpetrator_name,''), report_date, reporting_agency, status FROM per_identity_thefts WHERE id = $1`, id)
	var x models.IdentityTheft
	err := row.Scan(&x.RecordID, &x.RecordNumber, &x.VictimNIU, &x.VictimName, &x.FraudType, &x.DocumentTypeUsed, &x.PerpetratorKnown, &x.PerpetratorName, &x.ReportDate, &x.ReportingAgency, &x.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

// ── BioIdentityLink (dissociation ADN/identité) ─────────────────────────────

func (p *Postgres) CreateIdentityLink(ctx context.Context, l *models.BioIdentityLink) error {
	_, err := p.database.ExecContext(ctx, `
		INSERT INTO bio_identity_links (sample_id, niu, linked_by, court_order, linked_at)
		VALUES ($1,$2,$3,$4,NOW())`,
		l.SampleID, l.NIU, l.LinkedBy, nullIfEmpty(l.CourtOrder))
	return err
}

func (p *Postgres) GetIdentityLinkBySampleID(ctx context.Context, sampleID string) (*models.BioIdentityLink, error) {
	row := p.database.QueryRowContext(ctx, `SELECT sample_id, niu, linked_by, linked_at::text, COALESCE(court_order,'') FROM bio_identity_links WHERE sample_id = $1`, sampleID)
	var x models.BioIdentityLink
	err := row.Scan(&x.SampleID, &x.NIU, &x.LinkedBy, &x.LinkedAt, &x.CourtOrder)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &x, err
}

func (p *Postgres) QueryIdentityLinksByNIU(ctx context.Context, niu string) ([]models.BioIdentityLink, error) {
	rows, err := p.database.QueryContext(ctx, `SELECT sample_id, niu, linked_by, linked_at::text, COALESCE(court_order,'') FROM bio_identity_links WHERE niu = $1 ORDER BY linked_at DESC`, niu)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.BioIdentityLink
	for rows.Next() {
		var x models.BioIdentityLink
		if err := rows.Scan(&x.SampleID, &x.NIU, &x.LinkedBy, &x.LinkedAt, &x.CourtOrder); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	if items == nil {
		items = []models.BioIdentityLink{}
	}
	return items, rows.Err()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func zeroIfNil(f float64) any {
	if f == 0 {
		return nil
	}
	return f
}
