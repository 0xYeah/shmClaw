package orchestrator

import (
	"path/filepath"
	"testing"

	"github.com/0xYeah/shmClaw/pkg/shm"
	"github.com/0xYeah/shmClaw/pkg/tagmatrix"
)

// mockLLMProvider is a mock implementation of LLMProvider
type mockLLMProvider struct{}

func (m *mockLLMProvider) Generate(prompt string) (string, error) {
	return "Mock response for: " + prompt, nil
}

func TestContextBuilder(t *testing.T) {
	// Setup Shared Memory
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_orchestrator.shm")
	blockSize := int64(128)
	mem, err := shm.NewMemoryManager(path, blockSize*10)
	if err != nil {
		t.Fatalf("Failed to create memory manager: %v", err)
	}
	defer mem.Close()

	// Setup Tag Matrix
	matrix := tagmatrix.NewMatrix()

	// Write data to SHM and tag it
	text1 := []byte("Context part 1: Hello.")
	offset1, err := mem.Allocate(blockSize)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}
	mem.Write(offset1, text1)
	matrix.AddTags(offset1, []tagmatrix.Tag{
		{Key: "session", Value: "123"},
		{Key: "type", Value: "user"},
	})

	text2 := []byte("Context part 2: World.")
	offset2, err := mem.Allocate(blockSize)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}
	mem.Write(offset2, text2)
	matrix.AddTags(offset2, []tagmatrix.Tag{
		{Key: "session", Value: "123"},
		{Key: "type", Value: "assistant"},
	})

	text3 := []byte("Different session context.")
	offset3, err := mem.Allocate(blockSize)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}
	mem.Write(offset3, text3)
	matrix.AddTags(offset3, []tagmatrix.Tag{
		{Key: "session", Value: "456"},
	})

	// Build ContextBuilder
	builder := NewContextBuilder(matrix, mem, blockSize)

	// Test 1: Query session 123
	prompt, err := builder.BuildPrompt([]tagmatrix.Tag{
		{Key: "session", Value: "123"},
	})
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	expectedPrompt := "Context part 1: Hello.\nContext part 2: World."
	if prompt != expectedPrompt {
		t.Fatalf("Expected prompt %q, got %q", expectedPrompt, prompt)
	}

	// Test 2: Query session 123 and type user
	prompt2, err := builder.BuildPrompt([]tagmatrix.Tag{
		{Key: "session", Value: "123"},
		{Key: "type", Value: "user"},
	})
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}
	if prompt2 != "Context part 1: Hello." {
		t.Fatalf("Expected prompt 'Context part 1: Hello.', got %q", prompt2)
	}

	// Test 3: LLM Provider Mock
	llm := &mockLLMProvider{}
	res, _ := llm.Generate(prompt)
	if res != "Mock response for: "+expectedPrompt {
		t.Fatalf("LLM Provider failed to generate correctly")
	}
}
