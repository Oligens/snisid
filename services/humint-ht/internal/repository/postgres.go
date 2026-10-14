package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/snisid/humint-ht/internal/domain"
)

type HumintRepository interface {
	CreateSource(s domain.Source) (domain.Source, error)
	UpdateCredibility(code string, rating int, reliability string) (domain.Source, error)
	GetSourceByCode(code string) (domain.Source, error)
	GetReportsBySource(code string) ([]domain.IntelligenceReport, error)
	SubmitReport(r domain.IntelligenceReport) (domain.IntelligenceReport, error)
	LogDebriefing(d domain.DebriefingSession) (domain.DebriefingSession, error)
	GetHighRiskSources() ([]domain.Source, error)
	GetSourceNetwork() ([]domain.Source, []domain.IntelligenceReport, error)
	HealthCheck() error
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func scanSource(row interface {
	Scan(dest ...interface{}) error
}) (domain.Source, error) {
	var s domain.Source
	err := row.Scan(
		&s.CodeName, &s.CredibilityRating, &s.ReliabilityRating, &s.Status,
		&s.HandlingOfficerID, &s.PaymentAmount, &s.PaymentFrequency,
		&s.RiskLevel, &s.Compartment, &s.ReportsCount,
		&s.FirstRecruitedAt, &s.LastContactAt, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *PostgresRepo) CreateSource(s domain.Source) (domain.Source, error) {
	s.Status = domain.SourceStatusActive
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt

	_, err := r.db.Exec(
		`INSERT INTO humint_sources (code_name, credibility_rating, reliability_rating, status, handling_officer_id, payment_amount, payment_frequency, risk_level, compartment, reports_count, first_recruited_at, last_contact_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		s.CodeName, s.CredibilityRating, s.ReliabilityRating, s.Status,
		s.HandlingOfficerID, s.PaymentAmount, s.PaymentFrequency,
		s.RiskLevel, s.Compartment, s.ReportsCount,
		s.FirstRecruitedAt, s.LastContactAt, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return s, fmt.Errorf("insert source: %w", err)
	}
	return s, nil
}

func (r *PostgresRepo) UpdateCredibility(code string, rating int, reliability string) (domain.Source, error) {
	_, err := r.db.Exec(
		`UPDATE humint_sources SET credibility_rating = $1, reliability_rating = $2, updated_at = $3 WHERE code_name = $4`,
		rating, reliability, time.Now().UTC(), code,
	)
	if err != nil {
		return domain.Source{}, fmt.Errorf("update credibility: %w", err)
	}
	return r.GetSourceByCode(code)
}

func (r *PostgresRepo) GetSourceByCode(code string) (domain.Source, error) {
	row := r.db.QueryRow(
		`SELECT code_name, credibility_rating, reliability_rating, status, handling_officer_id, payment_amount, payment_frequency, risk_level, compartment, reports_count, first_recruited_at, last_contact_at, created_at, updated_at
		FROM humint_sources WHERE code_name = $1`, code,
	)
	return scanSource(row)
}

func (r *PostgresRepo) GetReportsBySource(code string) ([]domain.IntelligenceReport, error) {
	rows, err := r.db.Query(
		`SELECT id, source_code, classification, content_hash, threat_actors, sectors_targeted, veracity_score, verified_by, created_at
		FROM humint_reports WHERE source_code = $1 ORDER BY created_at DESC`, code,
	)
	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()

	var reports []domain.IntelligenceReport
	for rows.Next() {
		var rep domain.IntelligenceReport
		err := rows.Scan(
			&rep.ID, &rep.SourceCode, &rep.Classification, &rep.ContentHash,
			pq.Array(&rep.ThreatActors), pq.Array(&rep.SectorsTargeted),
			&rep.VeracityScore, pq.Array(&rep.VerifiedBy), &rep.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}

func (r *PostgresRepo) SubmitReport(rep domain.IntelligenceReport) (domain.IntelligenceReport, error) {
	rep.ID = uuid.New().String()
	rep.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(
		`INSERT INTO humint_reports (id, source_code, classification, content_hash, threat_actors, sectors_targeted, veracity_score, verified_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		rep.ID, rep.SourceCode, rep.Classification, rep.ContentHash,
		pq.Array(rep.ThreatActors), pq.Array(rep.SectorsTargeted),
		rep.VeracityScore, pq.Array(rep.VerifiedBy), rep.CreatedAt,
	)
	if err != nil {
		return rep, fmt.Errorf("insert report: %w", err)
	}

	_, err = r.db.Exec(
		`UPDATE humint_sources SET reports_count = reports_count + 1, updated_at = $1 WHERE code_name = $2`,
		time.Now().UTC(), rep.SourceCode,
	)
	if err != nil {
		return rep, fmt.Errorf("update source count: %w", err)
	}

	return rep, nil
}

func (r *PostgresRepo) LogDebriefing(d domain.DebriefingSession) (domain.DebriefingSession, error) {
	d.ID = uuid.New().String()
	d.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(
		`INSERT INTO humint_debriefings (id, source_code, officer_id, session_date, location_method, topics_covered, next_meeting_planned_at, risk_assessment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.ID, d.SourceCode, d.OfficerID, d.SessionDate, d.LocationMethod,
		pq.Array(d.TopicsCovered), d.NextMeetingPlannedAt, d.RiskAssessment, d.CreatedAt,
	)
	if err != nil {
		return d, fmt.Errorf("insert debriefing: %w", err)
	}

	_, err = r.db.Exec(
		`UPDATE humint_sources SET last_contact_at = $1, updated_at = $1 WHERE code_name = $2`,
		d.SessionDate, d.SourceCode,
	)
	return d, err
}

func (r *PostgresRepo) GetHighRiskSources() ([]domain.Source, error) {
	rows, err := r.db.Query(
		`SELECT code_name, credibility_rating, reliability_rating, status, handling_officer_id, payment_amount, payment_frequency, risk_level, compartment, reports_count, first_recruited_at, last_contact_at, created_at, updated_at
		FROM humint_sources WHERE risk_level IN ('HIGH', 'CRITICAL') AND status = 'ACTIVE' ORDER BY risk_level DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query high risk: %w", err)
	}
	defer rows.Close()

	var sources []domain.Source
	for rows.Next() {
		s, err := scanSource(rows)
		if err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *PostgresRepo) GetSourceNetwork() ([]domain.Source, []domain.IntelligenceReport, error) {
	sources, err := r.getActiveSources()
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.Query(
		`SELECT id, source_code, classification, content_hash, threat_actors, sectors_targeted, veracity_score, verified_by, created_at
		FROM humint_reports ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("query network reports: %w", err)
	}
	defer rows.Close()

	var reports []domain.IntelligenceReport
	for rows.Next() {
		var rep domain.IntelligenceReport
		err := rows.Scan(
			&rep.ID, &rep.SourceCode, &rep.Classification, &rep.ContentHash,
			pq.Array(&rep.ThreatActors), pq.Array(&rep.SectorsTargeted),
			&rep.VeracityScore, pq.Array(&rep.VerifiedBy), &rep.CreatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("scan network report: %w", err)
		}
		reports = append(reports, rep)
	}
	return sources, reports, rows.Err()
}

func (r *PostgresRepo) getActiveSources() ([]domain.Source, error) {
	rows, err := r.db.Query(
		`SELECT code_name, credibility_rating, reliability_rating, status, handling_officer_id, payment_amount, payment_frequency, risk_level, compartment, reports_count, first_recruited_at, last_contact_at, created_at, updated_at
		FROM humint_sources WHERE status = 'ACTIVE' ORDER BY code_name`,
	)
	if err != nil {
		return nil, fmt.Errorf("query active sources: %w", err)
	}
	defer rows.Close()

	var sources []domain.Source
	for rows.Next() {
		s, err := scanSource(rows)
		if err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *PostgresRepo) HealthCheck() error {
	return r.db.Ping()
}
