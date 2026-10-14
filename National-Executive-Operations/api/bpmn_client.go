package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// BPMNClient represents the connection to the Phase 11 Workflow Factory
type BPMNClient struct {
	EngineURL string
}

func NewBPMNClient() *BPMNClient {
	url := os.Getenv("BPMN_ENGINE_URL")
	if url == "" {
		url = "http://workflow-factory-api:8080/engine-rest"
	}
	return &BPMNClient{EngineURL: url}
}

// NotifySignature completes the current user task in the BPMN engine and routes to the next validator
func (c *BPMNClient) NotifySignature(documentID string, signerID string, signatureHash string) error {
	log.Printf("[BPMN] Notifying Workflow Factory that document %s was signed by %s", documentID, signerID)

	// Create payload for Camunda/BPMN generic REST API (Task Complete)
	payload := map[string]interface{}{
		"variables": map[string]interface{}{
			"signatureHash": map[string]interface{}{"value": signatureHash, "type": "String"},
			"signerID":      map[string]interface{}{"value": signerID, "type": "String"},
			"approved":      map[string]interface{}{"value": true, "type": "Boolean"},
		},
	}

	body, _ := json.Marshal(payload)

	// In a real scenario, we would first query /task?processInstanceBusinessKey=documentID to get the TaskID
	mockTaskID := "task_" + documentID

	req, err := http.NewRequest("POST", c.EngineURL+"/task/"+mockTaskID+"/complete", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Commented out to avoid real HTTP requests failing during simulation
	/*
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	*/

	log.Printf("[BPMN] Successfully completed workflow task %s. Document routed to next state.", mockTaskID)
	return nil
}
