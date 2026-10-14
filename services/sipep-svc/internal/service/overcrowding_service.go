package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OvercrowdingService struct {
	pool *pgxpool.Pool
}

func NewOvercrowdingService(pool *pgxpool.Pool) *OvercrowdingService {
	return &OvercrowdingService{pool: pool}
}

type FacilityOccupancy struct {
	Facility      string  `json:"facility"`
	CurrentCount  int     `json:"current_count"`
	Capacity      int     `json:"capacity"`
	OccupancyRate float64 `json:"occupancy_rate"`
	IsOvercrowded bool    `json:"is_overcrowded"`
	DeptCode      string  `json:"dept_code"`
}

type OvercrowdingAlert struct {
	Facility      string  `json:"facility"`
	CurrentCount  int     `json:"current_count"`
	Threshold     float64 `json:"threshold"`
	OccupancyRate float64 `json:"occupancy_rate"`
	DeptCode      string  `json:"dept_code"`
}

var facilityCapacities = map[string]int{
	"Pénitencier National P-au-P":  3500,
	"Prison Civile Cap-Haïtien":    800,
	"Prison Civile Gonaïves":       400,
	"Prison Civile Les Cayes":      300,
	"CERMICOL (Mineurs)":           100,
	"Établissement femmes (RESEK)": 150,
}

func (s *OvercrowdingService) GetFacilityOccupancy(ctx context.Context, threshold float64) ([]*FacilityOccupancy, error) {
	query := `
		SELECT current_facility, COUNT(*) as current_count, current_dept_code
		FROM sipep_inmates
		WHERE is_currently_detained = TRUE
		GROUP BY current_facility, current_dept_code
		ORDER BY current_facility
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query occupancy: %w", err)
	}
	defer rows.Close()

	var occupancies []*FacilityOccupancy
	for rows.Next() {
		occ := &FacilityOccupancy{}
		err := rows.Scan(&occ.Facility, &occ.CurrentCount, &occ.DeptCode)
		if err != nil {
			return nil, fmt.Errorf("failed to scan occupancy: %w", err)
		}

		if cap, ok := facilityCapacities[occ.Facility]; ok {
			occ.Capacity = cap
			occ.OccupancyRate = float64(occ.CurrentCount) / float64(cap)
			occ.IsOvercrowded = occ.OccupancyRate > threshold
		} else {
			occ.Capacity = 500
			occ.OccupancyRate = float64(occ.CurrentCount) / 500.0
			occ.IsOvercrowded = occ.OccupancyRate > threshold
		}

		occupancies = append(occupancies, occ)
	}
	return occupancies, nil
}

func (s *OvercrowdingService) GetOvercrowdingAlerts(ctx context.Context, threshold float64) ([]*OvercrowdingAlert, error) {
	occupancies, err := s.GetFacilityOccupancy(ctx, threshold)
	if err != nil {
		return nil, err
	}

	var alerts []*OvercrowdingAlert
	for _, occ := range occupancies {
		if occ.IsOvercrowded {
			alerts = append(alerts, &OvercrowdingAlert{
				Facility:      occ.Facility,
				CurrentCount:  occ.CurrentCount,
				Threshold:     threshold,
				OccupancyRate: occ.OccupancyRate,
				DeptCode:      occ.DeptCode,
			})
		}
	}
	return alerts, nil
}

func (s *OvercrowdingService) GetPreventiveDetentionStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_preventive,
			AVG(EXTRACT(DAY FROM NOW() - intake_date)) as avg_days_detained
		FROM sipep_detentions
		WHERE detention_basis = 'PREVENTIVE' AND release_date IS NULL
	`
	var totalPreventive int
	var avgDays float64
	err := s.pool.QueryRow(ctx, query).Scan(&totalPreventive, &avgDays)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}

	return map[string]interface{}{
		"total_preventive":  totalPreventive,
		"avg_days_detained": avgDays,
	}, nil
}
