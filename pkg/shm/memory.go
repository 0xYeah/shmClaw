package shm

import (
	"errors"
	"os"
	"sync"
)

var (
	ErrOutOfMemory   = errors.New("out of shared memory")
	ErrInvalidSize   = errors.New("invalid size")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrOutOfBounds   = errors.New("out of bounds")
)

type Block struct {
	Offset int64
	Size   int64
	Free   bool
}

type MemoryManager struct {
	file   *os.File
	data   []byte
	size   int64
	blocks []Block
	mu     sync.Mutex
}

// NewMemoryManager creates or opens a memory mapped file with the given size.
func NewMemoryManager(path string, size int64) (*MemoryManager, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	if fi.Size() < size {
		if err := file.Truncate(size); err != nil {
			file.Close()
			return nil, err
		}
	}

	data, err := mmapFile(file, size)
	if err != nil {
		file.Close()
		return nil, err
	}

	return &MemoryManager{
		file:   file,
		data:   data,
		size:   size,
		blocks: []Block{{Offset: 0, Size: size, Free: true}},
	}, nil
}

// Allocate allocates a block of memory of the given size.
func (m *MemoryManager) Allocate(size int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if size <= 0 {
		return 0, ErrInvalidSize
	}

	for i, b := range m.blocks {
		if b.Free && b.Size >= size {
			m.blocks[i].Free = false
			remaining := b.Size - size

			if remaining > 0 {
				m.blocks[i].Size = size
				newBlock := Block{
					Offset: b.Offset + size,
					Size:   remaining,
					Free:   true,
				}

				// Insert the new free block
				m.blocks = append(m.blocks[:i+1], append([]Block{newBlock}, m.blocks[i+1:]...)...)
			}
			return b.Offset, nil
		}
	}

	return 0, ErrOutOfMemory
}

// Free releases a previously allocated memory block.
func (m *MemoryManager) Free(offset int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, b := range m.blocks {
		if b.Offset == offset {
			if b.Free {
				return nil // Already free
			}
			m.blocks[i].Free = true
			m.mergeFreeBlocks()
			return nil
		}
	}

	return ErrInvalidOffset
}

// mergeFreeBlocks merges contiguous free blocks into a single block.
func (m *MemoryManager) mergeFreeBlocks() {
	if len(m.blocks) == 0 {
		return
	}
	merged := make([]Block, 0, len(m.blocks))
	current := m.blocks[0]

	for i := 1; i < len(m.blocks); i++ {
		b := m.blocks[i]
		if current.Free && b.Free {
			current.Size += b.Size
		} else {
			merged = append(merged, current)
			current = b
		}
	}
	merged = append(merged, current)
	m.blocks = merged
}

// Write writes data at a specific offset.
func (m *MemoryManager) Write(offset int64, b []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offset < 0 || offset+int64(len(b)) > m.size {
		return ErrOutOfBounds
	}

	copy(m.data[offset:], b)
	return nil
}

// Read reads data from a specific offset.
func (m *MemoryManager) Read(offset int64, b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offset < 0 || offset+int64(len(b)) > m.size {
		return 0, ErrOutOfBounds
	}

	n := copy(b, m.data[offset:])
	return n, nil
}

// Close unmaps the memory and closes the file.
func (m *MemoryManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error
	if m.data != nil {
		err = munmapFile(m.data)
		m.data = nil
	}
	if m.file != nil {
		if e := m.file.Close(); e != nil && err == nil {
			err = e
		}
		m.file = nil
	}
	return err
}
