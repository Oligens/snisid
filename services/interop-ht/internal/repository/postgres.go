package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/interop-ht/internal/domain"
)

type Repository interface {
	CreateAgreement(ctx context.Context, a *domain.DataExchangeAgreement) error
	GetAgreement(ctx context.Context, id uuid.UUID) (*domain.DataExchangeAgreement, error)
	LogExchange(ctx context.Context, l *domain.ExchangeLog) error
	GetExchangeLogs(ctx context.Context, agreementID uuid.UUID) ([]domain.ExchangeLog, error)
}

type postgresRepo struct{ db *sql.DB }

func NewPostgresRepo(db *sql.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) CreateAgreement(ctx context.Context, a *domain.DataExchangeAgreement) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO interop_data_exchange_agreements (agreement_id, provider_agency_id, consumer_agency_id, service_name, allowed_fields, legal_basis, rate_limit_per_min, is_active, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.AgreementID, a.ProviderAgencyID, a.ConsumerAgencyID, a.ServiceName, a.AllowedFields, a.LegalBasis, a.RateLimitPerMin, a.IsActive, time.Now().UTC())
	return err
}
func (r *postgresRepo) GetAgreement(ctx context.Context, id uuid.UUID) (*domain.DataExchangeAgreement, error) {
	a := &domain.DataExchangeAgreement{}
	err := r.db.QueryRowContext(ctx, `SELECT agreement_id, provider_agency_id, consumer_agency_id, service_name, allowed_fields, legal_basis, rate_limit_per_min, is_active, created_at FROM interop_data_exchange_agreements WHERE agreement_id = $1`, id).Scan(&a.AgreementID, &a.ProviderAgencyID, &a.ConsumerAgencyID, &a.ServiceName, &a.AllowedFields, &a.LegalBasis, &a.RateLimitPerMin, &a.IsActive, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("agreement not found")
	}
	return a, nil
}
func (r *postgresRepo) LogExchange(ctx context.Context, l *domain.ExchangeLog) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO interop_exchange_log (log_id, agreement_id, request_hash, response_size_bytes, status_code, duration_ms, exchanged_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		l.LogID, l.AgreementID, l.RequestHash, l.ResponseSizeBytes, l.StatusCode, l.DurationMs, time.Now().UTC())
	return err
}
func (r *postgresRepo) GetExchangeLogs(ctx context.Context, agreementID uuid.UUID) ([]domain.ExchangeLog, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT log_id, agreement_id, request_hash, response_size_bytes, status_code, duration_ms, exchanged_at FROM interop_exchange_log WHERE agreement_id = $1 ORDER BY exchanged_at DESC`, agreementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []domain.ExchangeLog
	for rows.Next() {
		var l domain.ExchangeLog
		rows.Scan(&l.LogID, &l.AgreementID, &l.RequestHash, &l.ResponseSizeBytes, &l.StatusCode, &l.DurationMs, &l.ExchangedAt)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
