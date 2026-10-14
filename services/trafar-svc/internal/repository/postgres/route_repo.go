package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/snisid/platform/services/trafar-svc/internal/domain"
)

type RouteRepository struct {
	pool *pgxpool.Pool
}

func NewRouteRepository(pool *pgxpool.Pool) *RouteRepository {
	return &RouteRepository{pool: pool}
}

func (r *RouteRepository) CreateRoute(route *domain.TrafarRoute) error {
	ctx := context.Background()
	route.CreatedAt = time.Now().UTC()
	route.UpdatedAt = time.Now().UTC()

	transitPoints, _ := json.Marshal(route.TransitPoints)
	gangIDs, _ := json.Marshal(route.AssociatedGangIDs)
	suppliers, _ := json.Marshal(route.KnownSuppliers)
	weapons, _ := json.Marshal(route.WeaponTypes)
	caseRefs, _ := json.Marshal(route.LinkedCaseRefs)
	biarIDs, _ := json.Marshal(route.BIARWeaponIDs)
	atfRefs, _ := json.Marshal(route.ATFCaseRefs)

	return r.pool.QueryRow(ctx,
		`INSERT INTO trafar_routes (
			route_name, route_type, trafficking_method,
			origin_country, origin_city, transit_points, entry_point_haiti,
			entry_dept_code, associated_gang_ids, known_suppliers,
			activity_level, estimated_volume_monthly, weapon_types,
			intel_confidence, first_detected, last_confirmed,
			linked_case_refs, biar_weapon_ids, atf_case_refs,
			unodc_ref, analyst_notes, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING route_id, created_at, updated_at`,
		route.RouteName, route.RouteType, route.TraffickingMethod,
		route.OriginCountry, route.OriginCity, transitPoints, route.EntryPointHaiti,
		route.EntryDeptCode, gangIDs, suppliers,
		route.ActivityLevel, route.EstimatedVolumeMonthly, weapons,
		route.IntelConfidence, route.FirstDetected, route.LastConfirmed,
		caseRefs, biarIDs, atfRefs,
		route.UNODCRef, route.AnalystNotes, route.CreatedBy,
	).Scan(&route.RouteID, &route.CreatedAt, &route.UpdatedAt)
}

