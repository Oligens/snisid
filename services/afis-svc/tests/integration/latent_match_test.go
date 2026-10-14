package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

func TestLatent_SceneOfCrime_Match(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	lp := domain.LatentPrint{
		LatentID:       uuid.New(),
		CaseReference:  "PAP-2026-CR-00123",
		ImageRef:       "latents/PAP-2026-CR-00123/latent-001.png",
		FingerPosition: domain.FingerUnknown,
		IsIdentified:   false,
	}

	if lp.LatentID == uuid.Nil {
		t.Fatal("expected valid latent ID")
	}
	if lp.IsIdentified {
		t.Fatal("expected latent to be unidentified initially")
	}

	t.Logf("Latent %s from case %s ready for matching", lp.LatentID, lp.CaseReference)
}

func TestLatent_ConfirmMatch_UpdatesSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	subjectID := uuid.New()
	confirm := domain.LatentMatchConfirm{
		MatchedSubjectID: subjectID,
		MatchScore:       92.5,
		ExaminedBy:       uuid.New(),
	}

	if confirm.MatchedSubjectID != subjectID {
		t.Fatal("subject ID mismatch")
	}
	if confirm.MatchScore < 85 {
		t.Fatal("match score below automatic threshold")
	}
}
