package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep/internal/domain"
	"github.com/snisid/platform/services/sipep/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestInmateIntake(t *testing.T) {
	svc := service.NewInmateService()

	req := domain.IntakeRequest{
		SNISIDPersonID: uuid.New(),
		Facility:       "PNPP",
		DetentionBasis: domain.DetentionBasisPreventive,
		IntakeOfficer:  uuid.New(),
		IsFemale:       false,
	}

	inmate, detention, err := svc.Intake(req)
	assert.NoError(t, err)
	assert.NotNil(t, inmate)
	assert.NotNil(t, detention)
	assert.Equal(t, "PNPP", inmate.CurrentFacility)
	assert.True(t, inmate.IsCurrentlyDetained)
	assert.Contains(t, inmate.NationalInmateID, "SIPEP-HT-")
	assert.Equal(t, domain.LegalStatusAwaitingTrial, detention.LegalStatus)
}

func TestGetInmateNotFound(t *testing.T) {
	svc := service.NewInmateService()
	_, err := svc.GetInmate(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProcessRelease(t *testing.T) {
	svc := service.NewInmateService()

	req := domain.IntakeRequest{
		SNISIDPersonID: uuid.New(),
		Facility:       "PNPP",
		DetentionBasis: domain.DetentionBasisSentenced,
		IntakeOfficer:  uuid.New(),
	}
	inmate, _, err := svc.Intake(req)
	assert.NoError(t, err)

	releaseReq := domain.ReleaseRequest{
		ReleaseType: domain.ReleaseTypeSentenceServed,
		Authority:   "Juge d'application des peines",
	}
	detention, err := svc.ProcessRelease(inmate.InmateID, releaseReq, uuid.New())
	assert.NoError(t, err)
	assert.NotNil(t, detention.ReleaseDate)
	assert.Equal(t, domain.ReleaseTypeSentenceServed, *detention.ReleaseType)

	updated, _ := svc.GetInmate(inmate.InmateID)
	assert.False(t, updated.IsCurrentlyDetained)
}

func TestReleaseAlreadyReleased(t *testing.T) {
	svc := service.NewInmateService()

	req := domain.IntakeRequest{
		SNISIDPersonID: uuid.New(),
		Facility:       "PCCH",
		DetentionBasis: domain.DetentionBasisPreventive,
		IntakeOfficer:  uuid.New(),
	}
	inmate, _, err := svc.Intake(req)
	assert.NoError(t, err)

	releaseReq := domain.ReleaseRequest{
		ReleaseType: domain.ReleaseTypeBail,
		Authority:   "Tribunal",
	}
	_, err = svc.ProcessRelease(inmate.InmateID, releaseReq, uuid.New())
	assert.NoError(t, err)

	_, err = svc.ProcessRelease(inmate.InmateID, releaseReq, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not currently detained")
}

func TestSearchInmates(t *testing.T) {
	svc := service.NewInmateService()

	personID := uuid.New()
	req := domain.IntakeRequest{
		SNISIDPersonID: personID,
		Facility:       "PNPP",
		DetentionBasis: domain.DetentionBasisPreventive,
		IntakeOfficer:  uuid.New(),
	}
	_, _, _ = svc.Intake(req)

	results, err := svc.Search(personID.String())
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = svc.Search("PNPP")
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = svc.Search("NONEXISTENT")
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestEscapeRelease(t *testing.T) {
	svc := service.NewInmateService()

	req := domain.IntakeRequest{
		SNISIDPersonID: uuid.New(),
		Facility:       "PNPP",
		DetentionBasis: domain.DetentionBasisPreventive,
		IntakeOfficer:  uuid.New(),
	}
	inmate, _, err := svc.Intake(req)
	assert.NoError(t, err)

	releaseReq := domain.ReleaseRequest{
		ReleaseType: domain.ReleaseTypeEscape,
		Authority:   "DAP",
	}
	detention, err := svc.ProcessRelease(inmate.InmateID, releaseReq, uuid.New())
	assert.NoError(t, err)
	assert.Equal(t, domain.ReleaseTypeEscape, *detention.ReleaseType)

	updated, _ := svc.GetInmate(inmate.InmateID)
	assert.False(t, updated.IsCurrentlyDetained)
}

func TestIntakeWithFullDetails(t *testing.T) {
	svc := service.NewInmateService()
	now := time.Now().AddDate(0, 0, 30)
	sentenceDays := 365

	req := domain.IntakeRequest{
		SNISIDPersonID:       uuid.New(),
		Facility:             "PCCH",
		CellBlock:            "B-12",
		DetentionBasis:       domain.DetentionBasisSentenced,
		LegalStatus:          domain.LegalStatusSentenced,
		CaseReference:        "PCCH-2024-0042",
		CourtName:            "Tribunal Cap-Haïtien",
		ArrestingAuthority:   "PNH",
		WarrantNumber:        "W-2024-8910",
		IntakeOfficer:        uuid.New(),
		IsMinor:              false,
		IsFemale:             true,
		HasSpecialNeeds:      true,
		SpecialNeedsNotes:    "Grossesse à risque",
		SentenceDurationDays: &sentenceDays,
		ExpectedReleaseDate:  &now,
	}

	inmate, detention, err := svc.Intake(req)
	assert.NoError(t, err)
	assert.Equal(t, "PCCH", inmate.CurrentFacility)
	assert.Equal(t, "B-12", inmate.CellBlock)
	assert.True(t, inmate.HasSpecialNeeds)
	assert.True(t, inmate.IsFemale)
	assert.Equal(t, domain.LegalStatusSentenced, detention.LegalStatus)
	assert.Equal(t, "PNH", detention.ArrestingAuthority)
}
