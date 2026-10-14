package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/snisid/platform/internal/platform/database"
	"github.com/snisid/platform/internal/platform/logger"
)

func main() {
	neoUri := getEnv("NEO4J_URI", "bolt://localhost:7687")
	neoUser := getEnv("NEO4J_USER", "neo4j")
	neoPass := getEnv("NEO4J_PASS", "neo4jpass")

	driver, err := database.NewNeo4jDriver(neoUri, neoUser, neoPass)
	if err != nil {
		logger.Fatal(context.Background(), "failed to connect to neo4j", err)
	}
	defer driver.Close(context.Background())

	// Replay state from a specific point in time
	replayTime := time.Now().Add(-2 * time.Hour)

	logger.Info(context.Background(), fmt.Sprintf("REPLAY: Mirroring infrastructure state from %s", replayTime.Format(time.RFC3339)))

	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Fetch historical state from Neo4j (assuming historical snapshots are stored)
	query := `
		MATCH (p:Pod)
		WHERE p.updatedAt >= $replayTime
		RETURN p.name as name, p.status as status, p.node as node
	`
	result, err := session.Run(ctx, query, map[string]interface{}{"replayTime": replayTime})
	if err != nil {
		logger.Fatal(context.Background(), "replay query failed", err)
	}

	for result.Next(ctx) {
		record := result.Record()
		fmt.Printf("REPLAY_STATE: Pod %s on Node %s Status: %s\n",
			record.Values[0], record.Values[2], record.Values[1])
	}

	logger.Info(context.Background(), "Digital Twin Replay complete.")
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
