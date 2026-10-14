package forensic

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type SignedAuditEntry struct {
	EventType  string         `json:"event_type"`
	TableName  string         `json:"table_name"`
	RecordID   string         `json:"record_id,omitempty"`
	SampleID   string         `json:"sample_id,omitempty"`
	OfficerNIU string         `json:"officer_niu,omitempty"`
	AgencyCode string         `json:"agency_code,omitempty"`
	Purpose    string         `json:"purpose,omitempty"`
	CaseNumber string         `json:"case_number,omitempty"`
	Action     string         `json:"action,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	Timestamp  string         `json:"timestamp"`
	Signature  string         `json:"signature,omitempty"`
}

func NewAuditEntry(eventType, tableName string) *SignedAuditEntry {
	return &SignedAuditEntry{
		EventType: eventType,
		TableName: tableName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (e *SignedAuditEntry) canonical() []byte {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		e.EventType, e.TableName, e.RecordID, e.SampleID,
		e.OfficerNIU, e.AgencyCode, e.Purpose, e.CaseNumber,
		e.Action, e.Timestamp,
	)
	return []byte(data)
}

func (e *SignedAuditEntry) Sign(privateKeyPEM []byte) error {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse EC key: %w", err)
	}
	digest := sha256.Sum256(e.canonical())
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		return fmt.Errorf("ecdsa sign: %w", err)
	}
	sig := append(r.Bytes(), s.Bytes()...)
	e.Signature = hex.EncodeToString(sig)
	return nil
}

func (e *SignedAuditEntry) Verify(publicKeyPEM []byte) (bool, error) {
	if e.Signature == "" {
		return false, nil
	}
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return false, fmt.Errorf("failed to decode PEM block")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := pubAny.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an ECDSA public key")
	}
	digest := sha256.Sum256(e.canonical())
	sigBytes, err := hex.DecodeString(e.Signature)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}
	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])
	return ecdsa.Verify(pub, digest[:], r, s), nil
}

func GenerateAuditKeypair() (privPEM, pubPEM []byte, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	privDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return privPEM, pubPEM, nil
}
