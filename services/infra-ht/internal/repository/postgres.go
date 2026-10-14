package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/snisid/infra-ht/internal/domain"
)

type Repository interface {
	GetDatacenters(ctx context.Context) ([]domain.Datacenter, error)
	GetClusters(ctx context.Context) ([]domain.K8sCluster, error)
	CreateDRDrill(ctx context.Context, d *domain.DRDrill) error
}

type postgresRepo struct{ db *sql.DB }

func NewPostgresRepo(db *sql.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) GetDatacenters(ctx context.Context) ([]domain.Datacenter, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT dc_id, dc_name, dc_role, dept_code, tier_rating, power_capacity_kw, has_generator_backup, has_redundant_internet, rack_count, is_active, created_at FROM infra_datacenters WHERE is_active = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dcs []domain.Datacenter
	for rows.Next() {
		var d domain.Datacenter
		rows.Scan(&d.DCID, &d.DCName, &d.DCRole, &d.DeptCode, &d.TierRating, &d.PowerCapacityKW, &d.HasGeneratorBackup, &d.HasRedundantInternet, &d.RackCount, &d.IsActive, &d.CreatedAt)
		dcs = append(dcs, d)
	}
	return dcs, rows.Err()
}
func (r *postgresRepo) GetClusters(ctx context.Context) ([]domain.K8sCluster, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT cluster_id, dc_id, cluster_name, distro, node_count, kubernetes_version, is_production, created_at FROM infra_k8s_clusters`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cls []domain.K8sCluster
	for rows.Next() {
		var c domain.K8sCluster
		rows.Scan(&c.ClusterID, &c.DCID, &c.ClusterName, &c.Distro, &c.NodeCount, &c.KubernetesVersion, &c.IsProduction, &c.CreatedAt)
		cls = append(cls, c)
	}
	return cls, rows.Err()
}
func (r *postgresRepo) CreateDRDrill(ctx context.Context, d *domain.DRDrill) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO infra_dr_drills (drill_id, drill_date, scenario, rto_target_minutes, rto_actual_minutes, rpo_target_minutes, rpo_actual_minutes, success, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.DrillID, d.DrillDate, d.Scenario, d.RTOTargetMin, d.RTOActualMin, d.RPOTargetMin, d.RPOActualMin, d.Success, d.Notes, time.Now().UTC())
	return err
}
