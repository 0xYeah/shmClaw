package server

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/0xYeah/shmClaw/pkg/mcp"
)

func TestServer_Initialize(t *testing.T) {
	srv := NewServer("TestServer", "1.0.0")

	reqPayload := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0"}}`
	respData, err := srv.HandleRequest([]byte(reqPayload))
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Expected no error, got %v", resp.Error)
	}

	// Result should be InitializeResult
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result map, got %T", resp.Result)
	}

	serverInfo := resultMap["serverInfo"].(map[string]interface{})
	if serverInfo["name"] != "TestServer" {
		t.Errorf("Expected server name 'TestServer', got %v", serverInfo["name"])
	}
}

func TestServer_CallTool(t *testing.T) {
	srv := NewServer("TestServer", "1.0.0")
	srv.RegisterTool("echo", func(args map[string]interface{}) (string, error) {
		msg, ok := args["message"].(string)
		if !ok {
			return "", errors.New("missing message")
		}
		return "Echo: " + msg, nil
	})

	reqPayload := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`
	respData, err := srv.HandleRequest([]byte(reqPayload))
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result map, got %T", resp.Result)
	}

	contentList := resultMap["content"].([]interface{})
	contentItem := contentList[0].(map[string]interface{})

	if contentItem["text"] != "Echo: hello" {
		t.Errorf("Expected 'Echo: hello', got %v", contentItem["text"])
	}
}

func TestServer_CallTool_NotFound(t *testing.T) {
	srv := NewServer("TestServer", "1.0.0")

	reqPayload := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}`
	respData, err := srv.HandleRequest([]byte(reqPayload))
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("Expected error for unknown tool")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
	}
}
