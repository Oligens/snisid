package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/iam-ht/internal/domain"
)

type Repository interface {
	UpsertAssurance(ctx context.Context, a *domain.IdentityAssurance) error
	GetAssurance(ctx context.Context, citizenID uuid.UUID) (*domain.IdentityAssurance, error)
	CreateClient(ctx context.Context, c *domain.AgencyClient) error
	GetClient(ctx context.Context, oauthClientID string) (*domain.AgencyClient, error)
	LogAccess(ctx context.Context, l *domain.AccessLog) error
}

type postgresRepo struct{ db *sql.DB }

func NewPostgresRepo(db *sql.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) UpsertAssurance(ctx context.Context, a *domain.IdentityAssurance) error {
	q := `INSERT INTO iam_identity_assurance (assurance_id, citizen_id, keycloak_user_id, assurance_level, biometric_verified_at, mfa_enrolled, last_login_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (keycloak_user_id) DO UPDATE SET assurance_level=$4, biometric_verified_at=$5, mfa_enrolled=$6, last_login_at=$7`
	_, err := r.db.ExecContext(ctx, q, a.AssuranceID, a.CitizenID, a.KeycloakUserID, a.AssuranceLevel, a.BiometricVerifiedAt, a.MFAEnrolled, a.LastLoginAt, time.Now().UTC())
	return err
}

func (r *postgresRepo) GetAssurance(ctx context.Context, citizenID uuid.UUID) (*domain.IdentityAssurance, error) {
	a := &domain.IdentityAssurance{}
	err := r.db.QueryRowContext(ctx, `SELECT assurance_id, citizen_id, keycloak_user_id, assurance_level, biometric_verified_at, mfa_enrolled, last_login_at, created_at FROM iam_identity_assurance WHERE citizen_id = $1`, citizenID).Scan(
		&a.AssuranceID, &a.CitizenID, &a.KeycloakUserID, &a.AssuranceLevel, &a.BiometricVerifiedAt, &a.MFAEnrolled, &a.LastLoginAt, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("assurance not found")
	}
	return a, nil
}

func (r *postgresRepo) CreateClient(ctx context.Context, c *domain.AgencyClient) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO iam_agency_clients (client_id, agency_name, oauth_client_id, allowed_scopes, redirect_uris, required_assurance_level, is_active, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ClientID, c.AgencyName, c.OAuthClientID, c.AllowedScopes, c.RedirectURIs, c.RequiredAssuranceLevel, c.IsActive, time.Now().UTC())
	return err
}

func (r *postgresRepo) GetClient(ctx context.Context, oauthClientID string) (*domain.AgencyClient, error) {
	c := &domain.AgencyClient{}
	err := r.db.QueryRowContext(ctx, `SELECT client_id, agency_name, oauth_client_id, allowed_scopes, redirect_uris, required_assurance_level, is_active, created_at FROM iam_agency_clients WHERE oauth_client_id = $1`, oauthClientID).Scan(
		&c.ClientID, &c.AgencyName, &c.OAuthClientID, &c.AllowedScopes, &c.RedirectURIs, &c.RequiredAssuranceLevel, &c.IsActive, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("client not found")
	}
	return c, nil
}

func (r *postgresRepo) LogAccess(ctx context.Context, l *domain.AccessLog) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO iam_access_log (log_id, citizen_id, client_id, action, ip_hash, accessed_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		l.LogID, l.CitizenID, l.ClientID, l.Action, l.IPHash, time.Now().UTC())
	return err
}
