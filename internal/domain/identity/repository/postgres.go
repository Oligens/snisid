package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/snisid/platform/internal/domain/identity/entity"
	"gorm.io/gorm"
)

type IdentityRepository interface {
	Create(ctx context.Context, ident *entity.Identity, reason, changedBy string) error
	Update(ctx context.Context, ident *entity.Identity, reason, changedBy string) error
	GetByID(ctx context.Context, id string) (*entity.Identity, error)
	GetHistory(ctx context.Context, id string) ([]entity.IdentityHistory, error)
	Delete(ctx context.Context, id, reason, changedBy string) error
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) IdentityRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, ident *entity.Identity, reason, changedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ident).Error; err != nil {
			return err
		}

		history := r.buildHistory(ident, reason, changedBy)
		return tx.Create(&history).Error
	})
}

func (r *postgresRepository) Update(ctx context.Context, ident *entity.Identity, reason, changedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ident.Version++
		if err := tx.Save(ident).Error; err != nil {
			return err
		}

		history := r.buildHistory(ident, reason, changedBy)
		return tx.Create(&history).Error
	})
}

func (r *postgresRepository) Delete(ctx context.Context, id, reason, changedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ident entity.Identity
		if err := tx.First(&ident, "id = ?", id).Error; err != nil {
			return err
		}

		ident.Status = entity.StateDeceased // Or soft delete
		ident.Version++
		if err := tx.Save(&ident).Error; err != nil {
			return err
		}

		history := r.buildHistory(&ident, reason, changedBy)
		return tx.Create(&history).Error
	})
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*entity.Identity, error) {
	var ident entity.Identity
	err := r.db.WithContext(ctx).
		Preload("Biometrics").
		Preload("Documents").
		First(&ident, "id = ?", id).Error
	return &ident, err
}

func (r *postgresRepository) GetHistory(ctx context.Context, id string) ([]entity.IdentityHistory, error) {
	var history []entity.IdentityHistory
	err := r.db.WithContext(ctx).Where("identity_id = ?", id).Order("version desc").Find(&history).Error
	return history, err
}

func (r *postgresRepository) buildHistory(ident *entity.Identity, reason, changedBy string) entity.IdentityHistory {
	return entity.IdentityHistory{
		HistoryID:  uuid.NewString(),
		IdentityID: ident.ID,
		FirstName:  ident.FirstName,
		LastName:   ident.LastName,
		DOB:        ident.DOB,
		Gender:     ident.Gender,
		Agency:     ident.Agency,
		Status:     ident.Status,
		Version:    ident.Version,
		ChangedAt:  ident.UpdatedAt,
		ChangedBy:  changedBy,
		Reason:     reason,
	}
}
