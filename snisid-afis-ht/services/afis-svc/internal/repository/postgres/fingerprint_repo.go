package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/snisid/afis-svc/internal/domain"
)

type FingerprintRepository interface {
	Create(ctx context.Context, fp *domain.Fingerprint) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Fingerprint, error)
	GetBySubjectID(ctx context.Context, subjectID uuid.UUID) ([]*domain.Fingerprint, error)
	GetPrimaryBySubjectID(ctx context.Context, subjectID uuid.UUID) (*domain.Fingerprint, error)
	Update(ctx context.Context, fp *domain.Fingerprint) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetPrimary(ctx context.Context, subjectID uuid.UUID, printID uuid.UUID) error
}

type fingerprintRepo struct {
	db *sql.DB
}

func NewFingerprintRepository(db *sql.DB) FingerprintRepository {
	return &fingerprintRepo{db: db}
}

func (r *fingerprintRepo) Create(ctx context.Context, fp *domain.Fingerprint) error {
	query := `
		INSERT INTO afis_fingerprints (
			print_id, subject_id, finger_position, capture_method,
			nfiq2_score, image_ref, minutiae_count, milvus_vector_id,
			template_version, is_primary, captured_at, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`
	_, err := r.db.ExecContext(ctx, query,
		fp.PrintID, fp.SubjectID, fp.FingerPosition, fp.CaptureMethod,
		fp.NFIQ2Score, fp.ImageRef, fp.MinutiaeCount, fp.MilvusVectorID,
		fp.TemplateVersion, fp.IsPrimary, fp.CapturedAt, fp.CreatedBy,
	)
	return err
}

func (r *fingerprintRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Fingerprint, error) {
	query := `
		SELECT print_id, subject_id, finger_position, capture_method,
		       nfiq2_score, quality_accepted, image_ref, minutiae_count,
		       milvus_vector_id, template_version, is_primary, captured_at,
		       created_by, created_at
		FROM afis_fingerprints WHERE print_id = $1
	`
	fp := &domain.Fingerprint{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&fp.PrintID, &fp.SubjectID, &fp.FingerPosition, &fp.CaptureMethod,
		&fp.NFIQ2Score, &fp.QualityAccepted, &fp.ImageRef, &fp.MinutiaeCount,
		&fp.MilvusVectorID, &fp.TemplateVersion, &fp.IsPrimary, &fp.CapturedAt,
		&fp.CreatedBy, &fp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPrintNotFound
	}
	return fp, err
}

func (r *fingerprintRepo) GetBySubjectID(ctx context.Context, subjectID uuid.UUID) ([]*domain.Fingerprint, error) {
	query := `
		SELECT print_id, subject_id, finger_position, capture_method,
		       nfiq2_score, quality_accepted, image_ref, minutiae_count,
		       milvus_vector_id, template_version, is_primary, captured_at,
		       created_by, created_at
		FROM afis_fingerprints WHERE subject_id = $1 ORDER BY finger_position
	`
	rows, err := r.db.QueryContext(ctx, query, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prints []*domain.Fingerprint
	for rows.Next() {
		fp := &domain.Fingerprint{}
		if err := rows.Scan(
			&fp.PrintID, &fp.SubjectID, &fp.FingerPosition, &fp.CaptureMethod,
			&fp.NFIQ2Score, &fp.QualityAccepted, &fp.ImageRef, &fp.MinutiaeCount,
			&fp.MilvusVectorID, &fp.TemplateVersion, &fp.IsPrimary, &fp.CapturedAt,
			&fp.CreatedBy, &fp.CreatedAt,
		); err != nil {
			return nil, err
		}
		prints = append(prints, fp)
	}
	return prints, rows.Err()
}

func (r *fingerprintRepo) GetPrimaryBySubjectID(ctx context.Context, subjectID uuid.UUID) (*domain.Fingerprint, error) {
	query := `
		SELECT print_id, subject_id, finger_position, capture_method,
		       nfiq2_score, quality_accepted, image_ref, minutiae_count,
		       milvus_vector_id, template_version, is_primary, captured_at,
		       created_by, created_at
		FROM afis_fingerprints WHERE subject_id = $1 AND is_primary = TRUE
	`
	fp := &domain.Fingerprint{}
	err := r.db.QueryRowContext(ctx, query, subjectID).Scan(
		&fp.PrintID, &fp.SubjectID, &fp.FingerPosition, &fp.CaptureMethod,
		&fp.NFIQ2Score, &fp.QualityAccepted, &fp.ImageRef, &fp.MinutiaeCount,
		&fp.MilvusVectorID, &fp.TemplateVersion, &fp.IsPrimary, &fp.CapturedAt,
		&fp.CreatedBy, &fp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPrintNotFound
	}
	return fp, err
}

func (r *fingerprintRepo) Update(ctx context.Context, fp *domain.Fingerprint) error {
	query := `
		UPDATE afis_fingerprints SET
			finger_position = $2, capture_method = $3, nfiq2_score = $4,
			image_ref = $5, minutiae_count = $6, milvus_vector_id = $7,
			template_version = $8, is_primary = $9
		WHERE print_id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		fp.PrintID, fp.FingerPosition, fp.CaptureMethod, fp.NFIQ2Score,
		fp.ImageRef, fp.MinutiaeCount, fp.MilvusVectorID,
		fp.TemplateVersion, fp.IsPrimary,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPrintNotFound
	}
	return nil
}

func (r *fingerprintRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM afis_fingerprints WHERE print_id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPrintNotFound
	}
	return nil
}

func (r *fingerprintRepo) SetPrimary(ctx context.Context, subjectID uuid.UUID, printID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE afis_fingerprints SET is_primary = FALSE WHERE subject_id = $1`, subjectID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE afis_fingerprints SET is_primary = TRUE WHERE print_id = $1 AND subject_id = $2`, printID, subjectID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("empreinte primaire non trouvée pour ce sujet")
	}

	return tx.Commit()
}
