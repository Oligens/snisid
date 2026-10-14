package audit

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type Entry struct {
	EventType  string         `json:"event_type"`
	TableName  string         `json:"table_name"`
	RecordID   string         `json:"record_id"`
	OfficerNIU string         `json:"officer_niu"`
	AgencyCode string         `json:"agency_code"`
	Purpose    string         `json:"purpose"`
	CaseNumber string         `json:"case_number,omitempty"`
	Action     string         `json:"action"`
	Details    map[string]any `json:"details,omitempty"`
	Signature  string         `json:"signature"`
	CreatedAt  int64          `json:"created_at"`
}

type Logger struct {
	privateKey *ecdsa.PrivateKey
}

func NewLogger() (*Logger, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ecdsa key: %w", err)
	}
	return &Logger{privateKey: key}, nil
}

func (l *Logger) Log(ctx context.Context, eventType, tableName, recordID, officerNIU, agencyCode, purpose, action string, details map[string]any) (*Entry, error) {
	entry := &Entry{
		EventType:  eventType,
		TableName:  tableName,
		RecordID:   recordID,
		OfficerNIU: officerNIU,
		AgencyCode: agencyCode,
		Purpose:    purpose,
		Action:     action,
		Details:    details,
		CreatedAt:  time.Now().UnixMilli(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("marshal entry: %w", err)
	}

	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, l.privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign entry: %w", err)
	}

	sig := append(r.Bytes(), s.Bytes()...)
	entry.Signature = fmt.Sprintf("%x", sig)

	return entry, nil
}

func (l *Logger) Verify(entry *Entry) bool {
	data, err := json.Marshal(entry)
	if err != nil {
		return false
	}
	hash := sha256.Sum256(data)

	sig, err := parseSignature(entry.Signature)
	if err != nil {
		return false
	}

	return ecdsa.Verify(&l.privateKey.PublicKey, hash[:], sig[0], sig[1])
}

func parseSignature(hex string) ([]*big.Int, error) {
	raw := []byte(hex)
	if len(raw) < 2 {
		return nil, fmt.Errorf("invalid signature length")
	}
	r := new(big.Int).SetBytes(raw[:len(raw)/2])
	s := new(big.Int).SetBytes(raw[len(raw)/2:])
	return []*big.Int{r, s}, nil
}
