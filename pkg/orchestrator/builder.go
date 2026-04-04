package orchestrator

import (
	"bytes"
	"sort"
	"strings"

	"github.com/0xYeah/shmClaw/pkg/shm"
	"github.com/0xYeah/shmClaw/pkg/tagmatrix"
)

// LLMProvider defines the interface for interacting with Language Models.
type LLMProvider interface {
	Generate(prompt string) (string, error)
}

// ContextBuilder is responsible for assembling prompts from shared memory.
type ContextBuilder struct {
	matrix *tagmatrix.Matrix
	memory *shm.MemoryManager
	// We need to know the block size used for allocations to read them properly,
	// or we just read up to the null terminator. Assuming blocks have a max known size.
	blockSize int64
}

// NewContextBuilder creates a new ContextBuilder.
func NewContextBuilder(matrix *tagmatrix.Matrix, memory *shm.MemoryManager, blockSize int64) *ContextBuilder {
	return &ContextBuilder{
		matrix:    matrix,
		memory:    memory,
		blockSize: blockSize,
	}
}

// BuildPrompt queries the tag matrix for matching blocks, reads them from shared memory,
// and concatenates them into a single prompt string.
func (cb *ContextBuilder) BuildPrompt(queryTags []tagmatrix.Tag) (string, error) {
	offsets := cb.matrix.Query(queryTags)
	if len(offsets) == 0 {
		return "", nil
	}

	// Sort offsets to maintain a consistent reading order (e.g., chronological if allocated sequentially)
	sort.Slice(offsets, func(i, j int) bool {
		return offsets[i] < offsets[j]
	})

	var sb strings.Builder
	buf := make([]byte, cb.blockSize)

	for _, offset := range offsets {
		n, err := cb.memory.Read(offset, buf)
		if err != nil {
			return "", err
		}

		// Extract valid string by trimming null characters
		validData := buf[:n]
		nullIdx := bytes.IndexByte(validData, 0)
		if nullIdx != -1 {
			validData = validData[:nullIdx]
		}

		if len(validData) > 0 {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.Write(validData)
		}
	}

	return sb.String(), nil
}
