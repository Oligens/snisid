package federation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFederationHub(t *testing.T) {
	peers := []string{"peer-1", "peer-2", "peer-3"}
	h := NewFederationHub("hub-ht", peers)

	require.NotNil(t, h)
	assert.Equal(t, "hub-ht", h.hubID)
	assert.Equal(t, peers, h.Peers)
	assert.NotNil(t, h.distributionLog)
	assert.NotNil(t, h.activePolicies)
}

func TestNewFederationHub_NoPeers(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{})
	require.NotNil(t, h)
	assert.Empty(t, h.Peers)
}

func TestDistributePolicy_Success(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1", "peer-2"})
	require.NotNil(t, h)

	p := PolicyPackage{
		ID:      "POL-001",
		Version: "v1.0",
		Rules:   []string{"rule-1", "rule-2"},
		Constraints: map[string]string{
			"DATA_RETENTION": "10y",
		},
		Priority:    1,
		Description: "Test policy",
	}

	results := h.DistributePolicy(p)
	require.Len(t, results, 2)
	for _, r := range results {
		assert.Equal(t, PolicyDistributed, r.Status)
		assert.Equal(t, 1, r.Attempted)
		assert.NotNil(t, r.AckedAt)
	}
}

func TestDistributePolicy_SigningAndVerification(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	require.NotNil(t, h)

	p := PolicyPackage{
		ID:       "POL-002",
		Version:  "v2.0",
		Rules:    []string{"critical-rule"},
		Priority: 5,
	}

	results := h.DistributePolicy(p)
	require.Len(t, results, 1)
	assert.Equal(t, PolicyDistributed, results[0].Status)

	active := h.GetActivePolicyIDs()
	assert.Contains(t, active, "POL-002")
}

func TestDistributePolicy_ChecksSignature(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	require.NotNil(t, h)

	p := PolicyPackage{
		ID:      "POL-SIG",
		Version: "v1.0",
	}
	h.DistributePolicy(p)

	stored, ok := h.activePolicies["POL-SIG"]
	require.True(t, ok)
	assert.NotEmpty(t, stored.Signature)
	assert.NotEmpty(t, stored.Hash)
}

func TestVerifyPolicy_EmptySignature(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	require.NotNil(t, h)

	p := PolicyPackage{ID: "POL-003"}
	assert.False(t, h.VerifyPolicy(p))
}

func TestVerifyPolicy_NoPeers(t *testing.T) {
	h := NewFederationHub("hub-empty", []string{})
	p := PolicyPackage{ID: "POL-003", Signature: "sig"}
	assert.False(t, h.VerifyPolicy(p))
}

func TestVerifyPolicy_HashMismatch(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	p := PolicyPackage{
		ID:        "POL-004",
		Version:   "v1.0",
		Signature: "valid",
		Hash:      "wronghash",
	}
	assert.False(t, h.VerifyPolicy(p))
}

func TestVerifyPolicy_Valid(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	p := PolicyPackage{
		ID:        "POL-VALID",
		Version:   "v1.0",
		Signature: "mocked-sig",
	}
	assert.True(t, h.VerifyPolicy(p))
}

func TestVerifySignature_Valid(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	h := NewFederationHub("hub-ht", []string{"peer-1"})
	h.privateKey = priv
	h.publicKey = pub

	p := PolicyPackage{ID: "POL-SIG", Version: "v1.0", Timestamp: time.Now().Unix()}
	sig, err := h.signPolicy(p)
	require.NoError(t, err)

	assert.True(t, h.VerifySignature(p, sig, pub))
}

func TestVerifySignature_Invalid(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	h := NewFederationHub("hub-ht", []string{"peer-1"})

	p := PolicyPackage{ID: "POL-SIG", Version: "v1.0"}
	assert.False(t, h.VerifySignature(p, "bad-base64!", pub))
	assert.False(t, h.VerifySignature(p, base64.StdEncoding.EncodeToString([]byte("bogus")), pub))
}

