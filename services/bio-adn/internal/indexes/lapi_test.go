package indexes

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/snisid/platform/services/bio-adn/pkg/models"
)

type mockLAPIDB struct {
	models.Database
	plateMap  map[string]*models.PlateHitResult
	queryTime time.Duration
}

func newMockLAPIDB() *mockLAPIDB {
	return &mockLAPIDB{
		plateMap: make(map[string]*models.PlateHitResult),
	}
}

func (m *mockLAPIDB) QueryPlateIndex(ctx context.Context, plate string) (*models.PlateHitResult, error) {
	time.Sleep(m.queryTime)
	hit, ok := m.plateMap[plate]
	if !ok {
		return nil, nil
	}
	return hit, nil
}

func (m *mockLAPIDB) QueryPlateClones(ctx context.Context, plate string) (int, error) {
	return 0, nil
}

func (m *mockLAPIDB) CreateIdentityLink(ctx context.Context, l *models.BioIdentityLink) error {
	return nil
}
func (m *mockLAPIDB) GetIdentityLinkBySampleID(ctx context.Context, sampleID string) (*models.BioIdentityLink, error) {
	return nil, nil
}
func (m *mockLAPIDB) QueryIdentityLinksByNIU(ctx context.Context, niu string) ([]models.BioIdentityLink, error) {
	return nil, nil
}
func (m *mockLAPIDB) CreateViolenceRecord(ctx context.Context, v *models.ViolenceRecord) error {
	return nil
}
func (m *mockLAPIDB) QueryViolenceRecords(ctx context.Context, niu, incidentType, status string, limit, offset int) ([]models.ViolenceRecord, int, error) {
	return nil, 0, nil
}
func (m *mockLAPIDB) GetViolenceRecordByID(ctx context.Context, id string) (*models.ViolenceRecord, error) {
	return nil, nil
}
func (m *mockLAPIDB) CreateIdentityTheft(ctx context.Context, i *models.IdentityTheft) error {
	return nil
}
func (m *mockLAPIDB) QueryIdentityThefts(ctx context.Context, victimNIU, fraudType, status string, limit, offset int) ([]models.IdentityTheft, int, error) {
	return nil, 0, nil
}
func (m *mockLAPIDB) GetIdentityTheftByID(ctx context.Context, id string) (*models.IdentityTheft, error) {
	return nil, nil
}

func TestLAPICache_HitFound(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	cache := NewLAPICache(mr.Addr())
	ctx := context.Background()

	mr.Set("lapi:plate:AA-1234", `{"hit_found":true,"hit_type":"STOLEN_VEHICLE","record_number":"V-001","alert_level":"HIGH","mco_contact":"PNH-PAP","response_ms":0}`)

	hit, err := cache.GetPlate(ctx, "AA-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit == nil || !hit.HitFound {
		t.Fatal("expected hit found")
	}
	if hit.HitType != "STOLEN_VEHICLE" {
		t.Fatalf("expected STOLEN_VEHICLE, got %s", hit.HitType)
	}
}

func TestLAPICache_CacheMiss(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	cache := NewLAPICache(mr.Addr())
	ctx := context.Background()

	hit, err := cache.GetPlate(ctx, "NONEXISTENT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit != nil {
		t.Fatal("expected nil for cache miss")
	}
}

func TestLAPICache_SetAndGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	cache := NewLAPICache(mr.Addr())
	ctx := context.Background()

	hit := &models.PlateHitResult{
		HitFound:     true,
		HitType:      "STOLEN_VEHICLE",
		RecordNumber: "V-001",
		AlertLevel:   "HIGH",
		MCOContact:   "PNH-PAP",
		ResponseMs:   45,
	}
	err = cache.SetPlateHit(ctx, "AA-1234", hit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cached, err := cache.GetPlate(ctx, "AA-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cached == nil || !cached.HitFound {
		t.Fatal("expected cached hit")
	}
}

func TestLAPICache_SetNoHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	cache := NewLAPICache(mr.Addr())
	ctx := context.Background()

	err = cache.SetPlateNoHit(ctx, "BB-5678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hit, err := cache.GetPlate(ctx, "BB-5678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit == nil || hit.HitFound {
		t.Fatal("expected hit_found=false")
	}
}

