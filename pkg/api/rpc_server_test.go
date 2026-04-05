package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRPCServer(t *testing.T) {
	handler, err := NewRPCServer()
	if err != nil {
		t.Fatalf("Failed to create RPC server: %v", err)
	}

	// Create JSON-RPC request
	reqBody := map[string]interface{}{
		"method": "ShmClaw.Execute",
		"params": []map[string]interface{}{
			{"command": "test command"},
		},
		"id": 1,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/rpc", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	var resp struct {
		Result map[string]string `json:"result"`
		Error  interface{}       `json:"error"`
		ID     int               `json:"id"`
	}

	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got %v", resp.Error)
	}

	expectedResult := "Executed: test command"
	if resp.Result["result"] != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, resp.Result["result"])
	}
}

func TestRPCServerInvalidMethod(t *testing.T) {
	handler, _ := NewRPCServer()
	req, _ := http.NewRequest("GET", "/rpc", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %v", rr.Code)
	}
}
