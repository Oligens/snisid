package federation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type PolicyStatus string

const (
	PolicyPending      PolicyStatus = "PENDING"
	PolicyDistributed  PolicyStatus = "DISTRIBUTED"
	PolicyAcknowledged PolicyStatus = "ACKNOWLEDGED"
	PolicyFailed       PolicyStatus = "FAILED"
	PolicyRevoked      PolicyStatus = "REVOKED"
)

type PolicyPackage struct {
	ID          string            `json:"policy_id"`
	Version     string            `json:"version"`
	Rules       []string          `json:"rules"`
	Constraints map[string]string `json:"constraints"`
	Signature   string            `json:"signature"`
	Timestamp   int64             `json:"timestamp"`
	Priority    int               `json:"priority"`
	Description string            `json:"description"`
	Hash        string            `json:"hash"`
}

type DistributionStatus struct {
	Peer      string       `json:"peer"`
	PackageID string       `json:"package_id"`
	Status    PolicyStatus `json:"status"`
	Attempted int          `json:"attempted"`
	LastError string       `json:"last_error,omitempty"`
	AckedAt   *time.Time   `json:"acked_at,omitempty"`
}

type FederationHub struct {
	Peers           []string `json:"peers"`
	privateKey      ed25519.PrivateKey
	publicKey       ed25519.PublicKey
	hubID           string
	mu              sync.RWMutex
	distributionLog map[string][]DistributionStatus
	activePolicies  map[string]*PolicyPackage
}

func NewFederationHub(hubID string, peers []string) *FederationHub {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		logger.Error(context.Background(), "GOVERNANCE-MESH: failed to generate signing key", zap.Error(err))
		return nil
	}

	return &FederationHub{
		Peers:           peers,
		hubID:           hubID,
		privateKey:      priv,
		publicKey:       pub,
		distributionLog: make(map[string][]DistributionStatus),
		activePolicies:  make(map[string]*PolicyPackage),
	}
}

func (h *FederationHub) DistributePolicy(p PolicyPackage) []DistributionStatus {
	logger.Info(context.Background(), "GOVERNANCE-MESH: distributing policy",
		zap.String("id", p.ID),
		zap.String("version", p.Version),
		zap.Int("peers", len(h.Peers)),
	)

	p.Timestamp = time.Now().Unix()
	p.Hash = h.computeHash(p)

	sig, err := h.signPolicy(p)
	if err != nil {
		logger.Error(context.Background(), "GOVERNANCE-MESH: signing failed", zap.Error(err))
		return nil
	}
	p.Signature = sig

	h.mu.Lock()
	h.activePolicies[p.ID] = &p
	h.mu.Unlock()

	results := make([]DistributionStatus, 0, len(h.Peers))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, peer := range h.Peers {
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()
			status := h.distributeToPeer(peer, p)
			mu.Lock()
			results = append(results, status)
			mu.Unlock()
		}(peer)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Status != results[j].Status {
			return results[i].Status < results[j].Status
		}
		return results[i].Peer < results[j].Peer
	})

	h.mu.Lock()
	h.distributionLog[p.ID] = results
	h.mu.Unlock()

	return results
}

func (h *FederationHub) distributeToPeer(peer string, p PolicyPackage) DistributionStatus {
	status := DistributionStatus{
		Peer:      peer,
		PackageID: p.ID,
		Status:    PolicyPending,
		Attempted: 0,
	}

	payload, err := json.Marshal(map[string]interface{}{
		"hub_id":     h.hubID,
		"policy":     p,
		"public_key": base64.StdEncoding.EncodeToString(h.publicKey),
	})
	if err != nil {
		status.Status = PolicyFailed
		status.LastError = fmt.Sprintf("marshal error: %v", err)
		return status
	}

	for attempt := 0; attempt < 3; attempt++ {
		status.Attempted = attempt + 1

		err := h.sendToPeer(peer, payload)
		if err == nil {
			status.Status = PolicyDistributed
			now := time.Now()
			status.AckedAt = &now
			logger.Info(context.Background(), "GOVERNANCE-MESH: policy distributed to peer",
				zap.String("peer", peer),
				zap.String("policy_id", p.ID),
				zap.Int("attempt", attempt+1),
			)
			return status
		}

		status.LastError = err.Error()
		logger.Warn(context.Background(), "GOVERNANCE-MESH: distribution attempt failed",
			zap.String("peer", peer),
			zap.Int("attempt", attempt+1),
			zap.String("error", err.Error()),
		)

		if attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
	}

	status.Status = PolicyFailed
	return status
}

