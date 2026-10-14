package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignDocumentHandler(t *testing.T) {
	reqBody := SignRequest{
		DocumentID: "DOC-123",
		SignerID:   "MIN-456",
		PinCode:    "1234",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/api/sign", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SignDocumentHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var doc Document
	json.NewDecoder(rr.Body).Decode(&doc)

	if doc.Status != "SIGNED" {
		t.Errorf("expected status SIGNED, got %v", doc.Status)
	}

	if doc.Signature == "" {
		t.Errorf("expected signature to be generated")
	}
}
