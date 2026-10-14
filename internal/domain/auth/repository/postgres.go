package repository

import (
	"context"

	"github.com/snisid/platform/internal/domain/auth/entity"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *entity.UserCredentials) error
	GetUserByUsername(ctx context.Context, username string) (*entity.UserCredentials, error)
	UpdateUser(ctx context.Context, user *entity.UserCredentials) error

	RegisterWebAuthn(ctx context.Context, cred *entity.WebAuthnCredential) error
	GetWebAuthnCredentials(ctx context.Context, userID string) ([]entity.WebAuthnCredential, error)
}

type postgresAuthRepo struct {
	db *gorm.DB
}

func NewPostgresAuthRepository(db *gorm.DB) AuthRepository {
	return &postgresAuthRepo{db: db}
}

func (r *postgresAuthRepo) CreateUser(ctx context.Context, user *entity.UserCredentials) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *postgresAuthRepo) GetUserByUsername(ctx context.Context, username string) (*entity.UserCredentials, error) {
	var user entity.UserCredentials
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *postgresAuthRepo) UpdateUser(ctx context.Context, user *entity.UserCredentials) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *postgresAuthRepo) RegisterWebAuthn(ctx context.Context, cred *entity.WebAuthnCredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

func (r *postgresAuthRepo) GetWebAuthnCredentials(ctx context.Context, userID string) ([]entity.WebAuthnCredential, error) {
	var creds []entity.WebAuthnCredential
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&creds).Error
	return creds, err
}
