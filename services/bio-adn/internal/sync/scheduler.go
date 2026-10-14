package sync

import (
	"context"
	"fmt"
	"log"
	"time"
)

type LDISUploader struct {
	store    UploadStore
	producer EventProducer
}

type SDISUploader struct {
	store    UploadStore
	producer EventProducer
}

type UploadStore interface {
	GetUnuploadedProfiles(ctx context.Context, level string) ([]map[string]any, error)
	MarkUploaded(ctx context.Context, id, level string) error
}

type EventProducer interface {
	PublishProfileCreated(ctx context.Context, event any) error
	PublishProfileUploaded(ctx context.Context, event any) error
}

type SyncScheduler struct {
	ldisUploader *LDISUploader
	sdisUploader *SDISUploader
	level        string
}

func NewSyncScheduler(level string, store UploadStore, producer EventProducer) *SyncScheduler {
	return &SyncScheduler{
		level:        level,
		ldisUploader: &LDISUploader{store: store, producer: producer},
		sdisUploader: &SDISUploader{store: store, producer: producer},
	}
}

func (s *SyncScheduler) Start(ctx context.Context) {
	switch s.level {
	case "LDIS":
		go s.runSchedule(ctx, "LDIS->SDIS", "02:00", s.ldisUploader.Upload)
	case "SDIS":
		go s.runSchedule(ctx, "SDIS->NDIS", "sunday-03:00", s.sdisUploader.Upload)
	}
}

func (s *SyncScheduler) runSchedule(ctx context.Context, name, schedule string, fn func(context.Context) error) {
	for {
		next := s.nextRun(schedule)
		select {
		case <-ctx.Done():
			log.Printf("[BIO-SYNC] Arrêt %s", name)
			return
		case <-time.After(time.Until(next)):
			log.Printf("[BIO-SYNC] Démarrage synchronisation %s", name)
			if err := fn(ctx); err != nil {
				log.Printf("[BIO-SYNC] ERREUR %s: %v", name, err)
			} else {
				log.Printf("[BIO-SYNC] Succès %s", name)
			}
		}
	}
}

func (s *SyncScheduler) nextRun(schedule string) time.Time {
	now := time.Now()
	loc, _ := time.LoadLocation("America/Port-au-Prince")
	if loc == nil {
		loc = time.FixedZone("HT", -4*60*60)
	}
	nowLocal := now.In(loc)

	switch {
	case len(schedule) >= 5 && schedule[0] >= '0' && schedule[0] <= '2':
		hour, min := 2, 0
		fmt.Sscanf(schedule, "%d:%d", &hour, &min)
		next := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), hour, min, 0, 0, loc)
		if !next.After(nowLocal) {
			next = next.Add(24 * time.Hour)
		}
		return next

	case schedule == "sunday-03:00":
		daysUntilSunday := (7 - int(nowLocal.Weekday())) % 7
		if daysUntilSunday == 0 && nowLocal.Hour() >= 3 {
			daysUntilSunday = 7
		} else if daysUntilSunday == 0 {
			daysUntilSunday = 0
		}
		next := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day()+daysUntilSunday, 3, 0, 0, 0, loc)
		return next

	default:
		return now.Add(24 * time.Hour)
	}
}

func (u *LDISUploader) Upload(ctx context.Context) error {
	profiles, err := u.store.GetUnuploadedProfiles(ctx, "LDIS")
	if err != nil {
		return fmt.Errorf("get ldis profiles: %w", err)
	}
	for _, p := range profiles {
		u.producer.PublishProfileCreated(ctx, p)
		u.store.MarkUploaded(ctx, p["id"].(string), "LDIS")
	}
	log.Printf("[BIO-SYNC] LDIS->SDIS: %d profils uploadés", len(profiles))
	return nil
}

func (u *SDISUploader) Upload(ctx context.Context) error {
	profiles, err := u.store.GetUnuploadedProfiles(ctx, "SDIS")
	if err != nil {
		return fmt.Errorf("get sdis profiles: %w", err)
	}
	for _, p := range profiles {
		u.producer.PublishProfileUploaded(ctx, p)
		u.store.MarkUploaded(ctx, p["id"].(string), "SDIS")
	}
	log.Printf("[BIO-SYNC] SDIS->NDIS: %d profils uploadés", len(profiles))
	return nil
}
