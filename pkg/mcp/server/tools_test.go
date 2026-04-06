package server

import (
	"strings"
	"testing"

	"github.com/0xYeah/shmClaw/pkg/shm"
	"github.com/0xYeah/shmClaw/pkg/tagmatrix"
)

// MockLLMProvider is a mock for testing generate_text
type MockLLMProvider struct{}

func (m *MockLLMProvider) Generate(prompt string) (string, error) {
	return "Mock response to: " + prompt, nil
}

func TestRegisterDefaultTools(t *testing.T) {
	srv := NewServer("TestServer", "1.0.0")
	mem, _ := shm.NewMemoryManager("test_mcp_tools.shm", 1024*1024)
	defer mem.Close()
	matrix := tagmatrix.NewMatrix()
	mockLLM := &MockLLMProvider{}

	srv.RegisterDefaultTools(mem, matrix, mockLLM)

	// 1. Test store_context
	storeHandler, ok := srv.Tools["store_context"]
	if !ok {
		t.Fatal("store_context tool not registered")
	}

	_, err := storeHandler(map[string]interface{}{
		"session_id": "test_session_1",
		"text":       "Hello from user 1",
	})
	if err != nil {
		t.Fatalf("store_context failed: %v", err)
	}

	// 2. Test query_context
	queryHandler, ok := srv.Tools["query_context"]
	if !ok {
		t.Fatal("query_context tool not registered")
	}

	res, err := queryHandler(map[string]interface{}{
		"session_id": "test_session_1",
	})
	if err != nil {
		t.Fatalf("query_context failed: %v", err)
	}

	if !strings.Contains(res, "Hello from user 1") {
		t.Fatalf("Expected context to contain 'Hello from user 1', got %q", res)
	}

	// 3. Test generate_text
	genHandler, ok := srv.Tools["generate_text"]
	if !ok {
		t.Fatal("generate_text tool not registered")
	}

	res, err = genHandler(map[string]interface{}{
		"prompt": "Say hi",
	})
	if err != nil {
		t.Fatalf("generate_text failed: %v", err)
	}

	if res != "Mock response to: Say hi" {
		t.Fatalf("Unexpected generate_text response: %s", res)
	}
}
