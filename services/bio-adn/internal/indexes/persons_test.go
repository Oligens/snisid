package indexes

import (
	"context"
	"testing"

	"github.com/snisid/platform/services/bio-adn/pkg/models"
)

type mockPersonsDB struct {
	models.Database
	wanted    []models.WantedPerson
	createErr error
	queryErr  error
	getErr    error
	updateErr error
}

func newMockPersonsDB() *mockPersonsDB {
	return &mockPersonsDB{
		wanted: make([]models.WantedPerson, 0),
	}
}

func (m *mockPersonsDB) CreateWantedPerson(ctx context.Context, p *models.WantedPerson) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.wanted = append(m.wanted, *p)
	return nil
}

func (m *mockPersonsDB) QueryWantedPersons(ctx context.Context, q *models.WantedQuery) ([]models.WantedPerson, int, error) {
	if m.queryErr != nil {
		return nil, 0, m.queryErr
	}
	var filtered []models.WantedPerson
	for _, p := range m.wanted {
		if q.LastName != "" && p.LastName != q.LastName {
			continue
		}
		if q.NIU != "" && p.NIU != q.NIU {
			continue
		}
		if q.Status != "" && p.Status != q.Status {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered, len(filtered), nil
}

func (m *mockPersonsDB) GetWantedByID(ctx context.Context, id string) (*models.WantedPerson, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, p := range m.wanted {
		if p.RecordID == id {
			return &p, nil
		}
	}
	return nil, nil
}

func (m *mockPersonsDB) UpdateWantedStatus(ctx context.Context, id, status string) error {
	return m.updateErr
}

func (m *mockPersonsDB) CreateForeignFugitive(ctx context.Context, f *models.ForeignFugitive) error {
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}

func (m *mockPersonsDB) QueryForeignFugitives(ctx context.Context, lastName, nationality, noticeType string, limit, offset int) ([]models.ForeignFugitive, int, error) {
	return nil, 0, nil
}

func (m *mockPersonsDB) GetForeignFugitiveByID(ctx context.Context, id string) (*models.ForeignFugitive, error) {
	return nil, nil
}

func (m *mockPersonsDB) CreateUnidentifiedPerson(ctx context.Context, u *models.UnidentifiedPerson) error {
	return nil
}

func (m *mockPersonsDB) QueryUnidentifiedPersons(ctx context.Context, dept, gender string, ageMin, ageMax, limit, offset int) ([]models.UnidentifiedPerson, int, error) {
	return nil, 0, nil
}

func (m *mockPersonsDB) GetUnidentifiedByID(ctx context.Context, id string) (*models.UnidentifiedPerson, error) {
	return nil, nil
}

func (m *mockPersonsDB) CreateTerrorismWatch(ctx context.Context, t *models.TerrorismWatch) error {
	return nil
}

func (m *mockPersonsDB) QueryTerrorismWatches(ctx context.Context, riskLevel, threatType, nationality string, limit, offset int) ([]models.TerrorismWatch, int, error) {
	return nil, 0, nil
}

func (m *mockPersonsDB) GetTerrorismWatchByID(ctx context.Context, id string) (*models.TerrorismWatch, error) {
	return nil, nil
}

func (m *mockPersonsDB) CreateProtectionOrder(ctx context.Context, p *models.ProtectionOrder) error {
	return nil
}

func (m *mockPersonsDB) QueryProtectionOrders(ctx context.Context, beneficiaryName, restrainedPerson, orderType string, limit, offset int) ([]models.ProtectionOrder, int, error) {
	return nil, 0, nil
}

func (m *mockPersonsDB) GetActiveProtectionOrdersByBeneficiary(ctx context.Context, beneficiaryNIU string) ([]models.ProtectionOrder, error) {
	return nil, nil
}

func (m *mockPersonsDB) CreateSupervisedRelease(ctx context.Context, s *models.SupervisedRelease) error {
	return nil
}

func (m *mockPersonsDB) QuerySupervisedReleases(ctx context.Context, niu, supervisionType, status string, limit, offset int) ([]models.SupervisedRelease, int, error) {
	return nil, 0, nil
}

func (m *mockPersonsDB) GetSupervisedReleaseByID(ctx context.Context, id string) (*models.SupervisedRelease, error) {
	return nil, nil
}

func (m *mockPersonsDB) UpdateSexOffenderRisk(ctx context.Context, id, riskLevel, address string) error {
	return nil
}

func (m *mockPersonsDB) RecordGangMemberReview(ctx context.Context, id string) error {
	return nil
}

func (m *mockPersonsDB) CreateLabEquipment(ctx context.Context, e *models.LabEquipment) error {
	return nil
}
func (m *mockPersonsDB) QueryLabEquipment(ctx context.Context, labCode string) ([]models.LabEquipment, error) {
	return nil, nil
}
func (m *mockPersonsDB) GetLabEquipmentByID(ctx context.Context, id string) (*models.LabEquipment, error) {
	return nil, nil
}
func (m *mockPersonsDB) UpdateEquipmentCalibration(ctx context.Context, id, calibrationDate, calibrationDue, status string) error {
	return nil
}
func (m *mockPersonsDB) CreateStaffTraining(ctx context.Context, t *models.StaffTraining) error {
	return nil
}
func (m *mockPersonsDB) QueryStaffTraining(ctx context.Context, staffNIU string) ([]models.StaffTraining, error) {
	return nil, nil
}
func (m *mockPersonsDB) GetStaffTrainingByID(ctx context.Context, id string) (*models.StaffTraining, error) {
	return nil, nil
}
func (m *mockPersonsDB) CheckDuplicateSpecimen(ctx context.Context, specimen string) (bool, error) {
	return false, nil
}
func (m *mockPersonsDB) MarkSpecimenSubmitted(ctx context.Context, specimen, sampleID string) error {
	return nil
}
func (m *mockPersonsDB) RecordCrossDeptHit(ctx context.Context, h *models.NdisCrossDeptHit) error {
	return nil
}
func (m *mockPersonsDB) QueryCrossDeptHits(ctx context.Context, sdis, matchType string, limit, offset int) ([]models.NdisCrossDeptHit, int, error) {
	return nil, 0, nil
}
func (m *mockPersonsDB) GetNdisStats(ctx context.Context) (*models.NdisStats, error) {
	return &models.NdisStats{}, nil
}
func (m *mockPersonsDB) CreateNdisReport(ctx context.Context, r *models.NdisReport) error { return nil }
func (m *mockPersonsDB) QueryNdisReports(ctx context.Context) ([]models.NdisReport, error) {
	return nil, nil
}
func (m *mockPersonsDB) CreateInterpolSubmission(ctx context.Context, s *models.InterpolSubmission) error {
	return nil
}
func (m *mockPersonsDB) CountInterpolSubmissionsThisWeek(ctx context.Context) (int, error) {
	return 0, nil
}
func (m *mockPersonsDB) CreateIdentityLink(ctx context.Context, l *models.BioIdentityLink) error {
	return nil
}
func (m *mockPersonsDB) GetIdentityLinkBySampleID(ctx context.Context, sampleID string) (*models.BioIdentityLink, error) {
	return nil, nil
}
func (m *mockPersonsDB) QueryIdentityLinksByNIU(ctx context.Context, niu string) ([]models.BioIdentityLink, error) {
	return nil, nil
}
func (m *mockPersonsDB) CreateViolenceRecord(ctx context.Context, v *models.ViolenceRecord) error {
	return nil
}
func (m *mockPersonsDB) QueryViolenceRecords(ctx context.Context, niu, incidentType, status string, limit, offset int) ([]models.ViolenceRecord, int, error) {
	return nil, 0, nil
}
func (m *mockPersonsDB) GetViolenceRecordByID(ctx context.Context, id string) (*models.ViolenceRecord, error) {
	return nil, nil
}
func (m *mockPersonsDB) CreateIdentityTheft(ctx context.Context, i *models.IdentityTheft) error {
	return nil
}
func (m *mockPersonsDB) QueryIdentityThefts(ctx context.Context, victimNIU, fraudType, status string, limit, offset int) ([]models.IdentityTheft, int, error) {
	return nil, 0, nil
}
func (m *mockPersonsDB) GetIdentityTheftByID(ctx context.Context, id string) (*models.IdentityTheft, error) {
	return nil, nil
}

func TestPersonsIndex_CreateWanted(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID:       "WP-001",
		LastName:       "Pierre",
		FirstName:      "Jean",
		WarrantType:    "ARREST",
		Charges:        []string{"Vol à main armée"},
		DangerLevel:    "HIGH",
		EnteringAgency: "PNH-DELMAS",
		Status:         "ACTIVE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.wanted) != 1 {
		t.Fatalf("expected 1 wanted person, got %d", len(mock.wanted))
	}
}

func TestPersonsIndex_CreateWanted_EmptyWarrantType(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID: "WP-002",
		LastName: "Dupont",
		Charges:  []string{"Fraude"},
	})
	if err == nil {
		t.Fatal("expected error for empty warrant_type")
	}
}

