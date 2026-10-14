package domain

import (
	"time"

	"github.com/google/uuid"
)

type SyncStatus string

const (
	SyncPending  SyncStatus = "PENDING"
	SyncSyncing  SyncStatus = "SYNCING"
	SyncSynced   SyncStatus = "SYNCED"
	SyncConflict SyncStatus = "CONFLICT"
	SyncFailed   SyncStatus = "FAILED"
)

type SyncQueueItem struct {
	ID         uuid.UUID  `json:"id"`
	TerminalID uuid.UUID  `json:"terminal_id"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	Action     string     `json:"action"`
	Payload    string     `json:"payload"`
	Status     SyncStatus `json:"status"`
	RetryCount int        `json:"retry_count"`
	ErrorMsg   *string    `json:"error_msg,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	SyncedAt   *time.Time `json:"synced_at,omitempty"`
}

type OfflineTerminal struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Location    string     `json:"location"`
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"`
	FirmwareVer string     `json:"firmware_ver"`
	IsOnline    bool       `json:"is_online"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PushQueueRequest struct {
	TerminalID string `json:"terminal_id" binding:"required"`
	EntityType string `json:"entity_type" binding:"required"`
	EntityID   string `json:"entity_id" binding:"required"`
	Action     string `json:"action" binding:"required"`
	Payload    string `json:"payload" binding:"required"`
}

type SyncRequest struct {
	TerminalID string `json:"-" uri:"terminal_id" binding:"required"`
}
