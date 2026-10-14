package alerter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type AlertEvent struct {
	AlertID       string    `json:"alert_id"`
	PlateNumber   string    `json:"plate_number"`
	CrimeCategory string    `json:"crime_category"`
	AlertLevel    string    `json:"alert_level"`
	ReportingUnit string    `json:"reporting_unit"`
	Timestamp     time.Time `json:"timestamp"`
}

type SightingAlertEvent struct {
	PlateNumber  string    `json:"plate_number"`
	AlertLevel   string    `json:"alert_level"`
	LocationLat  *float64  `json:"location_lat,omitempty"`
	LocationLng  *float64  `json:"location_lng,omitempty"`
	DeptCode     *string   `json:"dept_code,omitempty"`
	SightingTime time.Time `json:"sighting_time"`
	LAPIUnitID   *string   `json:"lapi_unit_id,omitempty"`
}

type Alerter struct {
	pushGatewayURL string
	smsGatewayURL  string
	radioEndpoint  string
	client         *http.Client
	logger         *zap.Logger
}

func New(pushURL, smsURL, radioURL string, logger *zap.Logger) *Alerter {
	return &Alerter{
		pushGatewayURL: pushURL,
		smsGatewayURL:  smsURL,
		radioEndpoint:  radioURL,
		client:         &http.Client{Timeout: 10 * time.Second},
		logger:         logger,
	}
}

func (a *Alerter) DispatchAlert(ctx context.Context, raw []byte) error {
	var evt AlertEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return fmt.Errorf("unmarshal alert event: %w", err)
	}

	if a.pushGatewayURL != "" {
		go a.sendPush(context.Background(), evt)
	}
	if a.smsGatewayURL != "" {
		go a.sendSMS(context.Background(), evt)
	}
	if a.radioEndpoint != "" && (evt.AlertLevel == "CRITICAL" || evt.AlertLevel == "WANTED") {
		go a.sendRadioBroadcast(context.Background(), evt)
	}

	a.logger.Info("Alerte dispatchée", zap.String("plate", evt.PlateNumber), zap.String("level", evt.AlertLevel))
	return nil
}

func (a *Alerter) DispatchSightingAlert(ctx context.Context, raw []byte) error {
	var evt SightingAlertEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return fmt.Errorf("unmarshal sighting event: %w", err)
	}
	a.logger.Info("Alerte visuelle dispatchée",
		zap.String("plate", evt.PlateNumber),
		zap.String("level", evt.AlertLevel),
	)
	return nil
}

func (a *Alerter) sendPush(ctx context.Context, evt AlertEvent) {
	body, _ := json.Marshal(evt)
	resp, err := a.client.Post(a.pushGatewayURL, "application/json", nil)
	if err != nil {
		a.logger.Warn("Échec notification push", zap.Error(err))
		return
	}
	defer resp.Body.Close()
	_ = body
}

func (a *Alerter) sendSMS(ctx context.Context, evt AlertEvent) {
	resp, err := a.client.Post(a.smsGatewayURL, "application/json", nil)
	if err != nil {
		a.logger.Warn("Échec envoi SMS", zap.Error(err))
		return
	}
	defer resp.Body.Close()
}

func (a *Alerter) sendRadioBroadcast(ctx context.Context, evt AlertEvent) {
	resp, err := a.client.Post(a.radioEndpoint, "application/json", nil)
	if err != nil {
		a.logger.Warn("Échec diffusion radio", zap.Error(err))
		return
	}
	defer resp.Body.Close()
}
