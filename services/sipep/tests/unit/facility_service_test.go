package unit

import (
	"testing"

	"github.com/snisid/platform/services/sipep/internal/domain"
	"github.com/snisid/platform/services/sipep/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestFacilitySeed(t *testing.T) {
	svc := service.NewFacilityService()
	facilities := svc.GetAll()
	assert.Len(t, facilities, 6)
}

func TestGetFacilityByCode(t *testing.T) {
	svc := service.NewFacilityService()

	f, err := svc.GetByCode("PNPP")
	assert.NoError(t, err)
	assert.Equal(t, "Pénitencier National P-au-P", f.Name)
	assert.Equal(t, domain.FacilityTypeNational, f.FacilityType)

	_, err = svc.GetByCode("INVALID")
	assert.Error(t, err)
}

func TestOccupancyReport(t *testing.T) {
	facilitySvc := service.NewFacilityService()

	counts := map[string]int{
		"PNPP": 3500,
		"PCCH": 900,
	}

	reports := facilitySvc.GetOccupancy(counts)
	assert.NotEmpty(t, reports)

	for _, r := range reports {
		if r.Facility == "Pénitencier National P-au-P" {
			assert.Equal(t, 3500, r.CurrentCount)
			assert.Equal(t, 3500, r.Capacity)
			assert.Equal(t, 1.0, r.OccupancyRate)
		}
	}
}

func TestOvercrowdingDetection(t *testing.T) {
	facilitySvc := service.NewFacilityService()

	counts := map[string]int{
		"PNPP": 5500,
		"PCCH": 800,
	}

	reports := facilitySvc.GetOccupancy(counts)
	var overcrowded []domain.OccupancyReport
	for _, r := range reports {
		if r.OccupancyRate > 1.5 {
			overcrowded = append(overcrowded, *r)
		}
	}
	assert.NotEmpty(t, overcrowded)
	for _, r := range overcrowded {
		assert.True(t, r.OccupancyRate > 1.5)
	}
}
