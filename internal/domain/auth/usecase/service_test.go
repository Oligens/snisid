package usecase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/snisid/platform/internal/domain/auth/entity"
	"github.com/snisid/platform/internal/platform/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthRepo struct {
	createUserFn        func(ctx context.Context, user *entity.UserCredentials) error
	getUserByUsernameFn func(ctx context.Context, username string) (*entity.UserCredentials, error)
	updateUserFn        func(ctx context.Context, user *entity.UserCredentials) error
	registerWebAuthnFn  func(ctx context.Context, cred *entity.WebAuthnCredential) error
	getWebAuthnCredsFn  func(ctx context.Context, userID string) ([]entity.WebAuthnCredential, error)
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, user *entity.UserCredentials) error {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, user)
	}
	return nil
}

func (m *mockAuthRepo) GetUserByUsername(ctx context.Context, username string) (*entity.UserCredentials, error) {
	if m.getUserByUsernameFn != nil {
		return m.getUserByUsernameFn(ctx, username)
	}
	return nil, errors.New("not found")
}

func (m *mockAuthRepo) UpdateUser(ctx context.Context, user *entity.UserCredentials) error {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, user)
	}
	return nil
}

func (m *mockAuthRepo) RegisterWebAuthn(ctx context.Context, cred *entity.WebAuthnCredential) error {
	if m.registerWebAuthnFn != nil {
		return m.registerWebAuthnFn(ctx, cred)
	}
	return nil
}

func (m *mockAuthRepo) GetWebAuthnCredentials(ctx context.Context, userID string) ([]entity.WebAuthnCredential, error) {
	if m.getWebAuthnCredsFn != nil {
		return m.getWebAuthnCredsFn(ctx, userID)
	}
	return nil, nil
}

type mockSessionRepo struct {
	storeSessionFn            func(ctx context.Context, session *entity.Session) error
	getSessionFn              func(ctx context.Context, sessionID string) (*entity.Session, error)
	revokeSessionFamilyFn     func(ctx context.Context, userID string) error
	incrementFailedAttemptsFn func(ctx context.Context, ip, username string) (int, error)
	resetFailedAttemptsFn     func(ctx context.Context, ip, username string) error
}

func (m *mockSessionRepo) StoreSession(ctx context.Context, session *entity.Session) error {
	if m.storeSessionFn != nil {
		return m.storeSessionFn(ctx, session)
	}
	return nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	if m.getSessionFn != nil {
		return m.getSessionFn(ctx, sessionID)
	}
	return nil, errors.New("session not found")
}

func (m *mockSessionRepo) RevokeSessionFamily(ctx context.Context, userID string) error {
	if m.revokeSessionFamilyFn != nil {
		return m.revokeSessionFamilyFn(ctx, userID)
	}
	return nil
}

func (m *mockSessionRepo) IncrementFailedAttempts(ctx context.Context, ip, username string) (int, error) {
	if m.incrementFailedAttemptsFn != nil {
		return m.incrementFailedAttemptsFn(ctx, ip, username)
	}
	return 0, nil
}

func (m *mockSessionRepo) ResetFailedAttempts(ctx context.Context, ip, username string) error {
	if m.resetFailedAttemptsFn != nil {
		return m.resetFailedAttemptsFn(ctx, ip, username)
	}
	return nil
}

func newTestAuthService(dbRepo *mockAuthRepo, sessionRepo *mockSessionRepo) AuthService {
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	return NewAuthService(dbRepo, sessionRepo, pk)
}

func TestRegister_Success(t *testing.T) {
	t.Parallel()
	var savedUser *entity.UserCredentials

	dbRepo := &mockAuthRepo{
		createUserFn: func(ctx context.Context, user *entity.UserCredentials) error {
			savedUser = user
			return nil
		},
	}
	sessionRepo := &mockSessionRepo{}
	svc := newTestAuthService(dbRepo, sessionRepo)

	err := svc.Register(context.Background(), "jdoe", "StrongP@ss1", "user")
	require.NoError(t, err)
	require.NotNil(t, savedUser)
	assert.NotEmpty(t, savedUser.UserID)
	assert.Equal(t, "jdoe", savedUser.Username)
	assert.NotEmpty(t, savedUser.PasswordHash)
	assert.Contains(t, savedUser.PasswordHash, "$argon2id$")
	assert.Equal(t, "user", savedUser.Roles)
	assert.False(t, savedUser.CreatedAt.IsZero())
}

