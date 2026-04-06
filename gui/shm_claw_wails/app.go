package main

import (
	"context"
	"fmt"

	"github.com/0xYeah/shmClaw/pkg/shm"
	"github.com/0xYeah/shmClaw/pkg/tagmatrix"
	"github.com/0xYeah/shmClaw/pkg/orchestrator"
)

// App struct
type App struct {
	ctx     context.Context
	memory  *shm.MemoryManager
	matrix  *tagmatrix.Matrix
	builder *orchestrator.ContextBuilder
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize shmClaw backend components
	mem, err := shm.NewMemoryManager("wails_shmclaw.shm", 1024*1024*10) // 10MB default
	if err != nil {
		fmt.Printf("Failed to initialize memory manager: %v\n", err)
	} else {
		a.memory = mem
	}

	a.matrix = tagmatrix.NewMatrix()
	a.builder = orchestrator.NewContextBuilder(a.matrix, a.memory, 128)
}

// shutdown is called when the app closes
func (a *App) shutdown(ctx context.Context) {
	if a.memory != nil {
		a.memory.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// StoreContext exposes the backend SHM storing capability to the frontend.
func (a *App) StoreContext(sessionID, text string) string {
	if a.memory == nil {
		return "Error: Memory manager not initialized"
	}
	
	offset, err := a.memory.Allocate(int64(len(text)) + 1)
	if err != nil {
		return fmt.Sprintf("Allocate failed: %v", err)
	}

	err = a.memory.Write(offset, []byte(text))
	if err != nil {
		return fmt.Sprintf("Write failed: %v", err)
	}

	a.matrix.AddTags(offset, []tagmatrix.Tag{
		{Key: "session", Value: sessionID},
	})

	return fmt.Sprintf("Stored context at offset %d", offset)
}

// QueryContext builds a prompt from all text associated with a given sessionID.
func (a *App) QueryContext(sessionID string) string {
	if a.builder == nil {
		return "Error: Builder not initialized"
	}

	prompt, err := a.builder.BuildPrompt([]tagmatrix.Tag{
		{Key: "session", Value: sessionID},
	})
	if err != nil {
		return fmt.Sprintf("BuildPrompt failed: %v", err)
	}

	if prompt == "" {
		return "No context found for session " + sessionID
	}
	return prompt
}
