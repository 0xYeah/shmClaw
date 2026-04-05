package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	logger, err := NewFileLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write a log
	details := map[string]string{"reason": "test"}
	err = logger.Log(LevelInfo, "READ_MEMORY", "user_123", "shm_block_01", "SUCCESS", details)
	if err != nil {
		t.Fatalf("Failed to log: %v", err)
	}

	err = logger.Close()
	if err != nil {
		t.Fatalf("Failed to close logger: %v", err)
	}

	// Verify file contents
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatal("Expected at least one line in log file")
	}

	var record AuditRecord
	if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
		t.Fatalf("Failed to parse log record: %v", err)
	}

	if record.Action != "READ_MEMORY" || record.Actor != "user_123" || record.Resource != "shm_block_01" {
		t.Fatalf("Log record mismatch: %+v", record)
	}
	if record.Level != LevelInfo {
		t.Fatalf("Expected level INFO, got %s", record.Level)
	}
	if record.Status != "SUCCESS" {
		t.Fatalf("Expected status SUCCESS, got %s", record.Status)
	}
}
