package federation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Gateway struct {
	connectors []AgencyConnector
	timeout    time.Duration
}

func NewGateway(connectors ...AgencyConnector) *Gateway {
	return &Gateway{
		connectors: connectors,
		timeout:    1 * time.Second,
	}
}

func (g *Gateway) Search(ctx context.Context, query string) ([]AgencyResult, error) {
	// 1. Create a timeout context for the overall federated query
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	results := make([]AgencyResult, len(g.connectors))
	var mu sync.Mutex

	logger.Info(ctx, "Starting federated cross-agency search",
		zap.String("query", query),
		zap.Int("connector_count", len(g.connectors)),
	)

	for i, conn := range g.connectors {
		idx := i
		c := conn
		eg.Go(func() error {
			res, err := c.Fetch(ctx, query)

			mu.Lock()
			if err != nil {
				logger.Warn(ctx, fmt.Sprintf("Connector %s failed", c.Name()), zap.Error(err))
				results[idx] = AgencyResult{Source: c.Name(), Error: err}
			} else {
				results[idx] = res
			}
			mu.Unlock()

			// We return nil to eg.Go because we handle errors internally to allow partial results
			return nil
		})
	}

	// Wait for all connectors or the timeout
	_ = eg.Wait()

	return results, nil
}
