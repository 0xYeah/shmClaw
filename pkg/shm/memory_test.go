package shm

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestMemoryManager(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.shm")
	size := int64(1024)

	// Create memory manager
	m, err := NewMemoryManager(path, size)
	if err != nil {
		t.Fatalf("Failed to create memory manager: %v", err)
	}
	defer m.Close()

	// Allocate a block
	allocSize := int64(256)
	offset, err := m.Allocate(allocSize)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}
	if offset != 0 {
		t.Fatalf("Expected offset 0, got %d", offset)
	}

	// Write data
	data := []byte("hello world")
	if err := m.Write(offset, data); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Read data
	readBuf := make([]byte, len(data))
	n, err := m.Read(offset, readBuf)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected read %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(readBuf, data) {
		t.Fatalf("Read data mismatch, got %s", readBuf)
	}

	// Allocate another block
	offset2, err := m.Allocate(allocSize)
	if err != nil {
		t.Fatalf("Failed to allocate second block: %v", err)
	}
	if offset2 != allocSize {
		t.Fatalf("Expected offset %d, got %d", allocSize, offset2)
	}

	// Free first block
	if err := m.Free(offset); err != nil {
		t.Fatalf("Failed to free: %v", err)
	}

	// Free second block
	if err := m.Free(offset2); err != nil {
		t.Fatalf("Failed to free second block: %v", err)
	}

	// Allocate a large block to test merging
	largeOffset, err := m.Allocate(allocSize * 2)
	if err != nil {
		t.Fatalf("Failed to allocate large block after free: %v", err)
	}
	if largeOffset != 0 {
		t.Fatalf("Expected offset 0, got %d", largeOffset)
	}

	// Test secure wiping (zero-out on free)
	testWipeOffset, _ := m.Allocate(10)
	testData := []byte("secret1234")
	m.Write(testWipeOffset, testData)
	m.Free(testWipeOffset)

	readAfterFree := make([]byte, 10)
	m.Read(testWipeOffset, readAfterFree)
	expectedZeros := make([]byte, 10)
	if !bytes.Equal(readAfterFree, expectedZeros) {
		t.Fatalf("Secure wiping failed: expected zeros, got %v", readAfterFree)
	}
}

func TestMemoryManagerErrors(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_errors.shm")
	size := int64(128)

	m, err := NewMemoryManager(path, size)
	if err != nil {
		t.Fatalf("Failed to create memory manager: %v", err)
	}
	defer m.Close()

	// Invalid allocate size
	if _, err := m.Allocate(0); err != ErrInvalidSize {
		t.Fatalf("Expected ErrInvalidSize, got %v", err)
	}

	// Out of memory
	if _, err := m.Allocate(size + 1); err != ErrOutOfMemory {
		t.Fatalf("Expected ErrOutOfMemory, got %v", err)
	}

	// Out of bounds write
	if err := m.Write(size, []byte("x")); err != ErrOutOfBounds {
		t.Fatalf("Expected ErrOutOfBounds, got %v", err)
	}

	// Out of bounds read
	if _, err := m.Read(size, make([]byte, 1)); err != ErrOutOfBounds {
		t.Fatalf("Expected ErrOutOfBounds, got %v", err)
	}

	// Invalid free
	if err := m.Free(size + 1); err != ErrInvalidOffset {
		t.Fatalf("Expected ErrInvalidOffset, got %v", err)
	}
}
