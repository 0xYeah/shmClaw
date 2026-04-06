package server

import (
	"errors"

	"github.com/0xYeah/shmClaw/pkg/orchestrator"
	"github.com/0xYeah/shmClaw/pkg/shm"
	"github.com/0xYeah/shmClaw/pkg/tagmatrix"
)

// RegisterDefaultTools registers the core capabilities of shmClaw as MCP tools.
func (s *Server) RegisterDefaultTools(
	memory *shm.MemoryManager,
	matrix *tagmatrix.Matrix,
	provider orchestrator.LLMProvider,
) {

	// 1. store_context: stores text associated with a session ID
	s.RegisterTool("store_context", func(args map[string]interface{}) (string, error) {
		sessionID, ok := args["session_id"].(string)
		if !ok || sessionID == "" {
			return "", errors.New("missing or invalid 'session_id'")
		}

		text, ok := args["text"].(string)
		if !ok || text == "" {
			return "", errors.New("missing or invalid 'text'")
		}

		if memory == nil || matrix == nil {
			return "", errors.New("memory or matrix not initialized")
		}

		offset, err := memory.Allocate(int64(len(text)) + 1)
		if err != nil {
			return "", err
		}

		if err := memory.Write(offset, []byte(text)); err != nil {
			return "", err
		}

		matrix.AddTags(offset, []tagmatrix.Tag{
			{Key: "session", Value: sessionID},
		})

		return "Context stored successfully", nil
	})

	// 2. query_context: builds the context prompt for a given session ID
	s.RegisterTool("query_context", func(args map[string]interface{}) (string, error) {
		sessionID, ok := args["session_id"].(string)
		if !ok || sessionID == "" {
			return "", errors.New("missing or invalid 'session_id'")
		}

		if memory == nil || matrix == nil {
			return "", errors.New("memory or matrix not initialized")
		}

		builder := orchestrator.NewContextBuilder(matrix, memory, 128)
		prompt, err := builder.BuildPrompt([]tagmatrix.Tag{
			{Key: "session", Value: sessionID},
		})
		if err != nil {
			return "", err
		}

		return prompt, nil
	})

	// 3. generate_text: queries the LLM
	s.RegisterTool("generate_text", func(args map[string]interface{}) (string, error) {
		prompt, ok := args["prompt"].(string)
		if !ok || prompt == "" {
			return "", errors.New("missing or invalid 'prompt'")
		}

		if provider == nil {
			return "", errors.New("LLM provider not initialized")
		}

		response, err := provider.Generate(prompt)
		if err != nil {
			return "", err
		}

		return response, nil
	})
}