func TestPersonsIndex_CreateWanted_NoCharges(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID:    "WP-003",
		WarrantType: "ARREST",
		Charges:     []string{},
	})
	if err == nil {
		t.Fatal("expected error for empty charges")
	}
}

func TestPersonsIndex_QueryWanted(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID: "WP-001", LastName: "Pierre", WarrantType: "ARREST",
		Charges: []string{"Vol"}, DangerLevel: "HIGH", EnteringAgency: "PNH", Status: "ACTIVE",
	})
	idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID: "WP-002", LastName: "Pierre", WarrantType: "ARREST",
		Charges: []string{"Meurtre"}, DangerLevel: "CRITICAL", EnteringAgency: "PNH", Status: "ACTIVE",
	})

	results, total, err := idx.QueryWanted(context.Background(), &models.WantedQuery{LastName: "Pierre", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 results, got %d", total)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestPersonsIndex_QueryWanted_DefaultLimit(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	for i := 0; i < 30; i++ {
		idx.CreateWanted(context.Background(), &models.WantedPerson{
			RecordID: "WP", LastName: "Test", WarrantType: "ARREST",
			Charges: []string{"Test"}, DangerLevel: "LOW", EnteringAgency: "PNH", Status: "ACTIVE",
		})
	}
	results, total, err := idx.QueryWanted(context.Background(), &models.WantedQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 30 {
		t.Fatalf("expected 30 total, got %d", total)
	}
	if len(results) > 20 {
		t.Fatalf("expected default limit 20, got %d", len(results))
	}
}

func TestPersonsIndex_GetWanted(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID: "WP-001", LastName: "Pierre", WarrantType: "ARREST",
		Charges: []string{"Vol"}, DangerLevel: "HIGH", EnteringAgency: "PNH", Status: "ACTIVE",
	})
	got, err := idx.GetWanted(context.Background(), "WP-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.LastName != "Pierre" {
		t.Fatalf("expected Pierre, got %v", got)
	}
}

