package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Auditor logs RBAC events for compliance and security
type Auditor struct {
	logFile *os.File
	mu      sync.Mutex
}

// AuditEvent represents a logged event
type AuditEvent struct {
	Timestamp time.Time         `json:"timestamp"`
	User      string            `json:"user"`
	Action    string            `json:"action"`
	Resource  string            `json:"resource"`
	Result    string            `json:"result"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewAuditor creates a new auditor
func NewAuditor(logPath string) (*Auditor, error) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	return &Auditor{
		logFile: f,
	}, nil
}

// Log logs an audit event
func (a *Auditor) Log(ctx context.Context, event AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	if _, err := a.logFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}

// Close closes the audit log
func (a *Auditor) Close() error {
	if a.logFile != nil {
		return a.logFile.Close()
	}
	return nil
}

// Query queries audit events (simplified implementation)
type AuditQuery struct {
	User      string
	Action    string
	Resource  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// QueryEvents queries audit events from the log
func (a *Auditor) QueryEvents(query AuditQuery) ([]AuditEvent, error) {
	// This is a simplified implementation
	// In production, you'd use a database or proper log aggregation system
	events := make([]AuditEvent, 0)

	// Read audit log file
	data, err := os.ReadFile(a.logFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log: %w", err)
	}

	// Parse each line as JSON
	lines := 0
	for i, b := range data {
		if b == '\n' {
			lines++
			var event AuditEvent
			if err := json.Unmarshal(data[i-len(data)+1:i], &event); err != nil {
				continue
			}

			// Apply filters
			if query.User != "" && event.User != query.User {
				continue
			}
			if query.Action != "" && event.Action != query.Action {
				continue
			}
			if query.Resource != "" && event.Resource != query.Resource {
				continue
			}
			if !query.StartTime.IsZero() && event.Timestamp.Before(query.StartTime) {
				continue
			}
			if !query.EndTime.IsZero() && event.Timestamp.After(query.EndTime) {
				continue
			}

			events = append(events, event)

			if query.Limit > 0 && len(events) >= query.Limit {
				break
			}
		}
	}

	return events, nil
}
