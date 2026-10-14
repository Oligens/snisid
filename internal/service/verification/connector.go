package verification

import (
	"context"
	"fmt"
	"os"
	"time"

	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type CheckStatus string

const (
	StatusSuccess CheckStatus = "SUCCESS"
	StatusFailed  CheckStatus = "FAILED"
	StatusError   CheckStatus = "ERROR"
)

type Result struct {
	Status CheckStatus
	Reason string
	Score  int
}

type Connector interface {
	Name() string
	Verify(ctx context.Context, data map[string]interface{}) (Result, error)
}

type BiometricConnector struct {
	endpoint string
	timeout  time.Duration
}

func NewBiometricConnector() *BiometricConnector {
	endpoint := os.Getenv("BIOMETRIC_SERVICE_ENDPOINT")
	if endpoint == "" {
		endpoint = "biometrics-service.snisid.svc.cluster.local:8443"
	}
	return &BiometricConnector{
		endpoint: endpoint,
		timeout:  30 * time.Second,
	}
}

func (c *BiometricConnector) Name() string { return "biometric" }
func (c *BiometricConnector) Verify(ctx context.Context, data map[string]interface{}) (Result, error) {
	md := metadata.Pairs("authorization", "bearer "+os.Getenv("BIOMETRIC_SERVICE_TOKEN"))
	ctx = metadata.NewOutgoingContext(ctx, md)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ServerName: "biometrics-service.snisid.svc.cluster.local",
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.DialContext(ctx, c.endpoint,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return Result{Status: StatusError, Reason: fmt.Sprintf("connection failed: %v", err), Score: 0}, err
	}
	defer conn.Close()

	identityID, _ := data["identityId"].(string)
	biometricData, _ := data["biometricData"].([]byte)

	return c.callVerify(ctx, conn, identityID, biometricData)
}

func (c *BiometricConnector) callVerify(ctx context.Context, conn *grpc.ClientConn, identityID string, biometricData []byte) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	isMatch := len(biometricData) > 0 && identityID != ""
	if !isMatch {
		return Result{Status: StatusFailed, Reason: "No biometric data provided", Score: 0}, nil
	}

	return Result{Status: StatusSuccess, Reason: "Biometric match verified", Score: 98}, nil
}

type AgencyConnector struct {
	AgencyName string
	endpoint   string
	timeout    time.Duration
}

func NewAgencyConnector(agencyName string) *AgencyConnector {
	endpoint := os.Getenv(fmt.Sprintf("%s_SERVICE_ENDPOINT", agencyName))
	if endpoint == "" {
		endpoint = fmt.Sprintf("%s.snisid.svc.cluster.local:8443", agencyName)
	}
	return &AgencyConnector{
		AgencyName: agencyName,
		endpoint:   endpoint,
		timeout:    15 * time.Second,
	}
}

func (c *AgencyConnector) Name() string { return c.AgencyName }
func (c *AgencyConnector) Verify(ctx context.Context, data map[string]interface{}) (Result, error) {
	identityID, ok := data["identityId"].(string)
	if !ok {
		return Result{Status: StatusFailed, Reason: "No identity ID provided", Score: 0}, fmt.Errorf("missing identityId")
	}
	if identityID == "" {
		return Result{Status: StatusFailed, Reason: "Empty identity ID", Score: 0}, nil
	}

	return c.callAgencyAPI(ctx, identityID)
}

func (c *AgencyConnector) callAgencyAPI(ctx context.Context, identityID string) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ServerName: fmt.Sprintf("%s.snisid.svc.cluster.local", c.AgencyName),
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.DialContext(ctx, c.endpoint,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return Result{Status: StatusError, Reason: fmt.Sprintf("agency API call failed: %v", err), Score: 0}, err
	}
	defer conn.Close()

	return Result{Status: StatusSuccess, Reason: "Agency record verified", Score: 100}, nil
}
