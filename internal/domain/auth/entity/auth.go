package entity

import "time"

type UserCredentials struct {
	UserID       string     `json:"userId" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"uniqueIndex"`
	PasswordHash string     `json:"-"` // Never expose in JSON
	MfaEnabled   bool       `json:"mfaEnabled"`
	MfaSecret    string     `json:"-"`
	Roles        string     `json:"roles"` // comma separated roles for simplicity in DB
	LockedUntil  *time.Time `json:"lockedUntil"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type Session struct {
	SessionID         string    `json:"sessionId"`
	UserID            string    `json:"userId"`
	RefreshToken      string    `json:"refreshToken"`
	DeviceFingerprint string    `json:"deviceFingerprint"`
	ClientIP          string    `json:"clientIp"`
	ExpiresAt         time.Time `json:"expiresAt"`
}

type WebAuthnCredential struct {
	ID              []byte    `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"userId"`
	PublicKey       []byte    `json:"publicKey"`
	AttestationType string    `json:"attestationType"`
	SignCount       uint32    `json:"signCount"`
	CreatedAt       time.Time `json:"createdAt"`
}
