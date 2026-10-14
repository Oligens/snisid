package domain

import (
	"time"

	"github.com/google/uuid"
)

type IARMSyncLog struct {
	SyncID       uuid.UUID  `json:"sync_id"`
	WeaponID     *uuid.UUID `json:"weapon_id,omitempty"`
	Direction    string     `json:"direction"`
	IARMSRef     *string    `json:"iarms_ref,omitempty"`
	SyncStatus   string     `json:"sync_status"`
	SyncedAt     *time.Time `json:"synced_at,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type SyncResult struct {
	EntriesProcessed int       `json:"entries_processed"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
	Error            *string   `json:"error,omitempty"`
}

type IARMSRecord struct {
	NCBRef          string `json:"ncb_ref"`
	OriginCountry   string `json:"origin_country"`
	SerialNumber    string `json:"serial_number"`
	Make            string `json:"make"`
	Model           string `json:"model"`
	Caliber         string `json:"caliber"`
	WeaponType      string `json:"weapon_type"`
	RecoveryDate    string `json:"recovery_date"`
	RecoveryCountry string `json:"recovery_country"`
	Notes           string `json:"notes"`
}

type IARMSEntry struct {
	IARMSRef        string `json:"iarms_ref"`
	SerialNumber    string `json:"serial_number"`
	Make            string `json:"make"`
	Model           string `json:"model"`
	Caliber         string `json:"caliber"`
	WeaponType      string `json:"weapon_type"`
	RecoveryDate    string `json:"recovery_date"`
	OriginCountry   string `json:"origin_country"`
	RecoveryCountry string `json:"recovery_country"`
}
