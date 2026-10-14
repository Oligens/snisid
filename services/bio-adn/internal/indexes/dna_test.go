package indexes

import (
	"context"
	"testing"

	"github.com/snisid/platform/services/bio-adn/pkg/models"
)

type mockDB struct {
	models.Database
	profiles    []models.DNAProfile
	createErr   error
	getErr      error
	searchErr   error
	uploadErr   error
	expungeErr  error
	hashIdx     map[string]int
	specimenIdx map[string]int
}

func newMockDB() *mockDB {
	return &mockDB{
		profiles:    make([]models.DNAProfile, 0),
		hashIdx:     make(map[string]int),
		specimenIdx: make(map[string]int),
	}
}

func (m *mockDB) CreateDNAProfile(ctx context.Context, p *models.DNAProfile) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.hashIdx[p.LociHash] = len(m.profiles)
	m.specimenIdx[p.SpecimenNumber] = len(m.profiles)
	m.profiles = append(m.profiles, *p)
	return nil
}

func (m *mockDB) GetDNAProfileByHash(ctx context.Context, hash string) (*models.DNAProfile, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if i, ok := m.hashIdx[hash]; ok {
		return &m.profiles[i], nil
	}
	return nil, nil
}

func (m *mockDB) GetDNAProfileBySpecimen(ctx context.Context, specimen string) (*models.DNAProfile, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if i, ok := m.specimenIdx[specimen]; ok {
		return &m.profiles[i], nil
	}
	return nil, nil
}

func (m *mockDB) SearchDNAProfiles(ctx context.Context, indexType string, limit, offset int) ([]models.DNAProfile, int, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	var filtered []models.DNAProfile
	for _, p := range m.profiles {
		if p.IndexType == indexType {
			filtered = append(filtered, p)
		}
	}
	return filtered, len(filtered), nil
}

func (m *mockDB) MarkUploaded(ctx context.Context, id, level string) error {
	return m.uploadErr
}

func (m *mockDB) MarkExpunged(ctx context.Context, id string) error {
	return m.expungeErr
}

func (m *mockDB) CreateIdentityLink(ctx context.Context, l *models.BioIdentityLink) error { return nil }
func (m *mockDB) GetIdentityLinkBySampleID(ctx context.Context, sampleID string) (*models.BioIdentityLink, error) {
	return nil, nil
}
func (m *mockDB) QueryIdentityLinksByNIU(ctx context.Context, niu string) ([]models.BioIdentityLink, error) {
	return nil, nil
}
func (m *mockDB) CreateViolenceRecord(ctx context.Context, v *models.ViolenceRecord) error {
	return nil
}
func (m *mockDB) QueryViolenceRecords(ctx context.Context, niu, incidentType, status string, limit, offset int) ([]models.ViolenceRecord, int, error) {
	return nil, 0, nil
}
func (m *mockDB) GetViolenceRecordByID(ctx context.Context, id string) (*models.ViolenceRecord, error) {
	return nil, nil
}
func (m *mockDB) CreateIdentityTheft(ctx context.Context, i *models.IdentityTheft) error { return nil }
func (m *mockDB) QueryIdentityThefts(ctx context.Context, victimNIU, fraudType, status string, limit, offset int) ([]models.IdentityTheft, int, error) {
	return nil, 0, nil
}
func (m *mockDB) GetIdentityTheftByID(ctx context.Context, id string) (*models.IdentityTheft, error) {
	return nil, nil
}

func TestDNAIndex_Create(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	err := idx.Create(context.Background(), &models.DNAProfile{
		SampleID:       "sample-001",
		SpecimenNumber: "FSC-2026-001",
		LociHash:       "abc123",
		IndexType:      "BIO-FSC",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(mock.profiles))
	}
}

func TestDNAIndex_Create_EmptySpecimen(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	err := idx.Create(context.Background(), &models.DNAProfile{
		SampleID:  "sample-002",
		LociHash:  "def456",
		IndexType: "BIO-CON",
	})
	if err == nil {
		t.Fatal("expected error for empty specimen_number")
	}
}

func TestDNAIndex_GetByHash(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	idx.Create(context.Background(), &models.DNAProfile{
		SampleID:       "sample-001",
		SpecimenNumber: "FSC-001",
		LociHash:       "hash001",
		IndexType:      "BIO-FSC",
	})
	got, err := idx.GetByHash(context.Background(), "hash001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.SampleID != "sample-001" {
		t.Fatalf("expected sample-001, got %v", got)
	}
}

func TestDNAIndex_GetByHash_NotFound(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	got, err := idx.GetByHash(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for nonexistent hash")
	}
}

func TestDNAIndex_GetBySpecimen(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	idx.Create(context.Background(), &models.DNAProfile{
		SampleID:       "sample-002",
		SpecimenNumber: "ARR-2026-050",
		LociHash:       "hash002",
		IndexType:      "BIO-ARR",
	})
	got, err := idx.GetBySpecimen(context.Background(), "ARR-2026-050")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.SampleID != "sample-002" {
		t.Fatalf("expected sample-002, got %v", got)
	}
}

func TestDNAIndex_SearchByIndexType(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	idx.Create(context.Background(), &models.DNAProfile{SampleID: "s1", SpecimenNumber: "FSC-1", LociHash: "h1", IndexType: "BIO-FSC"})
	idx.Create(context.Background(), &models.DNAProfile{SampleID: "s2", SpecimenNumber: "CON-1", LociHash: "h2", IndexType: "BIO-CON"})
	idx.Create(context.Background(), &models.DNAProfile{SampleID: "s3", SpecimenNumber: "FSC-2", LociHash: "h3", IndexType: "BIO-FSC"})

	results, total, err := idx.SearchByIndexType(context.Background(), "BIO-FSC", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 BIO-FSC profiles, got %d", total)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestDNAIndex_MarkUploaded(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	err := idx.MarkUploaded(context.Background(), "sample-001", "LDIS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDNAIndex_Expunge(t *testing.T) {
	mock := newMockDB()
	idx := NewDNAIndex(mock)
	err := idx.Expunge(context.Background(), "sample-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
