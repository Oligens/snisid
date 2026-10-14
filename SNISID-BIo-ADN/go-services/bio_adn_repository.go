package bio_adn

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DNAProfile struct {
	ProfileID         uuid.UUID  `json:"profile_id" db:"profile_id"`
	NIU               *string    `json:"niu,omitempty" db:"niu"`
	ProfileType       string     `json:"profile_type" db:"profile_type"`
	CaseReference     *string    `json:"case_reference,omitempty" db:"case_reference"`
	LabCaseNumber     *string    `json:"lab_case_number,omitempty" db:"lab_case_number"`
	ProfileHash       string     `json:"profile_hash" db:"profile_hash"`
	Status            string     `json:"status" db:"status"`
	SubmittingAgency  string     `json:"submitting_agency" db:"submitting_agency"`
	SubmittingOfficer *string    `json:"submitting_officer,omitempty" db:"submitting_officer"`
	AnalysisDate      *time.Time `json:"analysis_date,omitempty" db:"analysis_date"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type PersonRecord struct {
	RecordID           uuid.UUID `json:"record_id" db:"record_id"`
	NIU                string    `json:"niu" db:"niu"`
	RecordType         string    `json:"record_type" db:"record_type"`
	Status             string    `json:"status" db:"status"`
	PriorityLevel      int16     `json:"priority_level" db:"priority_level"`
	SubjectName        *string   `json:"subject_name,omitempty" db:"subject_name"`
	SubjectAlias       []string  `json:"subject_alias" db:"subject_alias"`
	SubjectDescription *string   `json:"subject_description,omitempty" db:"subject_description"`
	PhotoRefs          []string  `json:"photo_refs" db:"photo_refs"`
	LastKnownLocation  *string   `json:"last_known_location,omitempty" db:"last_known_location"`
	LastKnownDept      *string   `json:"last_known_dept,omitempty" db:"last_known_dept"`
	ReportingAgency    string    `json:"reporting_agency" db:"reporting_agency"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type PropertyRecord struct {
	RecordID        uuid.UUID `json:"record_id" db:"record_id"`
	RecordType      string    `json:"record_type" db:"record_type"`
	Status          string    `json:"status" db:"status"`
	ItemDescription string    `json:"item_description" db:"item_description"`
	SerialNumber    *string   `json:"serial_number,omitempty" db:"serial_number"`
	Make            *string   `json:"make,omitempty" db:"make"`
	Model           *string   `json:"model,omitempty" db:"model"`
	Color           *string   `json:"color,omitempty" db:"color"`
	VIN             *string   `json:"vin,omitempty" db:"vin"`
	PlateNumber     *string   `json:"plate_number,omitempty" db:"plate_number"`
	TheftDate       time.Time `json:"theft_date" db:"theft_date"`
	TheftLocation   *string   `json:"theft_location,omitempty" db:"theft_location"`
	TheftDept       *string   `json:"theft_dept,omitempty" db:"theft_dept"`
	CaseReference   *string   `json:"case_reference,omitempty" db:"case_reference"`
	ReportingAgency string    `json:"reporting_agency" db:"reporting_agency"`
	LinkedPersonNIU *string   `json:"linked_person_niu,omitempty" db:"linked_person_niu"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type DNAMatch struct {
	MatchID            uuid.UUID `json:"match_id" db:"match_id"`
	QueryProfileID     uuid.UUID `json:"query_profile_id" db:"query_profile_id"`
	CandidateProfileID uuid.UUID `json:"candidate_profile_id" db:"candidate_profile_id"`
	MatchType          string    `json:"match_type" db:"match_type"`
	MatchScore         float64   `json:"match_score" db:"match_score"`
	MatchDate          time.Time `json:"match_date" db:"match_date"`
	ReviewStatus       string    `json:"review_status" db:"review_status"`
}

type DNARepository struct {
	db *sql.DB
}

func NewDNARepository(db *sql.DB) *DNARepository {
	return &DNARepository{db: db}
}

func (r *DNARepository) CreateProfile(ctx context.Context, profile *DNAProfile) error {
	query := `
		INSERT INTO snisid_bio_adn.dna_profiles (
			profile_id, niu, profile_type, case_reference, lab_case_number,
			profile_hash, status, submitting_agency, submitting_officer,
			analysis_date, is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`
	_, err := r.db.ExecContext(ctx, query,
		profile.ProfileID, profile.NIU, profile.ProfileType, profile.CaseReference,
		profile.LabCaseNumber, profile.ProfileHash, profile.Status,
		profile.SubmittingAgency, profile.SubmittingOfficer,
		profile.AnalysisDate, profile.IsActive, profile.CreatedAt, profile.UpdatedAt,
	)
	return err
}

func (r *DNARepository) FindByID(ctx context.Context, id uuid.UUID) (*DNAProfile, error) {
	var profile DNAProfile
	query := `SELECT * FROM snisid_bio_adn.dna_profiles WHERE profile_id = $1`
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&profile.ProfileID, &profile.NIU, &profile.ProfileType, &profile.CaseReference,
		&profile.LabCaseNumber, &profile.ProfileHash, &profile.Status,
		&profile.SubmittingAgency, &profile.SubmittingOfficer,
		&profile.AnalysisDate, &profile.IsActive, &profile.CreatedAt, &profile.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *DNARepository) SearchByNIU(ctx context.Context, niu string) ([]*DNAProfile, error) {
	var profiles []*DNAProfile
	query := `SELECT * FROM snisid_bio_adn.dna_profiles WHERE niu = $1 AND is_active = TRUE`
	rows, err := r.db.QueryContext(ctx, query, niu)
	if err != nil {
		return nil, fmt.Errorf("search dna by niu: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p DNAProfile
		if err := rows.Scan(
			&p.ProfileID, &p.NIU, &p.ProfileType, &p.CaseReference,
			&p.LabCaseNumber, &p.ProfileHash, &p.Status,
			&p.SubmittingAgency, &p.SubmittingOfficer,
			&p.AnalysisDate, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, &p)
	}
	return profiles, nil
}

type PersonRepository struct {
	db *sql.DB
}

func NewPersonRepository(db *sql.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

func (r *PersonRepository) Create(ctx context.Context, record *PersonRecord) error {
	query := `
		INSERT INTO snisid_bio_adn.person_records (
			record_id, niu, record_type, status, priority_level,
			subject_name, subject_alias, subject_description,
			photo_refs, last_known_location, last_known_dept,
			reporting_agency, is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`
	_, err := r.db.ExecContext(ctx, query,
		record.RecordID, record.NIU, record.RecordType, record.Status,
		record.PriorityLevel, record.SubjectName, record.SubjectAlias,
		record.SubjectDescription, record.PhotoRefs, record.LastKnownLocation,
		record.LastKnownDept, record.ReportingAgency, record.IsActive,
		record.CreatedAt, record.UpdatedAt,
	)
	return err
}

func (r *PersonRepository) FindActiveByType(ctx context.Context, recordType string) ([]*PersonRecord, error) {
	var records []*PersonRecord
	query := `SELECT * FROM snisid_bio_adn.person_records WHERE record_type = $1 AND is_active = TRUE ORDER BY priority_level DESC`
	rows, err := r.db.QueryContext(ctx, query, recordType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rec PersonRecord
		if err := rows.Scan(
			&rec.RecordID, &rec.NIU, &rec.RecordType, &rec.Status,
			&rec.PriorityLevel, &rec.SubjectName, &rec.SubjectAlias,
			&rec.SubjectDescription, &rec.PhotoRefs, &rec.LastKnownLocation,
			&rec.LastKnownDept, &rec.ReportingAgency, &rec.IsActive,
			&rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, &rec)
	}
	return records, nil
}

type PropertyRepository struct {
	db *sql.DB
}

func NewPropertyRepository(db *sql.DB) *PropertyRepository {
	return &PropertyRepository{db: db}
}

func (r *PropertyRepository) Create(ctx context.Context, record *PropertyRecord) error {
	query := `
		INSERT INTO snisid_bio_adn.property_records (
			record_id, record_type, status, item_description,
			serial_number, make, model, color, vin, plate_number,
			theft_date, theft_location, theft_dept, case_reference,
			reporting_agency, linked_person_niu, is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
	`
	_, err := r.db.ExecContext(ctx, query,
		record.RecordID, record.RecordType, record.Status, record.ItemDescription,
		record.SerialNumber, record.Make, record.Model, record.Color, record.VIN,
		record.PlateNumber, record.TheftDate, record.TheftLocation, record.TheftDept,
		record.CaseReference, record.ReportingAgency, record.LinkedPersonNIU,
		record.IsActive, record.CreatedAt, record.UpdatedAt,
	)
	return err
}

func (r *PropertyRepository) SearchByPlate(ctx context.Context, plateNumber string) ([]*PropertyRecord, error) {
	var records []*PropertyRecord
	query := `SELECT * FROM snisid_bio_adn.property_records WHERE plate_number = $1 AND is_active = TRUE`
	rows, err := r.db.QueryContext(ctx, query, plateNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rec PropertyRecord
		if err := rows.Scan(
			&rec.RecordID, &rec.RecordType, &rec.Status, &rec.ItemDescription,
			&rec.SerialNumber, &rec.Make, &rec.Model, &rec.Color, &rec.VIN,
			&rec.PlateNumber, &rec.TheftDate, &rec.TheftLocation, &rec.TheftDept,
			&rec.CaseReference, &rec.ReportingAgency, &rec.LinkedPersonNIU,
			&rec.IsActive, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, &rec)
	}
	return records, nil
}
