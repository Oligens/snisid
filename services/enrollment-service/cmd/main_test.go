package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	db.AutoMigrate(&Enrollment{})
	return db
}

func setupTestService(t *testing.T) *EnrollmentService {
	db := setupTestDB(t)
	producer := events.NewProducer([]string{"localhost:9092"}, "test.events")
	return NewEnrollmentService(db, producer)
}

func TestNewEnrollmentService(t *testing.T) {
	svc := setupTestService(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.db)
}

func TestCreateEnrollment_Success(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName:    "Jean",
		LastName:     "Dupont",
		DateOfBirth:  "1990-01-15",
		PlaceOfBirth: "Port-au-Prince",
		Gender:       "M",
		Nationality:  "HTI",
		AgencyID:     "ONI-AGENCY",
		AgentID:      "AGENT-001",
		ConsentGiven: true,
	}

	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)
	assert.NotEmpty(t, enrollment.ID)
	assert.Equal(t, "Jean", enrollment.FirstName)
	assert.Equal(t, EnrollmentPending, enrollment.Status)
	assert.False(t, enrollment.ExpiresAt.IsZero())
	assert.Equal(t, 1, enrollment.Version)
}

func TestCreateEnrollment_WithAddress(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName:    "Marie",
		LastName:     "Pierre",
		DateOfBirth:  "1985-06-20",
		PlaceOfBirth: "Cap-Haïtien",
		Gender:       "F",
		Nationality:  "HTI",
		AgencyID:     "ONI-AGENCY",
		AgentID:      "AGENT-002",
		Address:      map[string]interface{}{"city": "Pétion-Ville", "street": "Rue 5"},
		ConsentGiven: true,
	}

	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)
	assert.Contains(t, enrollment.AddressJSON, "Pétion-Ville")
}

func TestCaptureBiometrics_Success(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Pierre", LastName: "Louis", DateOfBirth: "2000-01-01",
		PlaceOfBirth: "Jacmel", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-003", ConsentGiven: true,
	}
	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)

	updated, err := svc.CaptureBiometrics(enrollment.ID, "hash-123", 0.85)
	require.NoError(t, err)
	assert.Equal(t, EnrollmentBiometrics, updated.Status)
	assert.Equal(t, "hash-123", updated.BiometricHash)
	assert.Equal(t, 0.85, updated.QualityScore)
	assert.Equal(t, 2, updated.Version)
}

func TestCaptureBiometrics_LowQuality(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Alice", LastName: "Jean", DateOfBirth: "1995-03-10",
		PlaceOfBirth: "Gonaïves", Gender: "F", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-004", ConsentGiven: true,
	}
	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)

	_, err = svc.CaptureBiometrics(enrollment.ID, "hash-low", 0.3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quality too low")
}

func TestCaptureBiometrics_WrongState(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.CaptureBiometrics("nonexistent-id", "hash", 0.9)
	assert.Error(t, err)
}

func TestVerifyEnrollment_Success(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Robert", LastName: "Michel", DateOfBirth: "1988-11-22",
		PlaceOfBirth: "Les Cayes", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-005", LocationID: "CAY",
		ConsentGiven: true,
	}
	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)

	enrollment, err = svc.CaptureBiometrics(enrollment.ID, "hash-456", 0.92)
	require.NoError(t, err)

	verified, err := svc.VerifyEnrollment(enrollment.ID, true)
	require.NoError(t, err)
	assert.Equal(t, EnrollmentVerified, verified.Status)
	assert.NotEmpty(t, verified.NNU)
	assert.True(t, verified.OtpVerified)
}

func TestVerifyEnrollment_WrongState(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Test", LastName: "User", DateOfBirth: "2000-01-01",
		PlaceOfBirth: "Port-au-Prince", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-006", ConsentGiven: true,
	}
	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)

	_, err = svc.VerifyEnrollment(enrollment.ID, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot verify enrollment in status")
}

func TestCompleteEnrollment_Success(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Sophie", LastName: "Dumas", DateOfBirth: "1992-07-14",
		PlaceOfBirth: "Port-de-Paix", Gender: "F", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-007", LocationID: "PDP",
		ConsentGiven: true,
	}
	e, _ := svc.CreateEnrollment(req)
	e, _ = svc.CaptureBiometrics(e.ID, "hash-789", 0.95)
	e, _ = svc.VerifyEnrollment(e.ID, true)

	completed, err := svc.CompleteEnrollment(e.ID)
	require.NoError(t, err)
	assert.Equal(t, EnrollmentCompleted, completed.Status)
	require.NotNil(t, completed.CompletedAt)
}

