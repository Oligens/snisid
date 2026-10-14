package sync

import (
	"context"
	"testing"
	"time"
)

type mockUploadStore struct {
	profiles   []map[string]any
	getErr     error
	markErr    error
	getCalled  int
	markCalled int
}

func (m *mockUploadStore) GetUnuploadedProfiles(ctx context.Context, level string) ([]map[string]any, error) {
	m.getCalled++
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.profiles, nil
}

func (m *mockUploadStore) MarkUploaded(ctx context.Context, id, level string) error {
	m.markCalled++
	return m.markErr
}

type mockEventProducer struct {
	createdCalled  int
	uploadedCalled int
}

func (m *mockEventProducer) PublishProfileCreated(ctx context.Context, event any) error {
	m.createdCalled++
	return nil
}

func (m *mockEventProducer) PublishProfileUploaded(ctx context.Context, event any) error {
	m.uploadedCalled++
	return nil
}

func TestLDISUploader_Upload(t *testing.T) {
	store := &mockUploadStore{
		profiles: []map[string]any{
			{"id": "p1", "specimen_number": "FSC-001"},
			{"id": "p2", "specimen_number": "FSC-002"},
		},
	}
	producer := &mockEventProducer{}
	uploader := &LDISUploader{store: store, producer: producer}

	err := uploader.Upload(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.getCalled != 1 {
		t.Fatalf("expected 1 GetUnuploadedProfiles call, got %d", store.getCalled)
	}
	if producer.createdCalled != 2 {
		t.Fatalf("expected 2 PublishProfileCreated calls, got %d", producer.createdCalled)
	}
	if store.markCalled != 2 {
		t.Fatalf("expected 2 MarkUploaded calls, got %d", store.markCalled)
	}
}

func TestLDISUploader_Upload_Empty(t *testing.T) {
	store := &mockUploadStore{profiles: []map[string]any{}}
	producer := &mockEventProducer{}
	uploader := &LDISUploader{store: store, producer: producer}

	err := uploader.Upload(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if producer.createdCalled != 0 {
		t.Fatalf("expected 0 publishes for empty upload, got %d", producer.createdCalled)
	}
}

func TestSDISUploader_Upload(t *testing.T) {
	store := &mockUploadStore{
		profiles: []map[string]any{
			{"id": "p1", "specimen_number": "CON-001"},
		},
	}
	producer := &mockEventProducer{}
	uploader := &SDISUploader{store: store, producer: producer}

	err := uploader.Upload(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if producer.uploadedCalled != 1 {
		t.Fatalf("expected 1 PublishProfileUploaded call, got %d", producer.uploadedCalled)
	}
}

func TestNextRun_Daily_0200_Before(t *testing.T) {
	s := &SyncScheduler{}
	now := time.Date(2026, 6, 11, 1, 0, 0, 0, time.UTC)
	next := s.nextRunAt("02:00", now)
	if next.Hour() != 2 || next.Day() != 11 {
		t.Fatalf("expected 02:00 today, got %v", next)
	}
}

func TestNextRun_Daily_0200_After(t *testing.T) {
	s := &SyncScheduler{}
	now := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	next := s.nextRunAt("02:00", now)
	if next.Day() != 12 {
		t.Fatalf("expected tomorrow, got %v (day %d)", next, next.Day())
	}
	if next.Hour() != 2 {
		t.Fatalf("expected 02:00, got %d:00", next.Hour())
	}
}

func TestNextRun_Daily_0200_Exact(t *testing.T) {
	s := &SyncScheduler{}
	// At exactly 02:00, should run now (no delay)
	now := time.Date(2026, 6, 11, 2, 0, 0, 0, time.UTC)
	next := s.nextRunAt("02:00", now)
	if !next.Equal(now) {
		t.Fatalf("expected exact time now, got %v", next)
	}
}

func TestNextRun_Weekly_Sunday_0300_Before(t *testing.T) {
	// Wednesday June 10 2026
	s := &SyncScheduler{}
	now := time.Date(2026, 6, 10, 1, 0, 0, 0, time.UTC)
	next := s.nextRunAt("sunday-03:00", now)
	if next.Weekday() != time.Sunday {
		t.Fatalf("expected Sunday, got %s", next.Weekday())
	}
	if next.Hour() != 3 {
		t.Fatalf("expected 03:00, got %d:00", next.Hour())
	}
}

func TestNextRun_Weekly_Sunday_0300_OnSundayBefore(t *testing.T) {
	// Sunday at 01:00 → should run at 03:00 same day
	s := &SyncScheduler{}
	now := time.Date(2026, 6, 14, 1, 0, 0, 0, time.UTC)
	next := s.nextRunAt("sunday-03:00", now)
	if next.Day() != 14 {
		t.Fatalf("expected same day (Sunday), got day %d", next.Day())
	}
	if next.Hour() != 3 {
		t.Fatalf("expected 03:00, got %d:00", next.Hour())
	}
}

func TestNextRun_Weekly_Sunday_0300_OnSundayAfter(t *testing.T) {
	// Sunday at 10:00 → next Sunday
	s := &SyncScheduler{}
	now := time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC)
	next := s.nextRunAt("sunday-03:00", now)
	if next.Day() != 21 {
		t.Fatalf("expected next Sunday (day 21), got day %d", next.Day())
	}
}

func TestNextRun_Default(t *testing.T) {
	s := &SyncScheduler{}
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	next := s.nextRunAt("unknown-schedule", now)
	if !next.After(now) {
		t.Fatal("expected next to be in the future for unknown schedule")
	}
}

func TestSyncScheduler_StartLDIS(t *testing.T) {
	store := &mockUploadStore{profiles: []map[string]any{}}
	producer := &mockEventProducer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediate cancel to test clean shutdown

	s := NewSyncScheduler("LDIS", store, producer)
	s.Start(ctx) // Should return immediately without panic
}

func TestSyncScheduler_StartSDIS(t *testing.T) {
	store := &mockUploadStore{profiles: []map[string]any{}}
	producer := &mockEventProducer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := NewSyncScheduler("SDIS", store, producer)
	s.Start(ctx)
}

func TestSyncScheduler_StartUnknownLevel(t *testing.T) {
	store := &mockUploadStore{profiles: []map[string]any{}}
	producer := &mockEventProducer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := NewSyncScheduler("NDIS", store, producer)
	s.Start(ctx) // Should be a no-op
}

// Helper: nextRunAt allows passing a custom time for deterministic testing
func (s *SyncScheduler) nextRunAt(schedule string, now time.Time) time.Time {
	loc, _ := time.LoadLocation("America/Port-au-Prince")
	if loc == nil {
		loc = time.FixedZone("HT", -4*60*60)
	}
	nowLocal := now.In(loc)

	switch {
	case len(schedule) >= 5 && schedule[0] >= '0' && schedule[0] <= '2':
		hour, min := 2, 0
		for i := 0; i < len(schedule) && schedule[i] != ':'; i++ {
		}
		if len(schedule) >= 5 {
			hour = int(schedule[0]-'0')*10 + int(schedule[1]-'0')
			if len(schedule) >= 8 {
				min = int(schedule[3]-'0')*10 + int(schedule[4]-'0')
			}
		}
		next := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), hour, min, 0, 0, loc)
		if !next.After(nowLocal) {
			next = next.Add(24 * time.Hour)
		}
		return next

	case schedule == "sunday-03:00":
		daysUntilSunday := (7 - int(nowLocal.Weekday())) % 7
		if daysUntilSunday == 0 && nowLocal.Hour() >= 3 {
			daysUntilSunday = 7
		}
		next := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day()+daysUntilSunday, 3, 0, 0, 0, loc)
		return next

	default:
		return now.Add(24 * time.Hour)
	}
}
