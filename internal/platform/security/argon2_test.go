package security

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestArgon2_HashPasswordReturnsValidHash(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("secure-password-123", DefaultParams)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
}

func TestArgon2_VerifyPasswordSucceeds(t *testing.T) {
	t.Parallel()

	password := "MyStr0ng!Pass"
	hash, err := HashPassword(password, DefaultParams)
	require.NoError(t, err)

	match, err := ComparePasswordAndHash(password, hash)
	require.NoError(t, err)
	assert.True(t, match)
}

func TestArgon2_VerifyPasswordFailsWrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("correct-password", DefaultParams)
	require.NoError(t, err)

	match, err := ComparePasswordAndHash("wrong-password", hash)
	require.NoError(t, err)
	assert.False(t, match)
}

func TestArgon2_DifferentParamsValidHashes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    *Argon2Params
	}{
		{"default params", DefaultParams},
		{"low memory", &Argon2Params{Memory: 32 * 1024, Iterations: 2, Parallelism: 1, SaltLength: 16, KeyLength: 32}},
		{"high params", &Argon2Params{Memory: 128 * 1024, Iterations: 5, Parallelism: 4, SaltLength: 32, KeyLength: 64}},
		{"minimal params", &Argon2Params{Memory: 1024, Iterations: 1, Parallelism: 1, SaltLength: 8, KeyLength: 16}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword("test-password", tt.p)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.True(t, strings.HasPrefix(hash, "$argon2id$"))

			match, err := ComparePasswordAndHash("test-password", hash)
			require.NoError(t, err)
			assert.True(t, match)
		})
	}
}

func TestArgon2_HashFormatContainsExpectedFields(t *testing.T) {
	t.Parallel()

	p := DefaultParams
	hash, err := HashPassword("format-test", p)
	require.NoError(t, err)

	parts := strings.Split(hash, "$")
	require.Len(t, parts, 6)

	assert.Equal(t, "", parts[0])
	assert.Equal(t, "argon2id", parts[1])
	assert.Equal(t, "v="+strconv.Itoa(argon2.Version), parts[2])

	expectedParams := fmt.Sprintf("m=%d,t=%d,p=%d", p.Memory, p.Iterations, p.Parallelism)
	assert.Equal(t, expectedParams, parts[3])

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	require.NoError(t, err)
	assert.Equal(t, int(p.SaltLength), len(salt))

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	require.NoError(t, err)
	assert.Equal(t, int(p.KeyLength), len(decodedHash))
}

func TestArgon2_UniqueSalts(t *testing.T) {
	t.Parallel()

	password := "same-password"
	hash1, err := HashPassword(password, DefaultParams)
	require.NoError(t, err)

	hash2, err := HashPassword(password, DefaultParams)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "hashes should differ due to unique salts")
}

func TestArgon2_InvalidHashFormat(t *testing.T) {
	t.Parallel()

	_, err := ComparePasswordAndHash("password", "not-a-valid-argon2-hash")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHash)
}

func TestArgon2_EmptyPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("", DefaultParams)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	match, err := ComparePasswordAndHash("", hash)
	require.NoError(t, err)
	assert.True(t, match)
}

func TestArgon2_VerifyFailsForDifferentHashFormat(t *testing.T) {
	t.Parallel()

	hash := "$argon2id$v=19$m=65536,t=3,p=2$invalid-salt$invalid-hash"
	match, err := ComparePasswordAndHash("password", hash)
	assert.Error(t, err)
	assert.False(t, match)
}
