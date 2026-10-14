package indexes

import (
	"context"
	"testing"

	"github.com/snisid/platform/services/bio-adn/pkg/models"
)

type mockPropertyDB struct {
	models.Database
	vehicles   []models.StolenVehicle
	firearms   []models.StolenFirearm
	docs       []models.StolenDocument
	vessels    []models.StolenVessel
	articles   []models.StolenArticle
	securities []models.StolenSecurity
}

func newMockPropertyDB() *mockPropertyDB {
	return &mockPropertyDB{
		vehicles:   make([]models.StolenVehicle, 0),
		firearms:   make([]models.StolenFirearm, 0),
		docs:       make([]models.StolenDocument, 0),
		vessels:    make([]models.StolenVessel, 0),
		articles:   make([]models.StolenArticle, 0),
		securities: make([]models.StolenSecurity, 0),
	}
}

func (m *mockPropertyDB) CreateStolenVehicle(ctx context.Context, v *models.StolenVehicle) error {
	m.vehicles = append(m.vehicles, *v)
	return nil
}
func (m *mockPropertyDB) UpdateVehicleStatus(ctx context.Context, id, status, location, agency string) error {
	for i, v := range m.vehicles {
		if v.RecordID == id {
			m.vehicles[i].Status = status
			return nil
		}
	}
	return nil
}
func (m *mockPropertyDB) CreateStolenFirearm(ctx context.Context, f *models.StolenFirearm) error {
	m.firearms = append(m.firearms, *f)
	return nil
}
func (m *mockPropertyDB) CreateStolenDocument(ctx context.Context, d *models.StolenDocument) error {
	m.docs = append(m.docs, *d)
	return nil
}
func (m *mockPropertyDB) CreateStolenVessel(ctx context.Context, v *models.StolenVessel) error {
	m.vessels = append(m.vessels, *v)
	return nil
}
func (m *mockPropertyDB) CreateStolenArticle(ctx context.Context, a *models.StolenArticle) error {
	m.articles = append(m.articles, *a)
	return nil
}
func (m *mockPropertyDB) CreateStolenSecurity(ctx context.Context, s *models.StolenSecurity) error {
	m.securities = append(m.securities, *s)
	return nil
}

func TestPropertyIndex_CreateVehicle(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	err := idx.CreateVehicle(context.Background(), &models.StolenVehicle{
		RecordID: "V-001", PlateNumber: "AA-1234", VehicleMake: "Toyota",
		VehicleModel: "Hilux", VehicleYear: 2020, TheftDate: "2026-06-01",
		TheftLocation: "Delmas", EnteringAgency: "PNH-DELMAS", Status: "STOLEN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.vehicles) != 1 {
		t.Fatalf("expected 1 vehicle, got %d", len(mock.vehicles))
	}
}

func TestPropertyIndex_CreateFirearm(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	err := idx.CreateFirearm(context.Background(), &models.StolenFirearm{
		RecordID: "F-001", SerialNumber: "SN-12345", Make: "Glock",
		Model: "17", Caliber: "9mm", TheftDate: "2026-05-15",
		EnteringAgency: "PNH-PAP", Status: "STOLEN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.firearms) != 1 {
		t.Fatalf("expected 1 firearm, got %d", len(mock.firearms))
	}
}

func TestPropertyIndex_CreateDocument(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	err := idx.CreateDocument(context.Background(), &models.StolenDocument{
		RecordID: "D-001", DocumentType: "PASSPORT", DocumentNumber: "HT-123456",
		ReportDate: "2026-06-10", TheftType: "STOLEN",
		OwnerNIU: "NIU-001", Status: "STOLEN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(mock.docs))
	}
}

func TestPropertyIndex_CreateVessel(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	err := idx.CreateVessel(context.Background(), &models.StolenVessel{
		RecordID: "VS-001", VesselName: "Sea Spirit",
		RegistrationNumber: "REG-001", TheftLocation: "Port-au-Prince",
		TheftDate: "2026-04-20", Status: "STOLEN",
		HullIDNumber: "HIN-ABC-12345", HomePort: "Port-au-Prince",
		EngineSerial: "ENG-98765", DistinctiveMarks: "Blue hull, white stripe",
		HullColor: "Blue", VesselLengthM: 12.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.vessels) != 1 {
		t.Fatalf("expected 1 vessel, got %d", len(mock.vessels))
	}
	created := mock.vessels[0]
	if created.HomePort != "Port-au-Prince" {
		t.Fatalf("expected HomePort, got %s", created.HomePort)
	}
	if created.HullIDNumber != "HIN-ABC-12345" {
		t.Fatalf("expected HIN, got %s", created.HullIDNumber)
	}
}

func TestPropertyIndex_RecoverVehicle(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	idx.CreateVehicle(context.Background(), &models.StolenVehicle{
		RecordID: "V-001", PlateNumber: "AA-1234", VehicleMake: "Toyota",
		VehicleModel: "Hilux", VehicleYear: 2020, TheftDate: "2026-06-01",
		TheftLocation: "Delmas", EnteringAgency: "PNH-DELMAS", Status: "STOLEN",
	})
	err := idx.RecoverVehicle(context.Background(), "V-001", "Pétion-Ville", "PNH-PAP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPropertyIndex_CreateArticle(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	err := idx.CreateArticle(context.Background(), &models.StolenArticle{
		RecordID: "A-001", Category: "CATTLE", Description: "Zebu de 500kg, marque PNH sur flanc",
		SerialNumber: "EAR-12345", EstimatedValue: 150000, CurrencyCode: "HTG",
		TheftDate: "2026-06-10", TheftLocation: "Arcahaie",
		OwnerNIU: "NIU-001", Status: "STOLEN", EnteringAgency: "PNH-ARC",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(mock.articles))
	}
	if mock.articles[0].Category != "CATTLE" {
		t.Fatalf("expected CATTLE category, got %s", mock.articles[0].Category)
	}
}

func TestPropertyIndex_CreateSecurity(t *testing.T) {
	mock := newMockPropertyDB()
	idx := NewPropertyIndex(mock)
	err := idx.CreateSecurity(context.Background(), &models.StolenSecurity{
		RecordID: "S-001", SecurityType: "CHEQUE", Issuer: "BRH",
		SecurityNumber: "CHQ-2026-001", FaceValue: 500000, CurrencyCode: "HTG",
		TheftDate: "2026-06-05", TheftLocation: "PAP",
		Status: "STOLEN", EnteringAgency: "PNH-PAP",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.securities) != 1 {
		t.Fatalf("expected 1 security, got %d", len(mock.securities))
	}
	if mock.securities[0].SecurityType != "CHEQUE" {
		t.Fatalf("expected CHEQUE type, got %s", mock.securities[0].SecurityType)
	}
}
