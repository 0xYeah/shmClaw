package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProvider_Generate(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got %q", auth)
		}

		// Verify request body
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if req.Model != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %q", req.Model)
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "Hello, World!" {
			t.Errorf("Unexpected messages in request: %+v", req.Messages)
		}

		// Send mock response
		resp := chatResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			Message: struct {
				Content string `json:"content"`
			}{
				Content: "Hello from mock LLM!",
			},
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	provider := NewOpenAIProvider(mockServer.URL, "test-api-key", "gpt-4")
	result, err := provider.Generate("Hello, World!")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result != "Hello from mock LLM!" {
		t.Errorf("Expected 'Hello from mock LLM!', got %q", result)
	}
}

func TestOpenAIProvider_EmptyKey(t *testing.T) {
	provider := NewOpenAIProvider("", "", "gpt-4")
	_, err := provider.Generate("test")
	if err != ErrEmptyAPIKey {
		t.Errorf("Expected ErrEmptyAPIKey, got %v", err)
	}
}