func TestAcknowledgePolicy(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1", "peer-2"})
	p := PolicyPackage{ID: "POL-ACK", Version: "v1.0", Rules: []string{"r1"}}
	h.DistributePolicy(p)

	h.AcknowledgePolicy("POL-ACK", "peer-1")
	status := h.GetDistributionStatus("POL-ACK")

	require.NotNil(t, status)
	var found bool
	for _, s := range status {
		if s.Peer == "peer-1" {
			assert.Equal(t, PolicyAcknowledged, s.Status)
			assert.NotNil(t, s.AckedAt)
			found = true
		}
	}
	assert.True(t, found)
}

func TestRevokePolicy_Success(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	p := PolicyPackage{ID: "POL-REV", Version: "v1.0", Rules: []string{"r1"}}
	h.DistributePolicy(p)

	err := h.RevokePolicy("POL-REV")
	assert.NoError(t, err)

	stored, _ := h.activePolicies["POL-REV"]
	assert.Contains(t, stored.Signature, "REVOKED:")
}

func TestRevokePolicy_NotFound(t *testing.T) {
	h := NewFederationHub("hub-ht", nil)
	err := h.RevokePolicy("NONEXISTENT")
	assert.Error(t, err)
}

func TestAddPeer_Duplicate(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	h.AddPeer("peer-1")
	assert.Len(t, h.Peers, 1)
}

func TestAddPeer_New(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	h.AddPeer("peer-2")
	assert.Len(t, h.Peers, 2)
	assert.Contains(t, h.Peers, "peer-2")
}

func TestRemovePeer_Existing(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1", "peer-2"})
	h.RemovePeer("peer-1")
	assert.Len(t, h.Peers, 1)
	assert.NotContains(t, h.Peers, "peer-1")
}

func TestRemovePeer_NonExisting(t *testing.T) {
	h := NewFederationHub("hub-ht", []string{"peer-1"})
	h.RemovePeer("unknown")
	assert.Len(t, h.Peers, 1)
}

func TestGetDistributionStatus_NotFound(t *testing.T) {
	h := NewFederationHub("hub-ht", nil)
	assert.Nil(t, h.GetDistributionStatus("NONEXISTENT"))
}

func TestConcurrentDistribute(t *testing.T) {
	peers := make([]string, 5)
	for i := range peers {
		peers[i] = fmt.Sprintf("peer-%d", i)
	}
	h := NewFederationHub("hub-ht", peers)
	require.NotNil(t, h)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p := PolicyPackage{
				ID:      fmt.Sprintf("POL-CONC-%d", id),
				Version: "v1.0",
				Rules:   []string{fmt.Sprintf("rule-%d", id)},
			}
			h.DistributePolicy(p)
		}(i)
	}
	wg.Wait()

	ids := h.GetActivePolicyIDs()
	assert.Len(t, ids, 3)
}

func TestComputeHash_Deterministic(t *testing.T) {
	h := NewFederationHub("hub-ht", nil)
	p1 := PolicyPackage{ID: "POL-001", Version: "v1", Rules: []string{"a", "b"}, Priority: 1}
	p2 := PolicyPackage{ID: "POL-001", Version: "v1", Rules: []string{"a", "b"}, Priority: 1}

	h1 := h.computeHash(p1)
	h2 := h.computeHash(p2)
	assert.Equal(t, h1, h2)
}

func TestComputeHash_DifferentInputs(t *testing.T) {
	h := NewFederationHub("hub-ht", nil)
	p1 := PolicyPackage{ID: "POL-001", Version: "v1", Rules: []string{"a"}, Priority: 1}
	p2 := PolicyPackage{ID: "POL-001", Version: "v2", Rules: []string{"a"}, Priority: 1}

	h1 := h.computeHash(p1)
	h2 := h.computeHash(p2)
	assert.NotEqual(t, h1, h2)
}
