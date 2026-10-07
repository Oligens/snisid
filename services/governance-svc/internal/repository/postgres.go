package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/snisid/governance-svc/internal/domain"
)

type Repository interface {
	CreateLicense(ctx context.Context, l *domain.SoftwareLicense) error
	ListLicenses(ctx context.Context) ([]domain.SoftwareLicense, error)
	GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.SoftwareLicense, error)
	CreatePolicy(ctx context.Context, p *domain.GovernancePolicy) error
	ListPolicies(ctx context.Context) ([]domain.GovernancePolicy, error)
	CreatePolicyRule(ctx context.Context, r *domain.PolicyRule) error
	GetPolicyRules(ctx context.Context, policyID uuid.UUID) ([]domain.PolicyRule, error)
	CreateAudit(ctx context.Context, a *domain.LicenseAudit) error
	GetAudits(ctx context.Context, licenseID uuid.UUID) ([]domain.LicenseAudit, error)
	ListAllLicenses(ctx context.Context) ([]domain.SoftwareLicense, error)
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateLicense(ctx context.Context, l *domain.SoftwareLicense) error {
	query := `INSERT INTO software_licenses (license_id, name, spdx_id, license_type, version, publisher, is_osi_approved, text, registered_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query,
		l.LicenseID, l.Name, l.SPDXID, l.LicenseType, l.Version, l.Publisher,
		l.IsOsiApproved, l.Text, l.RegisteredAt, l.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert license: %w", err)
	}
	return nil
}

func (r *postgresRepo) ListLicenses(ctx context.Context) ([]domain.SoftwareLicense, error) {
	query := `SELECT license_id, name, spdx_id, license_type, version, publisher, is_osi_approved, text, registered_at, updated_at
		FROM software_licenses ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query licenses: %w", err)
	}
	defer rows.Close()

	var licenses []domain.SoftwareLicense
	for rows.Next() {
		var l domain.SoftwareLicense
		if err := rows.Scan(&l.LicenseID, &l.Name, &l.SPDXID, &l.LicenseType, &l.Version,
			&l.Publisher, &l.IsOsiApproved, &l.Text, &l.RegisteredAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		licenses = append(licenses, l)
	}
	return licenses, rows.Err()
}

func (r *postgresRepo) GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.SoftwareLicense, error) {
	query := `SELECT license_id, name, spdx_id, license_type, version, publisher, is_osi_approved, text, registered_at, updated_at
		FROM software_licenses WHERE license_id = $1`
	l := &domain.SoftwareLicense{}
	err := r.db.QueryRowContext(ctx, query, licenseID).Scan(
		&l.LicenseID, &l.Name, &l.SPDXID, &l.LicenseType, &l.Version,
		&l.Publisher, &l.IsOsiApproved, &l.Text, &l.RegisteredAt, &l.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("license not found: %s", licenseID)
		}
		return nil, fmt.Errorf("query license: %w", err)
	}
	return l, nil
}

func (r *postgresRepo) CreatePolicy(ctx context.Context, p *domain.GovernancePolicy) error {
	query := `INSERT INTO governance_policies (policy_id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		p.PolicyID, p.Name, p.Description, p.IsActive, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert policy: %w", err)
	}
	return nil
}

func (r *postgresRepo) ListPolicies(ctx context.Context) ([]domain.GovernancePolicy, error) {
	query := `SELECT policy_id, name, description, is_active, created_at, updated_at
		FROM governance_policies ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}
	defer rows.Close()

	var policies []domain.GovernancePolicy
	for rows.Next() {
		var p domain.GovernancePolicy
		if err := rows.Scan(&p.PolicyID, &p.Name, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (r *postgresRepo) CreatePolicyRule(ctx context.Context, rule *domain.PolicyRule) error {
	query := `INSERT INTO policy_rules (rule_id, policy_id, rule_type, condition, action, priority, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		rule.RuleID, rule.PolicyID, rule.RuleType, rule.Condition, rule.Action, rule.Priority, rule.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert policy rule: %w", err)
	}
	return nil
}

func (r *postgresRepo) GetPolicyRules(ctx context.Context, policyID uuid.UUID) ([]domain.PolicyRule, error) {
	query := `SELECT rule_id, policy_id, rule_type, condition, action, priority, created_at
		FROM policy_rules WHERE policy_id = $1 ORDER BY priority`
	rows, err := r.db.QueryContext(ctx, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("query policy rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.PolicyRule
	for rows.Next() {
		var rule domain.PolicyRule
		if err := rows.Scan(&rule.RuleID, &rule.PolicyID, &rule.RuleType, &rule.Condition,
			&rule.Action, &rule.Priority, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *postgresRepo) CreateAudit(ctx context.Context, a *domain.LicenseAudit) error {
	query := `INSERT INTO license_audits (audit_id, license_id, policy_id, status, findings, audited_at, reviewed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		a.AuditID, a.LicenseID, a.PolicyID, a.Status, a.Findings, a.AuditedAt, a.ReviewedBy,
	)
	if err != nil {
		return fmt.Errorf("insert license audit: %w", err)
	}
	return nil
}

func (r *postgresRepo) GetAudits(ctx context.Context, licenseID uuid.UUID) ([]domain.LicenseAudit, error) {
	query := `SELECT audit_id, license_id, policy_id, status, findings, audited_at, reviewed_by
		FROM license_audits WHERE license_id = $1 ORDER BY audited_at DESC`
	rows, err := r.db.QueryContext(ctx, query, licenseID)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}
	defer rows.Close()

	var audits []domain.LicenseAudit
	for rows.Next() {
		var a domain.LicenseAudit
		if err := rows.Scan(&a.AuditID, &a.LicenseID, &a.PolicyID, &a.Status, &a.Findings,
			&a.AuditedAt, &a.ReviewedBy); err != nil {
			return nil, err
		}
		audits = append(audits, a)
	}
	return audits, rows.Err()
}

func (r *postgresRepo) ListAllLicenses(ctx context.Context) ([]domain.SoftwareLicense, error) {
	return r.ListLicenses(ctx)
}
