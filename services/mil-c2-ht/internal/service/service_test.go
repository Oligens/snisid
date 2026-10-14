package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/mil-c2-ht/internal/domain"
	"github.com/snisid/mil-c2-ht/internal/kafka"
)

type mockRepo struct {
	createUnitFn    func(domain.MilitaryUnit) error
	getDeployedFn   func() ([]domain.MilitaryUnit, error)
	createOpFn      func(domain.Operation) error
	getActiveOpFn   func() ([]domain.Operation, error)
	createReportFn  func(domain.TacticalReport) error
	getReportsFn    func(uuid.UUID) ([]domain.TacticalReport, error)
	getAllUnitsFn   func() ([]domain.MilitaryUnit, error)
	getAllOpsFn     func() ([]domain.Operation, error)
	getAllReportsFn func() ([]domain.TacticalReport, error)
}

func (m *mockRepo) CreateUnit(u domain.MilitaryUnit) error             { return m.createUnitFn(u) }
func (m *mockRepo) GetDeployedUnits() ([]domain.MilitaryUnit, error)   { return m.getDeployedFn() }
func (m *mockRepo) CreateOperation(o domain.Operation) error           { return m.createOpFn(o) }
func (m *mockRepo) GetActiveOperations() ([]domain.Operation, error)   { return m.getActiveOpFn() }
func (m *mockRepo) CreateTacticalReport(r domain.TacticalReport) error { return m.createReportFn(r) }
func (m *mockRepo) GetReportsByOperation(id uuid.UUID) ([]domain.TacticalReport, error) {
	return m.getReportsFn(id)
}
func (m *mockRepo) GetAllUnits() ([]domain.MilitaryUnit, error)     { return m.getAllUnitsFn() }
func (m *mockRepo) GetAllOperations() ([]domain.Operation, error)   { return m.getAllOpsFn() }
func (m *mockRepo) GetAllReports() ([]domain.TacticalReport, error) { return m.getAllReportsFn() }

func TestCreateUnit(t *testing.T) {
	repo := &mockRepo{
		createUnitFn: func(u domain.MilitaryUnit) error {
			if u.UnitName == "" {
				return errors.New("empty unit name")
			}
			return nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	u := domain.MilitaryUnit{UnitName: "1st Battalion", Branch: domain.BranchArmy}
	if err := svc.CreateUnit(u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateUnitRepoError(t *testing.T) {
	repo := &mockRepo{
		createUnitFn: func(u domain.MilitaryUnit) error {
			return errors.New("db error")
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	u := domain.MilitaryUnit{UnitName: "1st Battalion", Branch: domain.BranchArmy}
	if err := svc.CreateUnit(u); err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestGetDeployedUnits(t *testing.T) {
	repo := &mockRepo{
		getDeployedFn: func() ([]domain.MilitaryUnit, error) {
			return []domain.MilitaryUnit{{UnitName: "Recon Unit"}}, nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	units, err := svc.GetDeployedUnits()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(units))
	}
}

func TestCreateOperation(t *testing.T) {
	repo := &mockRepo{
		createOpFn: func(o domain.Operation) error {
			if o.OperationName == "" {
				return errors.New("empty operation name")
			}
			return nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	op := domain.Operation{OperationName: "Op Guardian", OperationType: domain.OpTypeSecurity}
	if err := svc.CreateOperation(op); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetActiveOperations(t *testing.T) {
	repo := &mockRepo{
		getActiveOpFn: func() ([]domain.Operation, error) {
			return []domain.Operation{{OperationName: "Op Guardian"}}, nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	ops, err := svc.GetActiveOperations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}
}

func TestSubmitReport(t *testing.T) {
	repo := &mockRepo{
		createReportFn: func(r domain.TacticalReport) error {
			if r.ReportType == "" {
				return errors.New("empty report type")
			}
			return nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	report := domain.TacticalReport{ReportType: domain.ReportSITREP}
	if err := svc.SubmitReport(uuid.New(), report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubmitReportRepoError(t *testing.T) {
	repo := &mockRepo{
		createReportFn: func(r domain.TacticalReport) error {
			return errors.New("db error")
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	report := domain.TacticalReport{ReportType: domain.ReportSITREP}
	if err := svc.SubmitReport(uuid.New(), report); err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestGetOperationTimeline(t *testing.T) {
	opID := uuid.New()
	repo := &mockRepo{
		getReportsFn: func(id uuid.UUID) ([]domain.TacticalReport, error) {
			return []domain.TacticalReport{{OperationID: id}}, nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	reports, err := svc.GetOperationTimeline(opID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
}

func TestGetCommonOperatingPicture(t *testing.T) {
	repo := &mockRepo{
		getAllUnitsFn: func() ([]domain.MilitaryUnit, error) {
			return []domain.MilitaryUnit{{UnitName: "Unit A"}}, nil
		},
		getAllOpsFn: func() ([]domain.Operation, error) {
			return []domain.Operation{{OperationName: "Op A"}}, nil
		},
		getAllReportsFn: func() ([]domain.TacticalReport, error) {
			return []domain.TacticalReport{{ReportType: domain.ReportSITREP}}, nil
		},
	}
	svc := NewMilC2Service(repo, &kafka.Producer{})
	cop, err := svc.GetCommonOperatingPicture()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cop.Units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(cop.Units))
	}
	if len(cop.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(cop.Operations))
	}
	if len(cop.Reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(cop.Reports))
	}
}
