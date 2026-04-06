package mcp

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCSerialization(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"1.0","clientInfo":{"name":"test-client","version":"1.0.0"}}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded JSONRPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.Method != "initialize" {
		t.Fatalf("Expected method 'initialize', got %s", decoded.Method)
	}

	var initReq InitializeRequest
	if err := json.Unmarshal(decoded.Params, &initReq); err != nil {
		t.Fatalf("Failed to unmarshal params: %v", err)
	}

	if initReq.ClientInfo.Name != "test-client" {
		t.Fatalf("Expected client name 'test-client', got %s", initReq.ClientInfo.Name)
	}
}

func TestInitializeResult(t *testing.T) {
	res := InitializeResult{
		ProtocolVersion: "1.0",
	}
	res.ServerInfo.Name = "shmClaw"
	res.ServerInfo.Version = "v0.4.0"
	res.Capabilities.Tools = true

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("Failed to marshal InitializeResult: %v", err)
	}

	var decoded InitializeResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InitializeResult: %v", err)
	}

	if !decoded.Capabilities.Tools {
		t.Fatal("Expected capabilities.tools to be true")
	}
}
