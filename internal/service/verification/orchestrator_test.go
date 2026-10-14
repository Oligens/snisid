package verification

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConnector struct {
	name   string
	result Result
	err    error
}

func (c *mockConnector) Name() string { return c.name }
func (c *mockConnector) Verify(ctx context.Context, data map[string]interface{}) (Result, error) {
	if c.err != nil {
		return Result{}, c.err
	}
	return c.result, nil
}

func TestNewOrchestrator(t *testing.T) {
	o := NewOrchestrator()
	assert.NotNil(t, o)
	assert.Empty(t, o.connectors)
}

func TestOrchestrator_VerifyIdentity_SingleConnector(t *testing.T) {
	conn := &mockConnector{
		name:   "oni",
		result: Result{Status: StatusSuccess, Reason: "Verified", Score: 100},
	}
	o := NewOrchestrator(conn)

	results, err := o.VerifyIdentity(context.Background(), map[string]interface{}{"id": "123"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, StatusSuccess, results["oni"].Status)
	assert.Equal(t, 100, results["oni"].Score)
}

func TestOrchestrator_VerifyIdentity_MultipleConnectors(t *testing.T) {
	conn1 := &mockConnector{name: "biometric", result: Result{Status: StatusSuccess, Score: 95}}
	conn2 := &mockConnector{name: "dgi", result: Result{Status: StatusSuccess, Score: 100}}
	conn3 := &mockConnector{name: "police", result: Result{Status: StatusSuccess, Score: 90}}

	o := NewOrchestrator(conn1, conn2, conn3)
	results, err := o.VerifyIdentity(context.Background(), map[string]interface{}{"id": "123"})
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestOrchestrator_VerifyIdentity_ConnectorFailure_ReturnsError(t *testing.T) {
	conn := &mockConnector{
		name: "failing",
		err:  errors.New("connection refused"),
	}
	o := NewOrchestrator(conn)

	_, err := o.VerifyIdentity(context.Background(), map[string]interface{}{"id": "123"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification orchestration failed")
}

func TestOrchestrator_VerifyIdentity_EmptyData(t *testing.T) {
	conn := &mockConnector{
		name:   "oni",
		result: Result{Status: StatusFailed, Reason: "No data", Score: 0},
	}
	o := NewOrchestrator(conn)

	results, err := o.VerifyIdentity(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, results["oni"].Status)
}

func TestOrchestrator_VerifyIdentity_PartialFailure(t *testing.T) {
	conn1 := &mockConnector{name: "good", result: Result{Status: StatusSuccess, Score: 100}}
	conn2 := &mockConnector{name: "bad", err: errors.New("timeout")}

	o := NewOrchestrator(conn1, conn2)
	_, err := o.VerifyIdentity(context.Background(), map[string]interface{}{"id": "123"})
	assert.Error(t, err)
}