func (r *RouteRepository) FindByID(id uuid.UUID) (*domain.TrafarRoute, error) {
	ctx := context.Background()
	var route domain.TrafarRoute
	var transitPoints, gangIDs, suppliers, weapons, caseRefs, biarIDs, atfRefs []byte

	err := r.pool.QueryRow(ctx,
		`SELECT route_id, route_name, route_type, trafficking_method,
			origin_country, origin_city, transit_points, entry_point_haiti,
			entry_dept_code, associated_gang_ids, known_suppliers,
			activity_level, estimated_volume_monthly, weapon_types,
			intel_confidence, first_detected, last_confirmed,
			linked_case_refs, biar_weapon_ids, atf_case_refs,
			unodc_ref, analyst_notes, created_by, created_at, updated_at
		FROM trafar_routes WHERE route_id = $1`, id,
	).Scan(
		&route.RouteID, &route.RouteName, &route.RouteType, &route.TraffickingMethod,
		&route.OriginCountry, &route.OriginCity, &transitPoints, &route.EntryPointHaiti,
		&route.EntryDeptCode, &gangIDs, &suppliers,
		&route.ActivityLevel, &route.EstimatedVolumeMonthly, &weapons,
		&route.IntelConfidence, &route.FirstDetected, &route.LastConfirmed,
		&caseRefs, &biarIDs, &atfRefs,
		&route.UNODCRef, &route.AnalystNotes, &route.CreatedBy, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(transitPoints, &route.TransitPoints)
	json.Unmarshal(gangIDs, &route.AssociatedGangIDs)
	json.Unmarshal(suppliers, &route.KnownSuppliers)
	json.Unmarshal(weapons, &route.WeaponTypes)
	json.Unmarshal(caseRefs, &route.LinkedCaseRefs)
	json.Unmarshal(biarIDs, &route.BIARWeaponIDs)
	json.Unmarshal(atfRefs, &route.ATFCaseRefs)

	return &route, nil
}

func (r *RouteRepository) FindAll() ([]domain.TrafarRoute, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT route_id, route_name, route_type, trafficking_method,
			origin_country, origin_city, transit_points, entry_point_haiti,
			entry_dept_code, associated_gang_ids, known_suppliers,
			activity_level, estimated_volume_monthly, weapon_types,
			intel_confidence, first_detected, last_confirmed,
			linked_case_refs, biar_weapon_ids, atf_case_refs,
			unodc_ref, analyst_notes, created_by, created_at, updated_at
		FROM trafar_routes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []domain.TrafarRoute
	for rows.Next() {
		var route domain.TrafarRoute
		var transitPoints, gangIDs, suppliers, weapons, caseRefs, biarIDs, atfRefs []byte
		if err := rows.Scan(
			&route.RouteID, &route.RouteName, &route.RouteType, &route.TraffickingMethod,
			&route.OriginCountry, &route.OriginCity, &transitPoints, &route.EntryPointHaiti,
			&route.EntryDeptCode, &gangIDs, &suppliers,
			&route.ActivityLevel, &route.EstimatedVolumeMonthly, &weapons,
			&route.IntelConfidence, &route.FirstDetected, &route.LastConfirmed,
			&caseRefs, &biarIDs, &atfRefs,
			&route.UNODCRef, &route.AnalystNotes, &route.CreatedBy, &route.CreatedAt, &route.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(transitPoints, &route.TransitPoints)
		json.Unmarshal(gangIDs, &route.AssociatedGangIDs)
		json.Unmarshal(suppliers, &route.KnownSuppliers)
		json.Unmarshal(weapons, &route.WeaponTypes)
		json.Unmarshal(caseRefs, &route.LinkedCaseRefs)
		json.Unmarshal(biarIDs, &route.BIARWeaponIDs)
		json.Unmarshal(atfRefs, &route.ATFCaseRefs)
		routes = append(routes, route)
	}
	return routes, nil
}

func (r *RouteRepository) CreateShipment(shipment *domain.TrafarShipment) error {
	ctx := context.Background()
	weapons, _ := json.Marshal(shipment.WeaponsTypes)
	persons, _ := json.Marshal(shipment.LinkedPersons)

	return r.pool.QueryRow(ctx,
		`INSERT INTO trafar_shipments (
			route_id, shipment_date, intercepted,
			interception_date, interception_location, interception_unit,
			weapons_count, weapon_types, estimated_value_usd,
			linked_persons, port_ht_ref, mar_ht_ref, case_reference, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING shipment_id, created_at`,
		shipment.RouteID, shipment.ShipmentDate, shipment.Intercepted,
		shipment.InterceptionDate, shipment.InterceptionLocation, shipment.InterceptionUnit,
		shipment.WeaponsCount, weapons, shipment.EstimatedValueUSD,
		persons, shipment.PortHTRef, shipment.MARHTRef, shipment.CaseReference, shipment.Notes,
	).Scan(&shipment.ShipmentID, &shipment.CreatedAt)
}

func (r *RouteRepository) GetShipmentsByRoute(routeID uuid.UUID) ([]domain.TrafarShipment, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT shipment_id, route_id, shipment_date, intercepted,
			interception_date, interception_location, interception_unit,
			weapons_count, weapon_types, estimated_value_usd,
			linked_persons, port_ht_ref, mar_ht_ref, case_reference, notes, created_at
		FROM trafar_shipments WHERE route_id = $1 ORDER BY shipment_date DESC`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []domain.TrafarShipment
	for rows.Next() {
		var s domain.TrafarShipment
		var weapons, persons []byte
		if err := rows.Scan(
			&s.ShipmentID, &s.RouteID, &s.ShipmentDate, &s.Intercepted,
			&s.InterceptionDate, &s.InterceptionLocation, &s.InterceptionUnit,
			&s.WeaponsCount, &weapons, &s.EstimatedValueUSD,
			&persons, &s.PortHTRef, &s.MARHTRef, &s.CaseReference, &s.Notes, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(weapons, &s.WeaponsTypes)
		json.Unmarshal(persons, &s.LinkedPersons)
		shipments = append(shipments, s)
	}
	return shipments, nil
}

func (r *RouteRepository) GetStatsByOrigin() ([]map[string]interface{}, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT origin_country, COUNT(*) as route_count,
			COALESCE(SUM(estimated_volume_monthly), 0) as total_volume,
			COUNT(*) FILTER (WHERE activity_level = 'HIGH') as high_activity
		FROM trafar_routes GROUP BY origin_country ORDER BY route_count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var country string
		var routeCount, totalVolume, highActivity int
		if err := rows.Scan(&country, &routeCount, &totalVolume, &highActivity); err != nil {
			return nil, err
		}
		stats = append(stats, map[string]interface{}{
			"origin_country": country,
			"route_count":    routeCount,
			"total_volume":   totalVolume,
			"high_activity":  highActivity,
		})
	}
	return stats, nil
}

func (r *RouteRepository) GetSuppliers() ([]domain.TrafarSupplier, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT supplier_id, supplier_name, supplier_type, country, city,
			snisid_person_id, linked_routes, atf_subject_ref, interpol_notice_ref,
			is_active, created_at
		FROM trafar_suppliers WHERE is_active = true ORDER BY supplier_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []domain.TrafarSupplier
	for rows.Next() {
		var s domain.TrafarSupplier
		var routes []byte
		if err := rows.Scan(
			&s.SupplierID, &s.SupplierName, &s.SupplierType, &s.Country, &s.City,
			&s.SNISIDPersonID, &routes, &s.ATFSubjectRef, &s.InterpolNoticeRef,
			&s.IsActive, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(routes, &s.LinkedRoutes)
		suppliers = append(suppliers, s)
	}
	return suppliers, nil
}

func (r *RouteRepository) GetRoutesGeoJSON() (*domain.GeoJSONFeatureCollection, error) {
	routes, err := r.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	fc := &domain.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: []domain.GeoJSONFeature{},
	}

	for _, route := range routes {
		feature := domain.GeoJSONFeature{
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{0, 0},
			},
			Properties: map[string]interface{}{
				"route_id":       route.RouteID,
				"route_name":     route.RouteName,
				"route_type":     route.RouteType,
				"origin_country": route.OriginCountry,
				"entry_haiti":    route.EntryPointHaiti,
				"activity_level": route.ActivityLevel,
				"confidence":     route.IntelConfidence,
			},
		}
		fc.Features = append(fc.Features, feature)
	}
	return fc, nil
}

func scanShipments(rows pgx.Rows) ([]domain.TrafarShipment, error) {
	var shipments []domain.TrafarShipment
	for rows.Next() {
		var s domain.TrafarShipment
		var weapons, persons []byte
		if err := rows.Scan(
			&s.ShipmentID, &s.RouteID, &s.ShipmentDate, &s.Intercepted,
			&s.InterceptionDate, &s.InterceptionLocation, &s.InterceptionUnit,
			&s.WeaponsCount, &weapons, &s.EstimatedValueUSD,
			&persons, &s.PortHTRef, &s.MARHTRef, &s.CaseReference, &s.Notes, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(weapons, &s.WeaponsTypes)
		json.Unmarshal(persons, &s.LinkedPersons)
		shipments = append(shipments, s)
	}
	return shipments, nil
}
