package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// LogLevel indicates the severity of the audit event.
type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARNING"
	LevelError   LogLevel = "ERROR"
	LevelFatal   LogLevel = "FATAL"
)

// AuditRecord represents a single structured log entry.
type AuditRecord struct {
	Timestamp string      `json:"timestamp"`
	Level     LogLevel    `json:"level"`
	Action    string      `json:"action"`
	Actor     string      `json:"actor"`
	Resource  string      `json:"resource"`
	Status    string      `json:"status"`
	Details   interface{} `json:"details,omitempty"`
}

// Logger defines the interface for an audit logger.
type Logger interface {
	Log(level LogLevel, action, actor, resource, status string, details interface{}) error
	Close() error
}

// FileLogger writes audit logs to a file in an append-only, thread-safe manner.
type FileLogger struct {
	file *os.File
	mu   sync.Mutex
}

// NewFileLogger creates a new FileLogger pointing to the specified file path.
func NewFileLogger(path string) (*FileLogger, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &FileLogger{file: file}, nil
}

// Log records a single audit event in JSON format.
func (l *FileLogger) Log(level LogLevel, action, actor, resource, status string, details interface{}) error {
	record := AuditRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Action:    action,
		Actor:     actor,
		Resource:  resource,
		Status:    status,
		Details:   details,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Append newline
	data = append(data, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	_, err = l.file.Write(data)
	return err
}

// Close closes the underlying log file.
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
