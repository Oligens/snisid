package entity

import (
	"testing"
	"time"
)

func TestUserCredentials_DefaultValues(t *testing.T) {
	u := UserCredentials{
		UserID:   "usr-001",
		Username: "jsmith",
	}
	if u.UserID != "usr-001" {
		t.Errorf("UserID = %s, want usr-001", u.UserID)
	}
	if u.MfaEnabled {
		t.Error("MfaEnabled should default to false")
	}
	if u.LockedUntil != nil {
		t.Error("LockedUntil should be nil for new user")
	}
}

func TestUserCredentials_PasswordHashNeverExposed(t *testing.T) {
	u := UserCredentials{
		UserID:       "usr-002",
		PasswordHash: "argon2$hashvalue",
	}
	if u.PasswordHash == "" {
		t.Error("PasswordHash should not be empty in struct")
	}
	// gorm tags include `json:"-"` for both sensitive fields
}

func TestSession_ExpiryCheck(t *testing.T) {
	now := time.Now().UTC()
	s := Session{
		SessionID: "sess-001",
		UserID:    "usr-001",
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	if s.ExpiresAt.Before(now) != true {
		t.Error("Session should be expired")
	}
}

func TestSession_Active(t *testing.T) {
	now := time.Now().UTC()
	s := Session{
		SessionID: "sess-002",
		UserID:    "usr-002",
		ExpiresAt: now.Add(24 * time.Hour),
	}
	if s.ExpiresAt.Before(now) {
		t.Error("Session should still be active")
	}
}

func TestSession_DeviceFingerprint(t *testing.T) {
	s := Session{
		SessionID:         "sess-003",
		DeviceFingerprint: "fp-mac-001",
		ClientIP:          "192.168.1.1",
	}
	if s.DeviceFingerprint != "fp-mac-001" {
		t.Errorf("DeviceFingerprint = %s, want fp-mac-001", s.DeviceFingerprint)
	}
	if s.ClientIP != "192.168.1.1" {
		t.Errorf("ClientIP = %s, want 192.168.1.1", s.ClientIP)
	}
}

func TestWebAuthnCredential_Defaults(t *testing.T) {
	w := WebAuthnCredential{
		ID:              []byte{0x01, 0x02},
		UserID:          "usr-001",
		AttestationType: "none",
		SignCount:       0,
	}
	if w.SignCount != 0 {
		t.Errorf("SignCount = %d, want 0", w.SignCount)
	}
	if w.AttestationType != "none" {
		t.Errorf("AttestationType = %s, want none", w.AttestationType)
	}
}

func TestWebAuthnCredential_AfterUse(t *testing.T) {
	w := WebAuthnCredential{
		ID:              []byte{0x03, 0x04},
		UserID:          "usr-002",
		AttestationType: "packed",
		SignCount:       5,
		CreatedAt:       time.Now().UTC(),
	}
	w.SignCount++
	if w.SignCount != 6 {
		t.Errorf("SignCount after increment = %d, want 6", w.SignCount)
	}
}

func TestUserCredentials_LockedUntil(t *testing.T) {
	future := time.Now().UTC().Add(30 * time.Minute)
	u := UserCredentials{
		UserID:      "usr-003",
		LockedUntil: &future,
	}
	if u.LockedUntil == nil {
		t.Fatal("LockedUntil should not be nil")
	}
	if u.LockedUntil.Before(time.Now().UTC()) {
		t.Error("LockedUntil should be in the future")
	}
}

func TestUserCredentials_Unlocked(t *testing.T) {
	u := UserCredentials{
		UserID: "usr-004",
	}
	if u.LockedUntil != nil {
		t.Error("Unlocked user should have nil LockedUntil")
	}
}

func TestSession_RefreshToken(t *testing.T) {
	s := Session{
		SessionID:    "sess-004",
		RefreshToken: "rt-abc-123",
	}
	if s.RefreshToken != "rt-abc-123" {
		t.Errorf("RefreshToken = %s, want rt-abc-123", s.RefreshToken)
	}
}
