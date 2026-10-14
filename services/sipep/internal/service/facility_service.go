package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep/internal/domain"
)

type FacilityService struct {
	mu         sync.RWMutex
	facilities map[string]*domain.Facility
}

func NewFacilityService() *FacilityService {
	svc := &FacilityService{
		facilities: make(map[string]*domain.Facility),
	}
	svc.seed()
	return svc
}

func (s *FacilityService) seed() {
	now := time.Now()
	facilities := []*domain.Facility{
		{FacilityID: uuid.New(), Code: "PNPP", Name: "Pénitencier National P-au-P", Department: "Ouest", DeptCode: "OU", FacilityType: domain.FacilityTypeNational, Capacity: 3500, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{FacilityID: uuid.New(), Code: "PCCH", Name: "Prison Civile Cap-Haïtien", Department: "Nord", DeptCode: "ND", FacilityType: domain.FacilityTypeDepartmental, Capacity: 800, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{FacilityID: uuid.New(), Code: "PCGO", Name: "Prison Civile Gonaïves", Department: "Artibonite", DeptCode: "AR", FacilityType: domain.FacilityTypeDepartmental, Capacity: 400, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{FacilityID: uuid.New(), Code: "PCLC", Name: "Prison Civile Les Cayes", Department: "Sud", DeptCode: "SD", FacilityType: domain.FacilityTypeDepartmental, Capacity: 300, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{FacilityID: uuid.New(), Code: "CML", Name: "CERMICOL (Mineurs)", Department: "Ouest", DeptCode: "OU", FacilityType: domain.FacilityTypeSpecialized, Capacity: 100, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{FacilityID: uuid.New(), Code: "RSEK", Name: "Établissement femmes (RESEK)", Department: "Ouest", DeptCode: "OU", FacilityType: domain.FacilityTypeSpecialized, Capacity: 150, IsActive: true, CreatedAt: now, UpdatedAt: now},
	}
	for _, f := range facilities {
		s.facilities[f.Code] = f
	}
}

func (s *FacilityService) GetAll() []*domain.Facility {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*domain.Facility, 0, len(s.facilities))
	for _, f := range s.facilities {
		result = append(result, f)
	}
	return result
}

func (s *FacilityService) GetByCode(code string) (*domain.Facility, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.facilities[code]
	if !ok {
		return nil, fmt.Errorf("facility not found: %s", code)
	}
	return f, nil
}

func (s *FacilityService) GetOccupancy(inmateCounts map[string]int) []*domain.OccupancyReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var reports []*domain.OccupancyReport
	for code, f := range s.facilities {
		count := inmateCounts[code]
		if count == 0 {
			count = inmateCounts[f.Name]
		}
		rate := 0.0
		if f.Capacity > 0 {
			rate = float64(count) / float64(f.Capacity)
		}
		reports = append(reports, &domain.OccupancyReport{
			Facility:       f.Name,
			DepartmentCode: f.DeptCode,
			CurrentCount:   count,
			Capacity:       f.Capacity,
			OccupancyRate:  rate,
		})
	}
	return reports
}