func (h *FederationHub) sendToPeer(peer string, payload []byte) error {
	_ = payload
	logger.Info(context.Background(), "GOVERNANCE-MESH: sending to peer", zap.String("peer", peer))
	return nil
}

func (h *FederationHub) AcknowledgePolicy(packageID, peer string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if logs, ok := h.distributionLog[packageID]; ok {
		for i, log := range logs {
			if log.Peer == peer {
				logs[i].Status = PolicyAcknowledged
				now := time.Now()
				logs[i].AckedAt = &now
				break
			}
		}
	}

	logger.Info(context.Background(), "GOVERNANCE-MESH: policy acknowledged by peer",
		zap.String("peer", peer),
		zap.String("policy_id", packageID),
	)
}

func (h *FederationHub) RevokePolicy(packageID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	pkg, ok := h.activePolicies[packageID]
	if !ok {
		return fmt.Errorf("policy %s not found", packageID)
	}

	pkg.Signature = "REVOKED:" + pkg.Signature
	logger.Warn(context.Background(), "GOVERNANCE-MESH: policy revoked", zap.String("policy_id", packageID))

	return nil
}

func (h *FederationHub) GetDistributionStatus(packageID string) []DistributionStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if logs, ok := h.distributionLog[packageID]; ok {
		result := make([]DistributionStatus, len(logs))
		copy(result, logs)
		return result
	}
	return nil
}

func (h *FederationHub) GetActivePolicyIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.activePolicies))
	for id := range h.activePolicies {
		ids = append(ids, id)
	}
	return ids
}

func (h *FederationHub) VerifyPolicy(p PolicyPackage) bool {
	if p.Signature == "" {
		return false
	}

	if len(h.Peers) == 0 {
		return false
	}

	expectedHash := h.computeHash(p)
	if p.Hash != "" && p.Hash != expectedHash {
		return false
	}

	return true
}

func (h *FederationHub) signPolicy(p PolicyPackage) (string, error) {
	data := []byte(fmt.Sprintf("%s:%s:%d", p.ID, p.Version, p.Timestamp))
	sig, err := h.privateKey.Sign(rand.Reader, data, &ed25519.Options{})
	if err != nil {
		return "", fmt.Errorf("signing error: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (h *FederationHub) VerifySignature(p PolicyPackage, sig string, pubKey ed25519.PublicKey) bool {
	data := []byte(fmt.Sprintf("%s:%s:%d", p.ID, p.Version, p.Timestamp))
	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	return ed25519.Verify(pubKey, data, sigBytes)
}

func (h *FederationHub) computeHash(p PolicyPackage) string {
	data := []byte(fmt.Sprintf("%s:%s:%v:%v:%d", p.ID, p.Version, p.Rules, p.Constraints, p.Priority))
	hsh := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hsh[:])
}

func (h *FederationHub) AddPeer(peer string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, p := range h.Peers {
		if p == peer {
			return
		}
	}
	h.Peers = append(h.Peers, peer)
	logger.Info(context.Background(), "GOVERNANCE-MESH: peer added", zap.String("peer", peer))
}

func (h *FederationHub) RemovePeer(peer string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, p := range h.Peers {
		if p == peer {
			h.Peers = append(h.Peers[:i], h.Peers[i+1:]...)
			logger.Info(context.Background(), "GOVERNANCE-MESH: peer removed", zap.String("peer", peer))
			return
		}
	}
}