func TestLAPICache_TTLExpiry(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	cache := NewLAPICache(mr.Addr())
	ctx := context.Background()

	err = cache.SetPlateNoHit(ctx, "CC-9012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fast-forward 6 minutes (past the 5-min TTL)
	mr.FastForward(6 * time.Minute)

	hit, err := cache.GetPlate(ctx, "CC-9012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit != nil {
		t.Fatal("expected nil after TTL expiry")
	}
}

func TestLAPIQueryService_HitFound(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	mockDB.plateMap["AA-1234"] = &models.PlateHitResult{
		HitFound:     true,
		HitType:      "STOLEN_VEHICLE",
		RecordNumber: "V-001",
		AlertLevel:   "HIGH",
		MCOContact:   "PNH-DELMAS",
	}
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	hit, err := svc.QueryPlate(ctx, "AA-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit == nil || !hit.HitFound {
		t.Fatal("expected hit found")
	}
	if hit.RecordNumber != "V-001" {
		t.Fatalf("expected V-001, got %s", hit.RecordNumber)
	}
}

func TestLAPIQueryService_NoHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	hit, err := svc.QueryPlate(ctx, "ZZ-9999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit == nil {
		t.Fatal("expected result even without hit")
	}
	if hit.HitFound {
		t.Fatal("expected no hit")
	}
}

func TestLAPIQueryService_CacheHitServed(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	mockDB.queryTime = 100 * time.Millisecond // Simulate slow DB
	svc := NewLAPIQueryService(mr.Addr(), mockDB)

	mr.Set("lapi:plate:CACHED-001",
		`{"hit_found":true,"hit_type":"STOLEN_VEHICLE","record_number":"V-CACHE","alert_level":"MEDIUM","mco_contact":"PNH","response_ms":0}`)

	ctx := context.Background()
	hit, err := svc.QueryPlate(ctx, "CACHED-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit == nil || !hit.HitFound {
		t.Fatal("expected cached hit")
	}
	// Cache hit should be fast (< 10ms), not 100ms
	if hit.ResponseMs > 50 {
		t.Fatalf("cache hit should be fast, got %dms (DB has 100ms delay)", hit.ResponseMs)
	}
}

func TestLAPIQueryService_CachedNoHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	mr.Set("lapi:plate:NOHIT-001", `{"hit_found":false}`)
	hit, err := svc.QueryPlate(ctx, "NOHIT-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit.HitFound {
		t.Fatal("expected no hit from cache")
	}
}

func TestLAPIQueryService_DBTimeoutReturnsNoHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	mockDB.queryTime = 200 * time.Millisecond // Exceeds 150ms timeout
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	hit, err := svc.QueryPlate(ctx, "TIMEOUT-PLATE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit.HitFound {
		t.Fatal("expected no hit on DB timeout")
	}
}

func TestLAPIQueryService_VINQuery(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	mockDB.plateMap["VIN-12345"] = &models.PlateHitResult{
		HitFound:     true,
		HitType:      "STOLEN_VEHICLE",
		RecordNumber: "V-VIN-001",
		AlertLevel:   "HIGH",
		MCOContact:   "PNH-PAP",
	}
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	// Since mockLAPIDB doesn't handle VIN queries separately, use the plateMap directly
	// In a real scenario, VIN and plate are different queries
	hit, err := svc.QueryVIN(ctx, "VIN-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit == nil || !hit.HitFound {
		t.Fatal("expected VIN hit")
	}
	t.Logf("VIN query response: %dms", hit.ResponseMs)
}

func TestLAPIQueryService_VINNoMatch(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	hit, err := svc.QueryVIN(ctx, "UNKNOWN-VIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit.HitFound {
		t.Fatal("expected no hit for unknown VIN")
	}
}

func TestPlateSightingRecording(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	cache := NewLAPICache(mr.Addr())
	ctx := context.Background()

	err = cache.RecordPlateSighting(ctx, "AA-1234", "CAM-DELMAS-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sighting, err := cache.GetRecentSighting(ctx, "AA-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sighting == "" {
		t.Fatal("expected non-empty sighting")
	}
}

func TestSLABreachDetection(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mockDB := newMockLAPIDB()
	svc := NewLAPIQueryService(mr.Addr(), mockDB)
	ctx := context.Background()

	mockDB.plateMap["SLA-TEST"] = &models.PlateHitResult{
		HitFound: true, HitType: "STOLEN_VEHICLE", RecordNumber: "V-SLA",
		AlertLevel: "HIGH", MCOContact: "PNH",
	}
	mockDB.queryTime = 300 * time.Millisecond

	hit, err := svc.QueryPlate(ctx, "SLA-TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit.ResponseMs > 200 {
		t.Logf("SLA breach detected: %dms > 200ms SLA", hit.ResponseMs)
		// Verify SLA breach was published to Redis
		breachKey := "lapi:sla:breach:SLA-TEST"
		exists, _ := mr.Exists(breachKey)
		if !exists {
			t.Fatal("expected SLA breach key in Redis")
		}
	}
}
