package tagmatrix

import (
	"sync"
)

// Tag represents a single dimension label for a memory block.
type Tag struct {
	Key   string
	Value string
}

// Matrix is an inverted index mapping tags to memory block offsets.
// It allows for fast retrieval of context blocks stored in shared memory.
type Matrix struct {
	// index maps Key -> Value -> set of offsets
	index map[string]map[string]map[int64]struct{}
	// blockTags maps offset -> set of Tags
	blockTags map[int64][]Tag
	mu        sync.RWMutex
}

// NewMatrix initializes a new Tag Matrix.
func NewMatrix() *Matrix {
	return &Matrix{
		index:     make(map[string]map[string]map[int64]struct{}),
		blockTags: make(map[int64][]Tag),
	}
}

// AddTags tags a memory block (represented by its offset) with multiple tags.
func (m *Matrix) AddTags(offset int64, tags []Tag) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tag := range tags {
		if m.index[tag.Key] == nil {
			m.index[tag.Key] = make(map[string]map[int64]struct{})
		}
		if m.index[tag.Key][tag.Value] == nil {
			m.index[tag.Key][tag.Value] = make(map[int64]struct{})
		}
		m.index[tag.Key][tag.Value][offset] = struct{}{}
	}

	// Keep track of tags per block for easy removal
	m.blockTags[offset] = append(m.blockTags[offset], tags...)
}

// RemoveBlock removes all tags associated with a specific memory block.
func (m *Matrix) RemoveBlock(offset int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tags, ok := m.blockTags[offset]
	if !ok {
		return
	}

	for _, tag := range tags {
		if m.index[tag.Key] != nil && m.index[tag.Key][tag.Value] != nil {
			delete(m.index[tag.Key][tag.Value], offset)
			if len(m.index[tag.Key][tag.Value]) == 0 {
				delete(m.index[tag.Key], tag.Value)
			}
			if len(m.index[tag.Key]) == 0 {
				delete(m.index, tag.Key)
			}
		}
	}
	delete(m.blockTags, offset)
}

// Query returns a list of block offsets that match ALL the given tags (AND semantic).
func (m *Matrix) Query(tags []Tag) []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(tags) == 0 {
		return nil
	}

	// Find the smallest set to start with (query optimization)
	var smallestSet map[int64]struct{}

	for _, tag := range tags {
		if m.index[tag.Key] == nil || m.index[tag.Key][tag.Value] == nil {
			return nil // One of the tags has no matches, so intersection is empty
		}
		set := m.index[tag.Key][tag.Value]
		if smallestSet == nil || len(set) < len(smallestSet) {
			smallestSet = set
		}
	}

	if len(smallestSet) == 0 {
		return nil
	}

	// Verify all conditions for the candidates in the smallest set
	var results []int64
	for offset := range smallestSet {
		matchesAll := true
		for _, tag := range tags {
			if _, ok := m.index[tag.Key][tag.Value][offset]; !ok {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			results = append(results, offset)
		}
	}

	return results
}
