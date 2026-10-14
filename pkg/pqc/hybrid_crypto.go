//go:build pqc

// === pkg/pqc/hybrid_crypto.go ===
// SNISID - Cryptographie Post-Quantique Hybride
// Conforme NIST FIPS 203 (ML-KEM/Kyber) + FIPS 204 (ML-DSA/Dilithium)
package pqc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/open-quantum-safe/liboqs-go/oqs"
	"time"
)

// HybridKEM: RSA-4096 + ML-KEM-1024 (CRYSTALS-Kyber)
// Niveau securite 5 = 256-bit quantum security (NIST FIPS 203)
type HybridKEM struct {
	kemAlgo string // "ML-KEM-1024"
}

type HybridKEMKeyPair struct {
	RSAPrivKey   *rsa.PrivateKey
	RSAPubKey    *rsa.PublicKey
	KyberPubKey  []byte
	KyberPrivKey []byte // En HSM en production
}

func NewHybridKEM() *HybridKEM {
	return &HybridKEM{kemAlgo: "ML-KEM-1024"}
}

func (h *HybridKEM) GenerateKeyPair() (*HybridKEMKeyPair, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("RSA-4096 keygen: %w", err)
	}
	kem := oqs.KeyEncapsulation{}
	if err := kem.Init(h.kemAlgo, nil); err != nil {
		return nil, fmt.Errorf("ML-KEM-1024 init: %w", err)
	}
	defer kem.Clean()
	pubKey, err := kem.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("ML-KEM-1024 keygen: %w", err)
	}
	return &HybridKEMKeyPair{
		RSAPrivKey:   rsaKey,
		RSAPubKey:    &rsaKey.PublicKey,
		KyberPubKey:  pubKey,
		KyberPrivKey: kem.ExportSecretKey(),
	}, nil
}

// HybridSigner: RSA-PSS-4096 + ML-DSA-87 (CRYSTALS-Dilithium FIPS 204)
type HybridSigner struct {
	sigAlgo string // "ML-DSA-87"
}

type HybridSignature struct {
	RSASignature       []byte
	DilithiumSignature []byte
	Algorithm          string // "RSA-PSS-4096+ML-DSA-87"
	Timestamp          int64
}

func NewHybridSigner() *HybridSigner {
	return &HybridSigner{sigAlgo: "ML-DSA-87"}
}

// SignDecret signe un decret presidentiel avec signature hybride
// Double signature = securite classique + securite quantique
func (s *HybridSigner) SignDecret(
	docHash []byte,
	rsaKey *rsa.PrivateKey,
	dilithiumPrivKey []byte,
) (*HybridSignature, error) {

	// Signature RSA-PSS-4096 (compatibilite legacy)
	rsaSig, err := rsa.SignPSS(rand.Reader, rsaKey, crypto.SHA512, docHash,
		&rsa.PSSOptions{SaltLength: 64, Hash: crypto.SHA512})
	if err != nil {
		return nil, fmt.Errorf("RSA-PSS sign: %w", err)
	}

	// Signature ML-DSA-87 (protection post-quantique)
	sig := oqs.Signature{}
	if err := sig.Init(s.sigAlgo, dilithiumPrivKey); err != nil {
		return nil, fmt.Errorf("ML-DSA init: %w", err)
	}
	defer sig.Clean()
	dilithiumSig, err := sig.Sign(docHash)
	if err != nil {
		return nil, fmt.Errorf("ML-DSA sign: %w", err)
	}

	return &HybridSignature{
		RSASignature:       rsaSig,
		DilithiumSignature: dilithiumSig,
		Algorithm:          "RSA-PSS-4096+ML-DSA-87",
		Timestamp:          time.Now().UnixNano(),
	}, nil
}

// Plan de migration PKI en 5 phases:
// Phase 1 (M0-M6):   Generer cles PQC hybrides en parallele des RSA existantes
// Phase 2 (M6-M12):  Emettre certificats hybrides pour nouveaux services
// Phase 3 (M12-M24): Migrer progressivement les certificats existants
// Phase 4 (M24-M36): Deprecier RSA seul, PQC hybride obligatoire
// Phase 5 (M36+):    Evaluer transition vers PQC pur (post-quantum only)

// Algorithmes NIST recommandes:
// Echange de cles: ML-KEM-1024 (FIPS 203) = CRYSTALS-Kyber
// Signatures:      ML-DSA-87   (FIPS 204) = CRYSTALS-Dilithium
// Signatures HVA:  SLH-DSA     (FIPS 205) = SPHINCS+ pour decrets critiques
