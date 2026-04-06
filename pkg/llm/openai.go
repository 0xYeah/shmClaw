package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrEmptyAPIKey = errors.New("api key is required")
)

// OpenAIProvider implements orchestrator.LLMProvider using OpenAI's API format.
type OpenAIProvider struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI-compatible LLM provider.
// If baseURL is empty, it defaults to "https://api.openai.com/v1".
func NewOpenAIProvider(baseURL, apiKey, model string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &OpenAIProvider{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// chatRequest represents the JSON payload sent to the chat completions endpoint.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse represents the JSON payload received from the chat completions endpoint.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate sends the prompt to the LLM and returns the generated response.
func (p *OpenAIProvider) Generate(prompt string) (string, error) {
	if p.APIKey == "" {
		return "", ErrEmptyAPIKey
	}

	reqBody := chatRequest{
		Model: p.Model,
		Messages: []chatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.BaseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respData))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respData, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API returned error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("no choices returned from API")
	}

	return chatResp.Choices[0].Message.Content, nil
}
