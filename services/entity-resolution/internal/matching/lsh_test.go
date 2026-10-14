package matching

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetCoeffs() {
	coeffMu.Lock()
	coeffsA = nil
	coeffsB = nil
	coeffMu.Unlock()
}

func TestNewLSHIndex(t *testing.T) {
	idx := NewLSHIndex(100, 20, 5)
	require.NotNil(t, idx)
	assert.Equal(t, 100, idx.numHashes)
	assert.Equal(t, 20, idx.numBands)
	assert.Equal(t, 5, idx.rowsPerBand)
	assert.NotNil(t, idx.buckets)
	assert.NotNil(t, idx.signatures)
}

func TestLSHIndex_InsertAndQuery(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)

	features := make([]float64, 256)
	features[0] = 1.0
	features[10] = 0.5
	features[100] = 0.8

	idx.Insert("id-1", features)

	results := idx.Query(features)
	require.Len(t, results, 1)
	assert.Equal(t, "id-1", results[0])
}

func TestLSHIndex_QueryNoMatch(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)

	idx.Insert("id-1", []float64{1, 0, 0, 0})

	query := []float64{0, 1, 0, 0}
	for len(query) < 256 {
		query = append(query, 0)
	}
	f1 := make([]float64, 256)
	f1[0] = 1.0

	results := idx.Query(f1)
	require.Len(t, results, 1)
	assert.Equal(t, "id-1", results[0])
}

func TestLSHIndex_Remove(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)

	features := make([]float64, 256)
	features[0] = 1.0

	idx.Insert("id-1", features)
	idx.Remove("id-1")

	results := idx.Query(features)
	assert.Empty(t, results)
}

func TestLSHIndex_RemoveNonexistent(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)
	idx.Remove("nonexistent")
}

func TestLSHIndex_Build(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)

	identities := []IndexableIdentity{
		{ID: "id-1", FullName: "Jean Dupont", NNU: "NNU-001"},
		{ID: "id-2", FullName: "Marie Curie", NNU: "NNU-002"},
	}
	idx.Build(identities)

	require.Len(t, idx.signatures, 2)
	require.Contains(t, idx.signatures, "id-1")
	require.Contains(t, idx.signatures, "id-2")
}

func TestLSHIndex_BuildEmpty(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)
	idx.Build(nil)
	assert.Empty(t, idx.signatures)
}

func TestLSHIndex_ConcurrentAccess(t *testing.T) {
	resetCoeffs()
	idx := NewLSHIndex(50, 10, 5)

	done := make(chan bool, 2)
	go func() {
		for i := 0; i < 50; i++ {
			f := make([]float64, 256)
			f[i%256] = 1.0
			idx.Insert("id-1", f)
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 50; i++ {
			f := make([]float64, 256)
			f[(i+10)%256] = 1.0
			idx.Insert("id-2", f)
		}
		done <- true
	}()
	<-done
	<-done

	f := make([]float64, 256)
	f[0] = 1.0
	results := idx.Query(f)
	assert.NotNil(t, results)
}

func TestMinHashSignature_EmptySet(t *testing.T) {
	features := make([]float64, 256)
	sig := minHashSignature(features, 50)
	require.Len(t, sig, 50)
	for _, v := range sig {
		assert.Equal(t, lshPrime, v)
	}
}

func TestMinHashSignature_NonEmpty(t *testing.T) {
	resetCoeffs()
	features := make([]float64, 256)
	features[0] = 1.0
	features[100] = 0.5
	sig := minHashSignature(features, 50)
	require.Len(t, sig, 50)
	for _, v := range sig {
		assert.NotEqual(t, lshPrime, v)
	}
}

func TestHashBand_Deterministic(t *testing.T) {
	h1 := hashBand([]int{1, 2, 3}, 0)
	h2 := hashBand([]int{1, 2, 3}, 0)
	assert.Equal(t, h1, h2)
}

func TestHashBand_DifferentBands(t *testing.T) {
	h1 := hashBand([]int{1, 2, 3}, 0)
	h2 := hashBand([]int{1, 2, 3}, 1)
	assert.NotEqual(t, h1, h2)
}

func TestExtractFeatureIndices(t *testing.T) {
	id := IndexableIdentity{
		ID:        "test-1",
		FullName:  "Jean Dupont",
		FirstName: "Jean",
		LastName:  "Dupont",
		DOB:       "1990-01-01",
		TaxID:     "TAX-001",
		NNU:       "NNU-001",
	}
	indices := extractFeatureIndices(id)
	require.NotEmpty(t, indices)

	seen := make(map[int]bool)
	for _, idx := range indices {
		assert.False(t, seen[idx], "duplicate index %d", idx)
		seen[idx] = true
	}
}

func TestIdentitiesToIndexable(t *testing.T) {
	records := []IdentityStoreRecord{
		&testRecord{
			id:        "id-1",
			nnu:       "NNU-001",
			fullName:  "Jean Dupont",
			firstName: "Jean",
			lastName:  "Dupont",
			dob:       "1990-01-01",
			taxID:     "TAX-001",
		},
	}
	result := IdentitiesToIndexable(records)
	require.Len(t, result, 1)
	assert.Equal(t, "id-1", result[0].ID)
	assert.Equal(t, "NNU-001", result[0].NNU)
	assert.Equal(t, "Jean Dupont", result[0].FullName)
}

type testRecord struct {
	id, nnu, fullName, firstName, lastName, dob, taxID string
}

func (r *testRecord) GetID() string        { return r.id }
func (r *testRecord) GetNNU() string       { return r.nnu }
func (r *testRecord) GetFullName() string  { return r.fullName }
func (r *testRecord) GetFirstName() string { return r.firstName }
func (r *testRecord) GetLastName() string  { return r.lastName }
func (r *testRecord) GetDOB() string       { return r.dob }
func (r *testRecord) GetTaxID() string     { return r.taxID }

func TestGetCoeffs_ThreadSafe(t *testing.T) {
	resetCoeffs()
	done := make(chan bool, 2)
	go func() {
		getCoeffs(100)
		done <- true
	}()
	go func() {
		getCoeffs(100)
		done <- true
	}()
	<-done
	<-done
	a, b := getCoeffs(100)
	require.Len(t, a, 100)
	require.Len(t, b, 100)
}
