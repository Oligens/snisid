package domain

import (
	"github.com/google/uuid"
	"time"
)

type Datacenter struct {
	DCID                 uuid.UUID `json:"dc_id"`
	DCName               string    `json:"dc_name"`
	DCRole               string    `json:"dc_role"`
	DeptCode             string    `json:"dept_code"`
	TierRating           *string   `json:"tier_rating,omitempty"`
	PowerCapacityKW      *float64  `json:"power_capacity_kw,omitempty"`
	HasGeneratorBackup   bool      `json:"has_generator_backup"`
	HasRedundantInternet bool      `json:"has_redundant_internet"`
	RackCount            *int      `json:"rack_count,omitempty"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
}

type K8sCluster struct {
	ClusterID         uuid.UUID `json:"cluster_id"`
	DCID              uuid.UUID `json:"dc_id"`
	ClusterName       string    `json:"cluster_name"`
	Distro            string    `json:"distro"`
	NodeCount         int       `json:"node_count"`
	KubernetesVersion string    `json:"kubernetes_version"`
	IsProduction      bool      `json:"is_production"`
	CreatedAt         time.Time `json:"created_at"`
}

type DRDrill struct {
	DrillID      uuid.UUID `json:"drill_id"`
	DrillDate    time.Time `json:"drill_date"`
	Scenario     string    `json:"scenario"`
	RTOTargetMin int       `json:"rto_target_min"`
	RTOActualMin *int      `json:"rto_actual_min,omitempty"`
	RPOTargetMin int       `json:"rpo_target_min"`
	RPOActualMin *int      `json:"rpo_actual_min,omitempty"`
	Success      *bool     `json:"success,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
