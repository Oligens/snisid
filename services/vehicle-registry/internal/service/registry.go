package service

import (
	"fmt"
	"sync"
	"time"
)

type Vehicle struct {
	Plate            string    `json:"plate"`
	VIN              string    `json:"vin"`
	Make             string    `json:"make"`
	Model            string    `json:"model"`
	Year             int       `json:"year"`
	Color            string    `json:"color"`
	OwnerID          string    `json:"owner_id"`
	RegistrationDate time.Time `json:"registration_date"`
	Status           string    `json:"status"`
	InsuranceData    string    `json:"insurance_data,omitempty"`
}

type Registry struct {
	mu       sync.RWMutex
	vehicles map[string]*Vehicle
}

func NewRegistry() *Registry {
	return &Registry{
		vehicles: make(map[string]*Vehicle),
	}
}

func (r *Registry) RegisterVehicle(plate, vin, make, model string, year int, color, ownerID, insuranceData string) (*Vehicle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.vehicles[plate]; exists {
		return nil, fmt.Errorf("vehicle with plate %s already registered", plate)
	}

	v := &Vehicle{
		Plate:            plate,
		VIN:              vin,
		Make:             make,
		Model:            model,
		Year:             year,
		Color:            color,
		OwnerID:          ownerID,
		RegistrationDate: time.Now(),
		Status:           "active",
		InsuranceData:    insuranceData,
	}
	r.vehicles[plate] = v
	return v, nil
}

func (r *Registry) LookupByPlate(plate string) (*Vehicle, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.vehicles[plate]
	return v, ok
}

func (r *Registry) TransferOwnership(plate, newOwnerID string) (*Vehicle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.vehicles[plate]
	if !ok {
		return nil, fmt.Errorf("vehicle with plate %s not found", plate)
	}
	v.OwnerID = newOwnerID
	return v, nil
}

func (r *Registry) SearchVehicles(ownerID, make, model string) []*Vehicle {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Vehicle
	for _, v := range r.vehicles {
		if ownerID != "" && v.OwnerID != ownerID {
			continue
		}
		if make != "" && v.Make != make {
			continue
		}
		if model != "" && v.Model != model {
			continue
		}
		results = append(results, v)
	}
	return results
}