func TestCompleteEnrollment_WrongState(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.CompleteEnrollment("nonexistent")
	assert.Error(t, err)
}

func TestRejectEnrollment(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Bad", LastName: "User", DateOfBirth: "2000-01-01",
		PlaceOfBirth: "Port-au-Prince", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-008", ConsentGiven: true,
	}
	e, _ := svc.CreateEnrollment(req)

	rejected, err := svc.RejectEnrollment(e.ID, "Fraudulent documents")
	require.NoError(t, err)
	assert.Equal(t, EnrollmentRejected, rejected.Status)
	assert.Equal(t, "Fraudulent documents", rejected.StatusReason)
}

func TestGetEnrollment_NotFound(t *testing.T) {
	svc := setupTestService(t)
	_, err := svc.GetEnrollment("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "enrollment not found")
}

func TestGetEnrollment_Success(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Found", LastName: "User", DateOfBirth: "1990-01-01",
		PlaceOfBirth: "Port-au-Prince", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-009", ConsentGiven: true,
	}
	created, _ := svc.CreateEnrollment(req)

	fetched, err := svc.GetEnrollment(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Found", fetched.FirstName)
}

func TestSearchEnrollments_ByName(t *testing.T) {
	svc := setupTestService(t)
	svc.CreateEnrollment(CreateEnrollmentRequest{
		FirstName: "Jean", LastName: "Dupont", DateOfBirth: "1990-01-01",
		PlaceOfBirth: "Port-au-Prince", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-010", ConsentGiven: true,
	})

	results, total, err := svc.SearchEnrollments(SearchEnrollmentRequest{
		SearchTerm: "Dupont",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
}

func TestSearchEnrollments_ByStatus(t *testing.T) {
	svc := setupTestService(t)
	svc.CreateEnrollment(CreateEnrollmentRequest{
		FirstName: "Status", LastName: "User", DateOfBirth: "1990-01-01",
		PlaceOfBirth: "Port-au-Prince", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-011", ConsentGiven: true,
	})

	results, total, err := svc.SearchEnrollments(SearchEnrollmentRequest{
		Status: "PENDING",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
}

func TestSearchEnrollments_Empty(t *testing.T) {
	svc := setupTestService(t)
	results, total, err := svc.SearchEnrollments(SearchEnrollmentRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, results)
}

func TestGenerateNNU_Uniqueness(t *testing.T) {
	svc := setupTestService(t)
	enrollment := &Enrollment{
		ID:          "test-nnu-1",
		LocationID:  "PAP",
		DateOfBirth: "1990-01-15",
	}

	nnu1, err := svc.generateNNU(enrollment)
	require.NoError(t, err)
	assert.Len(t, nnu1, 12)

	nnu2, err := svc.generateNNU(enrollment)
	require.NoError(t, err)
	assert.NotEqual(t, nnu1, nnu2)
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

func TestNewUUID_Uniqueness(t *testing.T) {
	uuid1 := newUUID()
	uuid2 := newUUID()
	assert.NotEqual(t, uuid1, uuid2)
	assert.Contains(t, uuid1, "-")
}

func TestPublishEvent_NilProducer(t *testing.T) {
	svc := setupTestService(t)
	// Should not panic
	svc.publishEvent("test.event", "id-1", "NNU-001", "ACTIVE", nil)
}

func TestExpiredEnrollment(t *testing.T) {
	svc := setupTestService(t)
	req := CreateEnrollmentRequest{
		FirstName: "Expired", LastName: "User", DateOfBirth: "2000-01-01",
		PlaceOfBirth: "Port-au-Prince", Gender: "M", Nationality: "HTI",
		AgencyID: "ONI", AgentID: "AGENT-012", ConsentGiven: true,
	}
	enrollment, err := svc.CreateEnrollment(req)
	require.NoError(t, err)
	assert.True(t, enrollment.ExpiresAt.After(time.Now()))
	assert.True(t, enrollment.ExpiresAt.Before(time.Now().Add(96*time.Hour)))
}
