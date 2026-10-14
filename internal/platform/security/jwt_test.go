package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTService(ttl time.Duration) *JWTService {
	return &JWTService{
		secret:     "test-secret-key-for-jwt-testing",
		ttl:        ttl,
		refreshTTL: 24 * time.Hour,
	}
}

func TestJWT_GenerateValidToken(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(15 * time.Minute)

	token, err := svc.SignToken("user-123", "admin", "ONI")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)

	assert.Equal(t, "user-123", claims.Subject)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "ONI", claims.Agency)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
}

func TestJWT_ValidateTokenSuccessfully(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(15 * time.Minute)

	token, err := svc.SignToken("subject-1", "viewer", "DGI")
	require.NoError(t, err)

	claims, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "subject-1", claims.Subject)
	assert.Equal(t, "viewer", claims.Role)
	assert.Equal(t, "DGI", claims.Agency)
}

func TestJWT_RejectExpiredToken(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(-1 * time.Hour)

	token, err := svc.SignToken("expired-user", "admin", "POLICE")
	require.NoError(t, err)

	claims, err := svc.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestJWT_RejectWrongSignature(t *testing.T) {
	t.Parallel()

	svc1 := newTestJWTService(15 * time.Minute)
	token, err := svc1.SignToken("user-456", "admin", "ONI")
	require.NoError(t, err)

	svc2 := &JWTService{
		secret:     "different-secret-key",
		ttl:        15 * time.Minute,
		refreshTTL: 24 * time.Hour,
	}

	claims, err := svc2.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWT_ExtractClaims(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(15 * time.Minute)
	token, err := svc.SignToken("extract-user", "analyst", "ANH")
	require.NoError(t, err)

	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(svc.secret), nil
	})
	require.NoError(t, err)

	claims, ok := parsed.Claims.(*Claims)
	require.True(t, ok)
	assert.True(t, parsed.Valid)
	assert.Equal(t, "extract-user", claims.Subject)
	assert.Equal(t, "analyst", claims.Role)
	assert.Equal(t, "ANH", claims.Agency)
}

func TestJWT_KeyRotation(t *testing.T) {
	t.Parallel()

	oldSvc := newTestJWTService(15 * time.Minute)

	token, err := oldSvc.SignToken("rotation-user", "admin", "ONI")
	require.NoError(t, err)

	newSvc := &JWTService{
		secret:     "new-rotated-secret-key-12345",
		ttl:        15 * time.Minute,
		refreshTTL: 24 * time.Hour,
	}

	claims, err := newSvc.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)

	newToken, err := newSvc.SignToken("rotation-user", "admin", "ONI")
	require.NoError(t, err)

	claims, err = newSvc.ParseToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, "rotation-user", claims.Subject)
}

func TestJWT_TokenWithDifferentRoles(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(15 * time.Minute)

	tests := []struct {
		name   string
		role   string
		agency string
	}{
		{"admin ONI", "admin", "ONI"},
		{"viewer DGI", "viewer", "DGI"},
		{"analyst ANH", "analyst", "ANH"},
		{"operator DCPJ", "operator", "DCPJ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.SignToken("user", tt.role, tt.agency)
			require.NoError(t, err)

			claims, err := svc.ParseToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.role, claims.Role)
			assert.Equal(t, tt.agency, claims.Agency)
		})
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(15 * time.Minute)

	claims, err := svc.ParseToken("not-a-valid-jwt-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWT_EmptyToken(t *testing.T) {
	t.Parallel()

	svc := newTestJWTService(15 * time.Minute)

	claims, err := svc.ParseToken("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
