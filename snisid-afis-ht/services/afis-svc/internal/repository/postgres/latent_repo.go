package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/snisid/afis-svc/internal/domain"
)

type LatentRepository interface {
	Create(ctx context.Context, lp *domain.LatentPrint) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LatentPrint, error)
	GetByCaseReference(ctx context.Context, caseRef string) ([]*domain.LatentPrint, error)
	GetUnidentified(ctx context.Context, limit, offset int) ([]*domain.LatentPrint, int64, error)
	Update(ctx context.Context, lp *domain.LatentPrint) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetMatch(ctx context.Context, latentID, subjectID uuid.UUID, score float64, examinerID uuid.UUID) error
}

type latentRepo struct {
	db *sql.DB
}

func NewLatentRepository(db *sql.DB) LatentRepository {
	return &latentRepo{db: db}
}

func (r *latentRepo) Create(ctx context.Context, lp *domain.LatentPrint) error {
	query := `
		INSERT INTO afis_latent_prints (
			latent_id, case_reference, crime_scene_id, location_desc,
			dept_code, found_at, image_ref, nfiq2_score,
			finger_position, is_identified, matched_subject_id,
			match_score, examined_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`
	_, err := r.db.ExecContext(ctx, query,
		lp.LatentID, lp.CaseReference, lp.CrimeSceneID, lp.LocationDesc,
		lp.DeptCode, lp.FoundAt, lp.ImageRef, lp.NFIQ2Score,
		lp.FingerPosition, lp.IsIdentified, lp.MatchedSubjectID,
		lp.MatchScore, lp.ExaminedBy,
	)
	return err
}

func (r *latentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LatentPrint, error) {
	query := `
		SELECT latent_id, case_reference, crime_scene_id, location_desc,
		       dept_code, found_at, image_ref, nfiq2_score,
		       finger_position, is_identified, matched_subject_id,
		       match_score, examined_by, created_at
		FROM afis_latent_prints WHERE latent_id = $1
	`
	lp := &domain.LatentPrint{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lp.LatentID, &lp.CaseReference, &lp.CrimeSceneID, &lp.LocationDesc,
		&lp.DeptCode, &lp.FoundAt, &lp.ImageRef, &lp.NFIQ2Score,
		&lp.FingerPosition, &lp.IsIdentified, &lp.MatchedSubjectID,
		&lp.MatchScore, &lp.ExaminedBy, &lp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPrintNotFound
	}
	return lp, err
}

func (r *latentRepo) GetByCaseReference(ctx context.Context, caseRef string) ([]*domain.LatentPrint, error) {
	query := `
		SELECT latent_id, case_reference, crime_scene_id, location_desc,
		       dept_code, found_at, image_ref, nfiq2_score,
		       finger_position, is_identified, matched_subject_id,
		       match_score, examined_by, created_at
		FROM afis_latent_prints WHERE case_reference = $1 ORDER BY found_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, caseRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var latents []*domain.LatentPrint
	for rows.Next() {
		lp := &domain.LatentPrint{}
		if err := rows.Scan(
			&lp.LatentID, &lp.CaseReference, &lp.CrimeSceneID, &lp.LocationDesc,
			&lp.DeptCode, &lp.FoundAt, &lp.ImageRef, &lp.NFIQ2Score,
			&lp.FingerPosition, &lp.IsIdentified, &lp.MatchedSubjectID,
			&lp.MatchScore, &lp.ExaminedBy, &lp.CreatedAt,
		); err != nil {
			return nil, err
		}
		latents = append(latents, lp)
	}
	return latents, rows.Err()
}

func (r *latentRepo) GetUnidentified(ctx context.Context, limit, offset int) ([]*domain.LatentPrint, int64, error) {
	query := `
		SELECT latent_id, case_reference, crime_scene_id, location_desc,
		       dept_code, found_at, image_ref, nfiq2_score,
		       finger_position, is_identified, matched_subject_id,
		       match_score, examined_by, created_at
		FROM afis_latent_prints WHERE is_identified = FALSE
		ORDER BY found_at DESC LIMIT $1 OFFSET $2
	`
	countQuery := `SELECT COUNT(*) FROM afis_latent_prints WHERE is_identified = FALSE`

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var latents []*domain.LatentPrint
	for rows.Next() {
		lp := &domain.LatentPrint{}
		if err := rows.Scan(
			&lp.LatentID, &lp.CaseReference, &lp.CrimeSceneID, &lp.LocationDesc,
			&lp.DeptCode, &lp.FoundAt, &lp.ImageRef, &lp.NFIQ2Score,
			&lp.FingerPosition, &lp.IsIdentified, &lp.MatchedSubjectID,
			&lp.MatchScore, &lp.ExaminedBy, &lp.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		latents = append(latents, lp)
	}
	return latents, total, rows.Err()
}

func (r *latentRepo) Update(ctx context.Context, lp *domain.LatentPrint) error {
	query := `
		UPDATE afis_latent_prints SET
			case_reference = $2, crime_scene_id = $3, location_desc = $4,
			dept_code = $5, found_at = $6, image_ref = $7, nfiq2_score = $8,
			finger_position = $9, is_identified = $10, matched_subject_id = $11,
			match_score = $12, examined_by = $13
		WHERE latent_id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		lp.LatentID, lp.CaseReference, lp.CrimeSceneID, lp.LocationDesc,
		lp.DeptCode, lp.FoundAt, lp.ImageRef, lp.NFIQ2Score,
		lp.FingerPosition, lp.IsIdentified, lp.MatchedSubjectID,
		lp.MatchScore, lp.ExaminedBy,
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

func (r *latentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM afis_latent_prints WHERE latent_id = $1`
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

func (r *latentRepo) SetMatch(ctx context.Context, latentID, subjectID uuid.UUID, score float64, examinerID uuid.UUID) error {
	query := `
		UPDATE afis_latent_prints SET
			is_identified = TRUE, matched_subject_id = $2, match_score = $3,
			examined_by = $4
		WHERE latent_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, latentID, subjectID, score, examinerID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrPrintNotFound
	}
	return nil
}