func TestRegister_DuplicateEmail(t *testing.T) {
	t.Parallel()
	dbRepo := &mockAuthRepo{
		createUserFn: func(ctx context.Context, user *entity.UserCredentials) error {
			return errors.New("duplicate key value violates unique constraint")
		},
	}
	sessionRepo := &mockSessionRepo{}
	svc := newTestAuthService(dbRepo, sessionRepo)

	err := svc.Register(context.Background(), "existing", "StrongP@ss1", "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
}

func TestRegister_WeakPassword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		password string
	}{
		{name: "too short", password: "Ab1"},
		{name: "no digit", password: "abcdefgh"},
		{name: "no uppercase", password: "abcdef1gh"},
		{name: "empty", password: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var savedUser *entity.UserCredentials
			dbRepo := &mockAuthRepo{
				createUserFn: func(ctx context.Context, user *entity.UserCredentials) error {
					savedUser = user
					return nil
				},
			}
			sessionRepo := &mockSessionRepo{}
			svc := newTestAuthService(dbRepo, sessionRepo)

			err := svc.Register(context.Background(), "testuser", tt.password, "user")
			require.NoError(t, err, "Register accepts all passwords and hashes them")
			require.NotNil(t, savedUser)
			assert.NotEmpty(t, savedUser.PasswordHash)
			assert.Contains(t, savedUser.PasswordHash, "$argon2id$")
		})
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()
	password := "StrongP@ss1"
	hash, err := security.HashPassword(password, security.DefaultParams)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	var storedSession *entity.Session

	dbRepo := &mockAuthRepo{
		getUserByUsernameFn: func(ctx context.Context, username string) (*entity.UserCredentials, error) {
			return &entity.UserCredentials{
				UserID:       "u1",
				Username:     "jdoe",
				PasswordHash: hash,
				Roles:        "user",
			}, nil
		},
	}
	sessionRepo := &mockSessionRepo{
		incrementFailedAttemptsFn: func(ctx context.Context, ip, username string) (int, error) {
			return 0, nil
		},
		resetFailedAttemptsFn: func(ctx context.Context, ip, username string) error {
			return nil
		},
		storeSessionFn: func(ctx context.Context, session *entity.Session) error {
			storedSession = session
			return nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	pair, err := svc.Login(context.Background(), "jdoe", password, "192.168.1.1", "chrome-120")
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEmpty(t, pair.SessionID)
	require.NotNil(t, storedSession)
	assert.Equal(t, "u1", storedSession.UserID)
}

func TestLogin_WrongPassword(t *testing.T) {
	t.Parallel()
	hash, err := security.HashPassword("RealPassword1", security.DefaultParams)
	require.NoError(t, err)

	dbRepo := &mockAuthRepo{
		getUserByUsernameFn: func(ctx context.Context, username string) (*entity.UserCredentials, error) {
			return &entity.UserCredentials{
				UserID:       "u1",
				Username:     "jdoe",
				PasswordHash: hash,
				Roles:        "user",
			}, nil
		},
	}
	sessionRepo := &mockSessionRepo{
		incrementFailedAttemptsFn: func(ctx context.Context, ip, username string) (int, error) {
			return 1, nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	_, err = svc.Login(context.Background(), "jdoe", "WrongPassword1", "10.0.0.1", "firefox-110")
	require.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestLogin_AccountLockedAfterFiveAttempts(t *testing.T) {
	t.Parallel()
	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		incrementFailedAttemptsFn: func(ctx context.Context, ip, username string) (int, error) {
			return 6, nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	_, err := svc.Login(context.Background(), "lockeduser", "AnyP@ss1", "10.0.0.2", "safari-16")
	require.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)
}

func TestLogin_LockedUntil(t *testing.T) {
	t.Parallel()
	future := time.Now().Add(30 * time.Minute)
	dbRepo := &mockAuthRepo{
		getUserByUsernameFn: func(ctx context.Context, username string) (*entity.UserCredentials, error) {
			return &entity.UserCredentials{
				UserID:       "u3",
				Username:     "temp-locked",
				PasswordHash: "invalid-hash",
				Roles:        "user",
				LockedUntil:  &future,
			}, nil
		},
	}
	sessionRepo := &mockSessionRepo{
		incrementFailedAttemptsFn: func(ctx context.Context, ip, username string) (int, error) {
			return 0, nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	_, err := svc.Login(context.Background(), "temp-locked", "AnyP@ss1", "10.0.0.3", "edge-110")
	require.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)
}

func TestRefreshToken_Valid(t *testing.T) {
	t.Parallel()
	session := &entity.Session{
		SessionID:         "sess-1",
		UserID:            "u1",
		RefreshToken:      "refresh-token-abc",
		DeviceFingerprint: "device-fp-123",
		ClientIP:          "10.0.0.1",
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour),
	}

	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		getSessionFn: func(ctx context.Context, sessionID string) (*entity.Session, error) {
			return session, nil
		},
		storeSessionFn: func(ctx context.Context, s *entity.Session) error {
			return nil
		},
	}
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	svc := NewAuthService(dbRepo, sessionRepo, pk)

	pair, err := svc.Refresh(context.Background(), "sess-1", "refresh-token-abc", "10.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEmpty(t, pair.SessionID)
	assert.NotEqual(t, "sess-1", pair.SessionID)
}

func TestRefreshToken_Expired(t *testing.T) {
	t.Parallel()
	expired := time.Now().Add(-1 * time.Hour)
	session := &entity.Session{
		SessionID:    "sess-expired",
		UserID:       "u1",
		RefreshToken: "old-refresh",
		ExpiresAt:    expired,
	}

	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		getSessionFn: func(ctx context.Context, sessionID string) (*entity.Session, error) {
			return session, nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	_, err := svc.Refresh(context.Background(), "sess-expired", "old-refresh", "10.0.0.1")
	require.Error(t, err)
	assert.Equal(t, ErrInvalidSession, err)
}

func TestRefreshToken_WrongToken(t *testing.T) {
	t.Parallel()
	session := &entity.Session{
		SessionID:    "sess-2",
		UserID:       "u1",
		RefreshToken: "original-refresh-token",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	var revoked bool

	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		getSessionFn: func(ctx context.Context, sessionID string) (*entity.Session, error) {
			return session, nil
		},
		revokeSessionFamilyFn: func(ctx context.Context, userID string) error {
			revoked = true
			return nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	_, err := svc.Refresh(context.Background(), "sess-2", "wrong-token", "10.0.0.1")
	require.Error(t, err)
	assert.Equal(t, ErrInvalidSession, err)
	assert.True(t, revoked, "session family should be revoked on token theft detection")
}

func TestLogout_ValidSession(t *testing.T) {
	t.Parallel()
	var revokedUserID string

	session := &entity.Session{
		SessionID: "sess-3",
		UserID:    "u1",
	}
	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		getSessionFn: func(ctx context.Context, sessionID string) (*entity.Session, error) {
			return session, nil
		},
		revokeSessionFamilyFn: func(ctx context.Context, userID string) error {
			revokedUserID = userID
			return nil
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	err := svc.Logout(context.Background(), "sess-3")
	require.NoError(t, err)
	assert.Equal(t, "u1", revokedUserID)
}

func TestLogout_SessionNotFound(t *testing.T) {
	t.Parallel()
	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		getSessionFn: func(ctx context.Context, sessionID string) (*entity.Session, error) {
			return nil, errors.New("session not found")
		},
	}
	svc := newTestAuthService(dbRepo, sessionRepo)

	err := svc.Logout(context.Background(), "sess-nonexistent")
	assert.NoError(t, err)
}

func TestNewAuthService(t *testing.T) {
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	svc := NewAuthService(nil, nil, pk)
	assert.NotNil(t, svc)
}

func TestIssueTokens(t *testing.T) {
	t.Parallel()
	var storedSession *entity.Session

	dbRepo := &mockAuthRepo{}
	sessionRepo := &mockSessionRepo{
		storeSessionFn: func(ctx context.Context, session *entity.Session) error {
			storedSession = session
			return nil
		},
	}
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	svc := NewAuthService(dbRepo, sessionRepo, pk)

	user := &entity.UserCredentials{
		UserID: "u1",
		Roles:  "admin,auditor",
	}

	pair, err := svc.(*authService).issueTokens(context.Background(), user, "10.0.0.1", "chrome-120")
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEmpty(t, pair.SessionID)
	require.NotNil(t, storedSession)
	assert.Equal(t, "u1", storedSession.UserID)
}
