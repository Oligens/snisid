package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/gang/internal/domain"
)

type GangRepository interface {
	Create(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	List(ctx context.Context) ([]*domain.Organization, error)
	ByDeptCode(ctx context.Context, code string) ([]*domain.Organization, error)
	Sanctioned(ctx context.Context) ([]*domain.Organization, error)
}

type InMemoryGangRepo struct {
	mu   sync.RWMutex
	orgs map[uuid.UUID]*domain.Organization
}

func NewInMemoryGangRepo() *InMemoryGangRepo {
	return &InMemoryGangRepo{
		orgs: make(map[uuid.UUID]*domain.Organization),
	}
}

func (r *InMemoryGangRepo) Create(ctx context.Context, org *domain.Organization) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orgs[org.GangID] = org
	return nil
}

func (r *InMemoryGangRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	org, ok := r.orgs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return org, nil
}

func (r *InMemoryGangRepo) List(ctx context.Context) ([]*domain.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Organization, 0, len(r.orgs))
	for _, org := range r.orgs {
		if org.IsActive {
			result = append(result, org)
		}
	}
	return result, nil
}

func (r *InMemoryGangRepo) ByDeptCode(ctx context.Context, code string) ([]*domain.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Organization
	for _, org := range r.orgs {
		if org.IsActive && org.PrimaryDeptCode == code {
			result = append(result, org)
		}
	}
	return result, nil
}

func (r *InMemoryGangRepo) Sanctioned(ctx context.Context) ([]*domain.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Organization
	for _, org := range r.orgs {
		if org.OFACDesignation || org.UNDesignationDate != nil {
			result = append(result, org)
		}
	}
	return result, nil
}
