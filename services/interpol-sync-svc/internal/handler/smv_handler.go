package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type SMVHandler struct {
	db         *sqlx.DB
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewSMVHandler(db *sqlx.DB, baseURL, apiKey string, logger *zap.Logger) *SMVHandler {
	return &SMVHandler{
		db:      db,
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

type PendingSync struct {
	SyncID    string `db:"sync_id"`
	AlertID   string `db:"alert_id"`
	PlateNum  string `db:"plate_number"`
	Direction string `db:"sync_direction"`
}

func (h *SMVHandler) SyncPending(ctx context.Context) error {
	var pending []PendingSync
	query := `
		SELECT s.sync_id, s.alert_id, a.plate_number, s.sync_direction
		FROM sivc_interpol_sync_log s
		JOIN sivc_criminal_alerts a ON s.alert_id = a.alert_id
		WHERE s.sync_status = 'PENDING' AND s.sync_direction = 'OUTBOUND'
		ORDER BY s.sync_timestamp ASC
		LIMIT 50
	`

	if err := h.db.SelectContext(ctx, &pending, query); err != nil {
		return fmt.Errorf("failed to fetch pending syncs: %w", err)
	}

	for _, p := range pending {
		if err := h.syncRecord(ctx, p); err != nil {
			h.logger.Error("Failed to sync record",
				zap.String("sync_id", p.SyncID),
				zap.Error(err),
			)
			h.updateStatus(ctx, p.SyncID, "FAILED", err.Error())
			continue
		}
		h.updateStatus(ctx, p.SyncID, "SUCCESS", "")
	}

	return nil
}

func (h *SMVHandler) syncRecord(ctx context.Context, record PendingSync) error {
	payload, _ := json.Marshal(map[string]string{
		"alert_id":     record.AlertID,
		"plate_number": record.PlateNum,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/smv/vehicles", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.apiKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("INTERPOL returned status %d", resp.StatusCode)
	}

	return nil
}

func (h *SMVHandler) updateStatus(ctx context.Context, syncID, status, errorMsg string) {
	query := `UPDATE sivc_interpol_sync_log SET sync_status = $1, error_message = $2, processed_at = NOW() WHERE sync_id = $3`
	h.db.ExecContext(ctx, query, status, errorMsg, syncID)
}
