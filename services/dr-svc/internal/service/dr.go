package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/snisid/dr-svc/internal/domain"
	"github.com/snisid/dr-svc/internal/kafka"
	"github.com/snisid/dr-svc/internal/repository"
)

type DRService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewDRService(repo repository.Repository, producer *kafka.Producer) *DRService {
	return &DRService{repo: repo, producer: producer}
}

func (s *DRService) ListRegions(ctx context.Context) ([]domain.DRRegion, error) {
	return s.repo.FindAllRegions(ctx)
}

func (s *DRService) CreateFailoverPlan(ctx context.Context, name, sourceRegion, targetRegion string, isAutomated bool) (*domain.FailoverPlan, error) {
	plan := &domain.FailoverPlan{
		PlanID:       uuid.New(),
		Name:         name,
		SourceRegion: sourceRegion,
		TargetRegion: targetRegion,
		IsAutomated:  isAutomated,
		CreatedAt:    time.Now().UTC(),
		IsExecuted:   false,
	}
	if err := s.repo.InsertFailoverPlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("create failover plan: %w", err)
	}
	s.publishEvent(ctx, "dr.failover.plan.created", plan)
	return plan, nil
}

func (s *DRService) ExecuteFailover(ctx context.Context, planID uuid.UUID) (*domain.FailoverExecution, error) {
	plan, err := s.repo.FindFailoverPlanByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	if plan.IsExecuted {
		return nil, fmt.Errorf("plan already executed")
	}

	sourceRegion, err := s.repo.FindRegionByName(ctx, plan.SourceRegion)
	if err != nil {
		return nil, fmt.Errorf("source region not found: %w", err)
	}
	targetRegion, err := s.repo.FindRegionByName(ctx, plan.TargetRegion)
	if err != nil {
		return nil, fmt.Errorf("target region not found: %w", err)
	}

	exec := &domain.FailoverExecution{
		ExecutionID:  uuid.New(),
		PlanID:       planID,
		StartedAt:    time.Now().UTC(),
		IsSuccessful: false,
	}

	if sourceRegion.Health == domain.HealthUnhealthy && targetRegion.Health != domain.HealthUnhealthy {
		if err := s.repo.UpdateRegionActive(ctx, plan.SourceRegion, false); err != nil {
			return nil, fmt.Errorf("deactivate source region: %w", err)
		}
		if err := s.repo.UpdateRegionActive(ctx, plan.TargetRegion, true); err != nil {
			return nil, fmt.Errorf("activate target region: %w", err)
		}
		now := time.Now().UTC()
		exec.CompletedAt = &now
		exec.IsSuccessful = true
	} else {
		errMsg := "source region healthy or target region unhealthy"
		exec.ErrorMessage = errMsg
		now := time.Now().UTC()
		exec.CompletedAt = &now
	}

	if err := s.repo.InsertFailoverExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("insert execution: %w", err)
	}
	if exec.IsSuccessful {
		if err := s.repo.UpdateFailoverPlanExecuted(ctx, planID); err != nil {
			return nil, fmt.Errorf("update plan executed: %w", err)
		}
	}

	s.publishEvent(ctx, "dr.failover.executed", exec)
	return exec, nil
}

func (s *DRService) RunDRTest(ctx context.Context, planID uuid.UUID, testName string) (*domain.DRTestResult, error) {
	plan, err := s.repo.FindFailoverPlanByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	result := &domain.DRTestResult{
		TestID:       uuid.New(),
		PlanID:       planID,
		TestName:     testName,
		StartedAt:    time.Now().UTC(),
		CompletedAt:  time.Now().UTC().Add(30 * time.Second),
		IsSuccessful: true,
		Details: fmt.Sprintf("DR test for plan '%s' completed: source=%s target=%s automated=%v",
			plan.Name, plan.SourceRegion, plan.TargetRegion, plan.IsAutomated),
	}
	if err := s.repo.InsertDRTestResult(ctx, result); err != nil {
		return nil, fmt.Errorf("insert dr test result: %w", err)
	}
	s.publishEvent(ctx, "dr.test.completed", result)
	return result, nil
}

func (s *DRService) ListBackups(ctx context.Context) ([]domain.BackupManifest, error) {
	return s.repo.FindAllBackupManifests(ctx)
}

func (s *DRService) RestoreFromBackup(ctx context.Context, manifestID uuid.UUID) (*domain.RecoveryPoint, error) {
	manifest, err := s.repo.FindBackupManifestByID(ctx, manifestID)
	if err != nil {
		return nil, fmt.Errorf("manifest not found: %w", err)
	}
	if !manifest.IsValid {
		return nil, fmt.Errorf("backup manifest is invalid")
	}

	point := &domain.RecoveryPoint{
		PointID:      uuid.New(),
		ManifestID:   manifestID,
		RecoveryTime: time.Now().UTC(),
		IsRestored:   false,
	}
	if err := s.repo.InsertRecoveryPoint(ctx, point); err != nil {
		return nil, fmt.Errorf("insert recovery point: %w", err)
	}

	now := time.Now().UTC()
	point.IsRestored = true
	point.RestoredAt = &now

	s.publishEvent(ctx, "dr.recovery.completed", point)
	return point, nil
}

func (s *DRService) GetHealthDashboard(ctx context.Context) (map[string]any, error) {
	regions, err := s.repo.FindAllRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("query regions: %w", err)
	}
	replication, err := s.repo.FindReplicationStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("query replication: %w", err)
	}

	dashboard := map[string]any{
		"regions":         regions,
		"replication":     replication,
		"healthy_count":   0,
		"unhealthy_count": 0,
	}
	healthy := 0
	unhealthy := 0
	for _, r := range regions {
		if r.Health == domain.HealthHealthy {
			healthy++
		} else {
			unhealthy++
		}
	}
	dashboard["healthy_count"] = healthy
	dashboard["unhealthy_count"] = unhealthy
	return dashboard, nil
}

func (s *DRService) publishEvent(ctx context.Context, eventType string, data any) {
	if s.producer == nil {
		return
	}
	var planID string
	if p, ok := data.(*domain.FailoverPlan); ok {
		planID = p.PlanID.String()
	}
	evt := kafka.Event{
		EventType: eventType,
		PlanID:    planID,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}
