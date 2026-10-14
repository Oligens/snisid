package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/sigint-ht/internal/domain"
)

type SigintRepository interface {
	CreateTarget(target domain.InterceptionTarget) (domain.InterceptionTarget, error)
	GetActiveTargets() ([]domain.InterceptionTarget, error)
	GetTargetByID(id string) (domain.InterceptionTarget, error)
	RecordCommunication(comm domain.InterceptedCommunication) (domain.InterceptedCommunication, error)
	GetCommunicationsByTarget(targetID string) ([]domain.InterceptedCommunication, error)
	AnalyzeCDR(phone string) ([]domain.CDRAnalysis, error)
	CreateEmergencyTarget(target domain.InterceptionTarget) (domain.InterceptionTarget, error)
	HealthCheck() error
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func scanTarget(row interface {
	Scan(dest ...interface{}) error
}) (domain.InterceptionTarget, error) {
	var t domain.InterceptionTarget
	err := row.Scan(
		&t.ID, &t.TargetType, &t.Status, &t.AuthorizationRef,
		&t.JudgeName, &t.IssuingCourt, &t.StartDate, &t.EndDate,
		&t.TargetIdentifier, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

func (r *PostgresRepo) CreateTarget(t domain.InterceptionTarget) (domain.InterceptionTarget, error) {
	t.ID = uuid.New().String()
	t.Status = domain.TargetStatusActive
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt

	_, err := r.db.Exec(
		`INSERT INTO sigint_targets (id, target_type, status, authorization_ref, judge_name, issuing_court, start_date, end_date, target_identifier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		t.ID, t.TargetType, t.Status, t.AuthorizationRef, t.JudgeName, t.IssuingCourt,
		t.StartDate, t.EndDate, t.TargetIdentifier, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return t, fmt.Errorf("insert target: %w", err)
	}
	return t, nil
}

func (r *PostgresRepo) GetActiveTargets() ([]domain.InterceptionTarget, error) {
	rows, err := r.db.Query(
		`SELECT id, target_type, status, authorization_ref, judge_name, issuing_court, start_date, end_date, target_identifier, created_at, updated_at
		FROM sigint_targets WHERE status = 'ACTIVE' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query active targets: %w", err)
	}
	defer rows.Close()

	var targets []domain.InterceptionTarget
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, fmt.Errorf("scan target: %w", err)
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

func (r *PostgresRepo) GetTargetByID(id string) (domain.InterceptionTarget, error) {
	row := r.db.QueryRow(
		`SELECT id, target_type, status, authorization_ref, judge_name, issuing_court, start_date, end_date, target_identifier, created_at, updated_at
		FROM sigint_targets WHERE id = $1`, id,
	)
	return scanTarget(row)
}

func (r *PostgresRepo) RecordCommunication(c domain.InterceptedCommunication) (domain.InterceptedCommunication, error) {
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(
		`INSERT INTO sigint_intercepted_comms (id, source_target_id, comm_type, metadata, content_ref, intercepted_at, collector_node, case_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		c.ID, c.SourceTargetID, c.CommType, c.Metadata, c.ContentRef, c.InterceptedAt,
		c.CollectorNode, c.CaseNumber, c.CreatedAt,
	)
	if err != nil {
		return c, fmt.Errorf("insert communication: %w", err)
	}
	return c, nil
}

func scanComm(row interface {
	Scan(dest ...interface{}) error
}) (domain.InterceptedCommunication, error) {
	var c domain.InterceptedCommunication
	err := row.Scan(
		&c.ID, &c.SourceTargetID, &c.CommType, &c.Metadata,
		&c.ContentRef, &c.InterceptedAt, &c.CollectorNode, &c.CaseNumber, &c.CreatedAt,
	)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (r *PostgresRepo) GetCommunicationsByTarget(targetID string) ([]domain.InterceptedCommunication, error) {
	rows, err := r.db.Query(
		`SELECT id, source_target_id, comm_type, metadata, content_ref, intercepted_at, collector_node, case_number, created_at
		FROM sigint_intercepted_comms WHERE source_target_id = $1 ORDER BY intercepted_at DESC`, targetID,
	)
	if err != nil {
		return nil, fmt.Errorf("query comms: %w", err)
	}
	defer rows.Close()

	var comms []domain.InterceptedCommunication
	for rows.Next() {
		c, err := scanComm(rows)
		if err != nil {
			return nil, fmt.Errorf("scan comm: %w", err)
		}
		comms = append(comms, c)
	}
	return comms, rows.Err()
}

func scanCDR(row interface {
	Scan(dest ...interface{}) error
}) (domain.CDRAnalysis, error) {
	var c domain.CDRAnalysis
	err := row.Scan(
		&c.ID, &c.Caller, &c.Callee, &c.Duration, &c.TowerLocation,
		&c.IMSI, &c.IMEI, &c.Timestamp, &c.CreatedAt,
	)
	return c, err
}

func (r *PostgresRepo) AnalyzeCDR(phone string) ([]domain.CDRAnalysis, error) {
	rows, err := r.db.Query(
		`SELECT id, caller, callee, duration, tower_location, imsi, imei, timestamp, created_at
		FROM sigint_cdr_analysis WHERE caller = $1 OR callee = $1 ORDER BY timestamp DESC`, phone,
	)
	if err != nil {
		return nil, fmt.Errorf("query CDR: %w", err)
	}
	defer rows.Close()

	var records []domain.CDRAnalysis
	for rows.Next() {
		c, err := scanCDR(rows)
		if err != nil {
			return nil, fmt.Errorf("scan CDR: %w", err)
		}
		records = append(records, c)
	}
	return records, rows.Err()
}

func (r *PostgresRepo) CreateEmergencyTarget(t domain.InterceptionTarget) (domain.InterceptionTarget, error) {
	return r.CreateTarget(t)
}

func (r *PostgresRepo) HealthCheck() error {
	return r.db.Ping()
}