func TestPersonsIndex_GetWanted_NotFound(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetWanted(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for nonexistent")
	}
}

func TestPersonsIndex_UpdateStatus(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.UpdateStatus(context.Background(), "WP-001", "CLEARED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_UpdateStatus_Invalid(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.UpdateStatus(context.Background(), "WP-001", "INVALID")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestPersonsIndex_CreateWanted_RequiresWarrantNumberForArrest(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateWanted(context.Background(), &models.WantedPerson{
		RecordID: "WP-004", WarrantType: "MAN-ARR", Charges: []string{"Test"},
		DangerLevel: "HIGH", EnteringAgency: "PNH", Status: "ACTIVE",
	})
	if err == nil {
		t.Fatal("expected error for missing warrant_number on MAN-ARR")
	}
}

func TestPersonsIndex_CreateForeignFugitive(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateForeignFugitive(context.Background(), &models.ForeignFugitive{
		InterpolNoticeNumber: "RED-2026-001",
		LastName:             "Garcia", IssuingCountry: "COL",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_CreateForeignFugitive_MissingNotice(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateForeignFugitive(context.Background(), &models.ForeignFugitive{
		LastName: "Garcia", IssuingCountry: "COL",
	})
	if err == nil {
		t.Fatal("expected error for missing interpol_notice_number")
	}
}

func TestPersonsIndex_QueryForeignFugitives(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QueryForeignFugitives(context.Background(), "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 total")
	}
	_ = results
}

func TestPersonsIndex_GetForeignFugitive(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetForeignFugitive(context.Background(), "FUG-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestPersonsIndex_CreateUnidentifiedPerson(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateUnidentifiedPerson(context.Background(), &models.UnidentifiedPerson{
		DiscoveryDate: "2026-06-01", DiscoveryLocation: "Rue Centre",
		EnteringAgency: "PNH",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_QueryUnidentifiedPersons(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QueryUnidentifiedPersons(context.Background(), "", "", 0, 0, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = results
	_ = total
}

func TestPersonsIndex_GetUnidentifiedPerson(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetUnidentifiedPerson(context.Background(), "NID-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestPersonsIndex_CreateTerrorismWatch(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateTerrorismWatch(context.Background(), &models.TerrorismWatch{
		LastName: "Mohamed", ThreatType: "RADICALISATION",
		EnteringAgency: "DCPJ", ApprovedByDirector: "DIR", ApprovedByPG: "PG",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_CreateTerrorismWatch_RequiresApproval(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateTerrorismWatch(context.Background(), &models.TerrorismWatch{
		LastName: "Mohamed", ThreatType: "RADICALISATION",
		EnteringAgency: "DCPJ",
	})
	if err == nil {
		t.Fatal("expected error for missing dual approval")
	}
}

func TestPersonsIndex_QueryTerrorismWatches(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QueryTerrorismWatches(context.Background(), "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = results
	_ = total
}

func TestPersonsIndex_GetTerrorismWatch(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetTerrorismWatch(context.Background(), "TER-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestPersonsIndex_CreateProtectionOrder(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateProtectionOrder(context.Background(), &models.ProtectionOrder{
		OrderType: "RESTRAINING", IssuingCourt: "TPP", IssuingJudge: "Juge X",
		BeneficiaryName: "Marie", ProtectedPerson: "Marie", RestrainedPerson: "Pierre",
		Restrictions: []string{"distance 500m"}, IssueDate: "2026-06-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_QueryProtectionOrders(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QueryProtectionOrders(context.Background(), "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = results
	_ = total
}

func TestPersonsIndex_GetActiveProtectionOrders(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	orders, err := idx.GetActiveProtectionOrders(context.Background(), "NIU-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orders == nil {
		t.Fatal("expected empty slice, not nil")
	}
}

func TestPersonsIndex_CreateSupervisedRelease(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateSupervisedRelease(context.Background(), &models.SupervisedRelease{
		NIU: "NIU-123", LastName: "Pierre", SupervisionType: "PAROLE",
		StartDate: "2026-06-01", Conditions: []string{"pointage"},
		SupervisingOfficer: "OFF X", SupervisingAgency: "SPA",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_QuerySupervisedReleases(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QuerySupervisedReleases(context.Background(), "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = results
	_ = total
}

func TestPersonsIndex_GetSupervisedRelease(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetSupervisedRelease(context.Background(), "LIB-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestPersonsIndex_CreateViolenceRecord(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateViolenceRecord(context.Background(), &models.ViolenceRecord{
		RecordID: "VIO-001", RecordNumber: "VIO-2026-000001",
		IncidentType: "DOMESTIC_VIOLENCE", IncidentDate: "2026-06-01",
		Location: "Delmas 33", ArrestingAgency: "PNH-DELMAS", RiskLevel: "HIGH",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_QueryViolenceRecords(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QueryViolenceRecords(context.Background(), "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = results
	_ = total
}

func TestPersonsIndex_GetViolenceRecord(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetViolenceRecord(context.Background(), "VIO-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestPersonsIndex_CreateIdentityTheft(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	err := idx.CreateIdentityTheft(context.Background(), &models.IdentityTheft{
		RecordID: "IDV-001", RecordNumber: "IDV-2026-000001",
		VictimNIU: "NIU-IDV-001", FraudType: "CIN_FRAUD",
		ReportDate: "2026-06-01", ReportingAgency: "PNH-DELMAS",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersonsIndex_QueryIdentityThefts(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	results, total, err := idx.QueryIdentityThefts(context.Background(), "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = results
	_ = total
}

func TestPersonsIndex_GetIdentityTheft(t *testing.T) {
	mock := newMockPersonsDB()
	idx := NewPersonsIndex(mock)
	got, err := idx.GetIdentityTheft(context.Background(), "IDV-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}
