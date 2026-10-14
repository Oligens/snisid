package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/snisid/afis-svc/internal/domain"
)

type SubjectRepository interface {
	Create(ctx context.Context, s *domain.Subject) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subject, error)
	GetBySNISIDPersonID(ctx context.Context, snisidPersonID uuid.UUID) (*domain.Subject, error)
	GetByFIRRecordID(ctx context.Context, firRecordID uuid.UUID) (*domain.Subject, error)
	GetByNationalAFISID(ctx context.Context, nationalAFISID string) (*domain.Subject, error)
	Update(ctx context.Context, s *domain.Subject) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, subjectType *domain.SubjectType, limit, offset int) ([]*domain.Subject, int64, error)
}

type subjectRepo struct {
	db *sql.DB
}

func NewSubjectRepository(db *sql.DB) SubjectRepository {
	return &subjectRepo{db: db}
}

func (r *subjectRepo) Create(ctx context.Context, s *domain.Subject) error {
	query := `
		INSERT INTO afis_subjects (
			subject_id, snisid_person_id, fir_record_id, subject_type,
			national_afis_id, alias_ids, enrolment_date, enrolling_unit,
			enrolling_officer
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`
	_, err := r.db.ExecContext(ctx, query,
		s.SubjectID, s.SNISIDPersonID, s.FIRRecordID, s.SubjectType,
		s.NationalAFISID, s.AliasIDs, s.EnrolmentDate, s.EnrollingUnit,
		s.EnrollingOfficer,
	)
	return err
}

func (r *subjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subject, error) {
	query := `
		SELECT subject_id, snisid_person_id, fir_record_id, subject_type,
		       national_afis_id, alias_ids, enrolment_date, enrolling_unit,
		       enrolling_officer, created_at, updated_at
		FROM afis_subjects WHERE subject_id = $1
	`
	s := &domain.Subject{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.SubjectID, &s.SNISIDPersonID, &s.FIRRecordID, &s.SubjectType,
		&s.NationalAFISID, &s.AliasIDs, &s.EnrolmentDate, &s.EnrollingUnit,
		&s.EnrollingOfficer, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSubjectNotFound
	}
	return s, err
}

func (r *subjectRepo) GetBySNISIDPersonID(ctx context.Context, snisidPersonID uuid.UUID) (*domain.Subject, error) {
	query := `
		SELECT subject_id, snisid_person_id, fir_record_id, subject_type,
		       national_afis_id, alias_ids, enrolment_date, enrolling_unit,
		       enrolling_officer, created_at, updated_at
		FROM afis_subjects WHERE snisid_person_id = $1
	`
	s := &domain.Subject{}
	err := r.db.QueryRowContext(ctx, query, snisidPersonID).Scan(
		&s.SubjectID, &s.SNISIDPersonID, &s.FIRRecordID, &s.SubjectType,
		&s.NationalAFISID, &s.AliasIDs, &s.EnrolmentDate, &s.EnrollingUnit,
		&s.EnrollingOfficer, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSubjectNotFound
	}
	return s, err
}

func (r *subjectRepo) GetByFIRRecordID(ctx context.Context, firRecordID uuid.UUID) (*domain.Subject, error) {
	query := `
		SELECT subject_id, snisid_person_id, fir_record_id, subject_type,
		       national_afis_id, alias_ids, enrolment_date, enrolling_unit,
		       enrolling_officer, created_at, updated_at
		FROM afis_subjects WHERE fir_record_id = $1
	`
	s := &domain.Subject{}
	err := r.db.QueryRowContext(ctx, query, firRecordID).Scan(
		&s.SubjectID, &s.SNISIDPersonID, &s.FIRRecordID, &s.SubjectType,
		&s.NationalAFISID, &s.AliasIDs, &s.EnrolmentDate, &s.EnrollingUnit,
		&s.EnrollingOfficer, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSubjectNotFound
	}
	return s, err
}

func (r *subjectRepo) GetByNationalAFISID(ctx context.Context, nationalAFISID string) (*domain.Subject, error) {
	query := `
		SELECT subject_id, snisid_person_id, fir_record_id, subject_type,
		       national_afis_id, alias_ids, enrolment_date, enrolling_unit,
		       enrolling_officer, created_at, updated_at
		FROM afis_subjects WHERE national_afis_id = $1
	`
	s := &domain.Subject{}
	err := r.db.QueryRowContext(ctx, query, nationalAFISID).Scan(
		&s.SubjectID, &s.SNISIDPersonID, &s.FIRRecordID, &s.SubjectType,
		&s.NationalAFISID, &s.AliasIDs, &s.EnrolmentDate, &s.EnrollingUnit,
		&s.EnrollingOfficer, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSubjectNotFound
	}
	return s, err
}

func (r *subjectRepo) Update(ctx context.Context, s *domain.Subject) error {
	query := `
		UPDATE afis_subjects SET
			snisid_person_id = $2, fir_record_id = $3, subject_type = $4,
			national_afis_id = $5, alias_ids = $6, enrolling_unit = $7,
			enrolling_officer = $8, updated_at = NOW()
		WHERE subject_id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		s.SubjectID, s.SNISIDPersonID, s.FIRRecordID, s.SubjectType,
		s.NationalAFISID, s.AliasIDs, s.EnrollingUnit, s.EnrollingOfficer,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSubjectNotFound
	}
	return nil
}

func (r *subjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM afis_subjects WHERE subject_id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSubjectNotFound
	}
	return nil
}

func (r *subjectRepo) List(ctx context.Context, subjectType *domain.SubjectType, limit, offset int) ([]*domain.Subject, int64, error) {
	var args []interface{}
	query := `SELECT subject_id, snisid_person_id, fir_record_id, subject_type,
		national_afis_id, alias_ids, enrolment_date, enrolling_unit,
		enrolling_officer, created_at, updated_at
		FROM afis_subjects`
	countQuery := `SELECT COUNT(*) FROM afis_subjects`

	if subjectType != nil {
		query += ` WHERE subject_type = $1`
		countQuery += ` WHERE subject_type = $1`
		args = append(args, *subjectType)
	}

	query += ` ORDER BY enrolment_date DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subjects []*domain.Subject
	for rows.Next() {
		s := &domain.Subject{}
		if err := rows.Scan(
			&s.SubjectID, &s.SNISIDPersonID, &s.FIRRecordID, &s.SubjectType,
			&s.NationalAFISID, &s.AliasIDs, &s.EnrolmentDate, &s.EnrollingUnit,
			&s.EnrollingOfficer, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		subjects = append(subjects, s)
	}
	return subjects, total, rows.Err()
}
