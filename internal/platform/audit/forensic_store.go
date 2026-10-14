package audit

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type ForensicEntry struct {
	EventID  string
	Payload  []byte
	PrevHash string
	Hash     string
}

type ForensicStore struct {
	mu       sync.Mutex
	lastHash string
}

func NewForensicStore() *ForensicStore {
	return &ForensicStore{
		lastHash: "NATIONAL_GENESIS_BLOCK",
	}
}

func (s *ForensicStore) Commit(ctx context.Context, eventID string, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Info(ctx, "Committing event to Forensic Audit Ledger", zap.String("event_id", eventID))

	// 1. Calculate Cryptographic Chain Link
	dataToHash := fmt.Sprintf("%s:%s:%s", s.lastHash, eventID, string(payload))
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(dataToHash)))

	// 2. Mock: Write to Immutable Storage (e.g., ClickHouse with Object Lock)
	s.lastHash = hash

	logger.Info(ctx, "Forensic commitment successful",
		zap.String("event_id", eventID),
		zap.String("chain_hash", hash),
	)

	return nil
}
