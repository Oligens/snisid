package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep/internal/domain"
	"github.com/snisid/platform/services/sipep/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestTransfer(t *testing.T) {
	ms := service.NewMovementService()

	req := domain.TransferRequest{
		InmateID:     uuid.New(),
		ToFacility:   "PCCH",
		Reason:       "Sécurité renforcée",
		AuthorizedBy: uuid.New(),
	}

	m, err := ms.Transfer(req)
	assert.NoError(t, err)
	assert.Equal(t, "PCCH", m.ToFacility)
	assert.Equal(t, domain.MovementTypeTransfer, m.MovementType)
}

func TestCellChange(t *testing.T) {
	ms := service.NewMovementService()
	inmateID := uuid.New()

	m, err := ms.CellChange(inmateID, "A-10", "B-05", uuid.New())
	assert.NoError(t, err)
	assert.Equal(t, "A-10", m.FromBlock)
	assert.Equal(t, "B-05", m.ToBlock)
	assert.Equal(t, domain.MovementTypeCellChange, m.MovementType)
}

func TestGetMovementsByInmate(t *testing.T) {
	ms := service.NewMovementService()
	inmateID := uuid.New()

	_, _ = ms.CellChange(inmateID, "A-10", "B-05", uuid.New())

	req := domain.TransferRequest{
		InmateID:     inmateID,
		ToFacility:   "PCCH",
		AuthorizedBy: uuid.New(),
	}
	_, _ = ms.Transfer(req)

	movements, err := ms.GetByInmate(inmateID)
	assert.NoError(t, err)
	assert.Len(t, movements, 2)
}

func TestGetMovementsNotFound(t *testing.T) {
	ms := service.NewMovementService()
	_, err := ms.GetByInmate(uuid.New())
	assert.Error(t, err)
}
