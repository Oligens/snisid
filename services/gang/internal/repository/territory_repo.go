package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang/internal/domain"
)

type TerritoryRepository interface {
	Create(ctx context.Context, t *domain.Territory) error
	ByGangID(ctx context.Context, gangID uuid.UUID) ([]*domain.Territory, error)
}

type InMemoryTerritoryRepo struct {
	mu          sync.RWMutex
	territories map[uuid.UUID]*domain.Territory
}

func NewInMemoryTerritoryRepo() *InMemoryTerritoryRepo {
	return &InMemoryTerritoryRepo{
		territories: make(map[uuid.UUID]*domain.Territory),
	}
}

func (r *InMemoryTerritoryRepo) Create(ctx context.Context, t *domain.Territory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.territories[t.TerritoryID] = t
	return nil
}

func (r *InMemoryTerritoryRepo) ByGangID(ctx context.Context, gangID uuid.UUID) ([]*domain.Territory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Territory
	for _, t := range r.territories {
		if t.GangID == gangID {
			result = append(result, t)
		}
	}
	return result, nil
}
