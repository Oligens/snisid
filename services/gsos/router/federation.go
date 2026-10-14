package router

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type AlertLevel int

const (
	AlertInfo     AlertLevel = 1
	AlertWarning  AlertLevel = 2
	AlertCritical AlertLevel = 3
	AlertGlobal   AlertLevel = 4
)

type GSPEvent struct {
	ID         string     `json:"id"`
	Country    string     `json:"country"`
	Category   string     `json:"category"`
	Risk       float64    `json:"risk"`
	AlertLevel AlertLevel `json:"alert_level"`
	Payload    []byte     `json:"payload"`
	Timestamp  int64      `json:"timestamp"`
}

type GlobalAlert struct {
	ID                string     `json:"id"`
	SourceEvent       GSPEvent   `json:"source_event"`
	Severity          AlertLevel `json:"severity"`
	AffectedCountries []string   `json:"affected_countries"`
	PropagatedAt      int64      `json:"propagated_at"`
	AcksReceived      int        `json:"acks_received"`
}

type FederationRouter struct {
	ActiveNodes   []string            `json:"active_nodes"`
	nodeStatus    map[string]bool     // node -> online
	alertHistory  []GlobalAlert       `json:"alert_history"`
	policyCache   map[string][]string // country -> policies
	mu            sync.RWMutex
	propagationFn func(alert GlobalAlert, target string) error
}

func NewFederationRouter(nodes []string) *FederationRouter {
	nodeStatus := make(map[string]bool)
	for _, n := range nodes {
		nodeStatus[n] = true
	}
	return &FederationRouter{
		ActiveNodes:  nodes,
		nodeStatus:   nodeStatus,
		alertHistory: []GlobalAlert{},
		policyCache:  make(map[string][]string),
	}
}

func (r *FederationRouter) SetPropagationFn(fn func(alert GlobalAlert, target string) error) {
	r.propagationFn = fn
}

func (r *FederationRouter) RouteEvent(event GSPEvent) {
	logger.Info(context.Background(), "GSOS-ROUTER: routing event",
		zap.String("id", event.ID),
		zap.String("country", event.Country),
		zap.String("category", event.Category),
		zap.Any("risk", event.Risk),
	)

	alertLevel := r.computeAlertLevel(event)
	event.AlertLevel = alertLevel

	if alertLevel >= AlertCritical {
		alert := GlobalAlert{
			ID:                fmt.Sprintf("GA-%s-%d", event.Country, event.Timestamp),
			SourceEvent:       event,
			Severity:          alertLevel,
			AffectedCountries: r.computeAffectedCountries(event),
			PropagatedAt:      time.Now().Unix(),
		}
		r.PropagateGlobalAlert(alert)
	}

	r.mu.Lock()
	r.alertHistory = append(r.alertHistory, GlobalAlert{
		ID:          event.ID,
		SourceEvent: event,
		Severity:    alertLevel,
	})
	if len(r.alertHistory) > 1000 {
		r.alertHistory = r.alertHistory[len(r.alertHistory)-1000:]
	}
	r.mu.Unlock()
}

func (r *FederationRouter) computeAlertLevel(event GSPEvent) AlertLevel {
	switch {
	case event.Risk > 0.95:
		return AlertGlobal
	case event.Risk > 0.85:
		return AlertCritical
	case event.Risk > 0.7:
		return AlertWarning
	default:
		return AlertInfo
	}
}

func (r *FederationRouter) computeAffectedCountries(event GSPEvent) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var affected []string
	for _, node := range r.ActiveNodes {
		if node != event.Country {
			affected = append(affected, node)
		}
	}
	return affected
}

func (r *FederationRouter) PropagateGlobalAlert(alert GlobalAlert) {
	logger.Warn(context.Background(), "GSOS-ROUTER: propagating global alert",
		zap.String("alert_id", alert.ID),
		zap.String("severity", alert.Severity),
		zap.Int("targets", len(alert.AffectedCountries)),
	)

	var wg sync.WaitGroup
	ackCount := 0
	var mu sync.Mutex

	for _, target := range alert.AffectedCountries {
		if !r.isNodeOnline(target) {
			logger.Warn(context.Background(), "GSOS-ROUTER: target node offline, skipping", zap.String("node", target))
			continue
		}

		wg.Add(1)
		go func(target string) {
			defer wg.Done()

			if r.propagationFn != nil {
				if err := r.propagationFn(alert, target); err != nil {
					logger.Error(context.Background(), "GSOS-ROUTER: propagation failed", zap.String("target", target), zap.Error(err))
					return
				}
			}

			mu.Lock()
			ackCount++
			mu.Unlock()

			logger.Info(context.Background(), "GSOS-ROUTER: alert acknowledged by", zap.String("node", target))
		}(target)
	}

	wg.Wait()

	alert.AcksReceived = ackCount
	r.mu.Lock()
	for i, a := range r.alertHistory {
		if a.ID == alert.ID {
			r.alertHistory[i].AcksReceived = ackCount
			r.alertHistory[i].PropagatedAt = alert.PropagatedAt
			break
		}
	}
	r.mu.Unlock()

	logger.Info(context.Background(), "GSOS-ROUTER: global alert propagated",
		zap.String("alert_id", alert.ID),
		zap.Int("acks", ackCount),
		zap.Int("total_targets", len(alert.AffectedCountries)),
	)
}

func (r *FederationRouter) SyncLocalPolicies(country string) []string {
	r.mu.RLock()
	policies, ok := r.policyCache[country]
	r.mu.RUnlock()

	if ok {
		return policies
	}

	defaultPolicies := []string{
		"EVENT_SHARING_AGREEMENT",
		"IDENTITY_CROSSCHECK",
		"FPR_REALTIME_SYNC",
		"BIOMETRIC_EXCHANGE",
		"FRAUD_ALERT_PROPAGATION",
	}

	r.mu.Lock()
	r.policyCache[country] = defaultPolicies
	r.mu.Unlock()

	logger.Info(context.Background(), "GSOS-ROUTER: synchronized policies for", zap.String("country", country), zap.Int("policies", len(defaultPolicies)))
	return defaultPolicies
}

func (r *FederationRouter) isNodeOnline(node string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodeStatus[node]
}

func (r *FederationRouter) SetNodeStatus(node string, online bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeStatus[node] = online
	logger.Info(context.Background(), "GSOS-ROUTER: node status changed", zap.String("node", node), zap.Bool("online", online))
}

func (r *FederationRouter) GetAlertHistory(since int64) []GlobalAlert {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []GlobalAlert
	for _, a := range r.alertHistory {
		if a.PropagatedAt >= since || since == 0 {
			result = append(result, a)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].PropagatedAt > result[j].PropagatedAt
	})

	return result
}
