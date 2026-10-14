package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"github.com/snisid/platform/internal/domain/authorization/entity"
	"github.com/snisid/platform/internal/domain/authorization/repository"
	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
)

type AuthorizationEngine interface {
	Enforce(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error)
	RefreshPolicies(ctx context.Context) error
}

type opaEngine struct {
	repo          repository.PolicyRepository
	producer      events.ProducerInterface
	compiledQuery *rego.PreparedEvalQuery
	mu            sync.RWMutex
}

func NewOPAEngine(repo repository.PolicyRepository, producer events.ProducerInterface) AuthorizationEngine {
	engine := &opaEngine{
		repo:     repo,
		producer: producer,
	}
	// Initial load will happen externally or handled gracefully
	return engine
}

func (e *opaEngine) RefreshPolicies(ctx context.Context) error {
	policies, err := e.repo.GetActivePolicies(ctx)
	if err != nil {
		return err
	}

	grants, err := e.repo.GetRoleGrants(ctx)
	if err != nil {
		return err
	}

	// Format grants for OPA data
	roleGrantsData := make(map[string][]map[string]interface{})
	for _, g := range grants {
		roleGrantsData[g.Role] = append(roleGrantsData[g.Role], map[string]interface{}{
			"action":   g.Action,
			"resource": g.Resource,
		})
	}

	// Prepare rego options
	options := []func(*rego.Rego){
		rego.Query("data.snisid.abac.allow"), // Default entrypoint
		rego.Store(nil),                      // Ideally an in-memory store for data
	}

	// Load modules
	for _, p := range policies {
		options = append(options, rego.Module(p.Name, p.Module))
	}

	// Prepare query
	query, err := rego.New(options...).PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego query: %w", err)
	}

	e.mu.Lock()
	e.compiledQuery = &query
	e.mu.Unlock()

	logger.Info(ctx, "Authorization policies refreshed")
	return nil
}

func (e *opaEngine) Enforce(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error) {
	e.mu.RLock()
	query := e.compiledQuery
	e.mu.RUnlock()

	if query == nil {
		return nil, fmt.Errorf("policy engine not initialized")
	}

	input := map[string]interface{}{
		"user": map[string]interface{}{
			"id":        req.Subject.UserID,
			"roles":     req.Subject.Roles,
			"agency":    req.Subject.Agency,
			"clearance": req.Subject.Clearance,
		},
		"action":              req.Action,
		"resource":            req.Resource,
		"resource_attributes": req.Attributes,
	}

	rs, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate policy: %w", err)
	}

	allowed := false
	if len(rs) > 0 && len(rs[0].Expressions) > 0 {
		if val, ok := rs[0].Expressions[0].Value.(bool); ok {
			allowed = val
		}
	}

	decision := &entity.AuthorizationDecision{
		Allowed: allowed,
	}

	if !allowed {
		decision.Reason = "Denied by OPA policy"
	}

	// Async audit logging
	go e.logDecision(req, decision)

	return decision, nil
}

func (e *opaEngine) logDecision(req *entity.AuthorizationRequest, dec *entity.AuthorizationDecision) {
	if e.producer == nil {
		return
	}

	evt := map[string]interface{}{
		"userId":    req.Subject.UserID,
		"action":    req.Action,
		"resource":  req.Resource,
		"allowed":   dec.Allowed,
		"reason":    dec.Reason,
		"timestamp": time.Now().UTC(),
	}

	_ = e.producer.Publish(context.Background(), req.Subject.UserID, evt)
}
